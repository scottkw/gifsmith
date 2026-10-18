// Package search sits in front of every provider: TTL cache, 429 budget and
// the dedupe policy. Domain code only: no Wails imports, nothing on disk.
package search

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"time"

	"gifsmith/internal/provider"
)

const (
	TrendingTTL = 10 * time.Minute
	SearchTTL   = 5 * time.Minute
	MinBlock    = 60 * time.Second
)

var ErrUnknownProvider = errors.New("unknown provider")

type entry struct {
	page    provider.Page
	expires time.Time
}

type block struct {
	until time.Time
	known bool // the vendor sent a Retry-After
}

// Service caches API metadata in memory only and never retries into a 429.
type Service struct {
	providers map[provider.ID]provider.Provider
	now       func() time.Time

	mu      sync.Mutex // never held across a provider call
	cache   map[string]entry
	blocked map[provider.ID]block
	gen     map[provider.ID]int // per budget, bumped by Invalidate: an old key's late answer is never recorded
}

func New(providers map[provider.ID]provider.Provider, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{providers: providers, now: now, cache: map[string]entry{}, blocked: map[provider.ID]block{}, gen: map[provider.ID]int{}}
}

func (s *Service) Search(ctx context.Context, p provider.ID, q, cursor string) (provider.Page, error) {
	q = strings.TrimSpace(q)
	return s.get(p, "search|"+q+"|"+cursor, SearchTTL, func(pr provider.Provider) (provider.Page, error) {
		return pr.Search(ctx, q, cursor)
	})
}

func (s *Service) Trending(ctx context.Context, p provider.ID, cursor string) (provider.Page, error) {
	return s.get(p, "trending||"+cursor, TrendingTTL, func(pr provider.Provider) (provider.Page, error) {
		return pr.Trending(ctx, cursor)
	})
}

// Seed stores a trending page fetched elsewhere (key verification doubles as the first trending fetch).
func (s *Service) Seed(p provider.ID, cursor string, pg provider.Page) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[string(p)+"|trending||"+cursor] = entry{dedupe(p, pg), s.now().Add(TrendingTTL)}
}

// budget is the key and 429 budget p draws from: GIPHY Clips use the GIPHY key.
func budget(p provider.ID) provider.ID {
	if p == provider.Clips {
		return provider.GIPHY
	}
	return p
}

// Invalidate drops the cached pages and the 429 block of every provider on p's key (the key changed).
func (s *Service) Invalidate(p provider.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.cache {
		if id, _, _ := strings.Cut(k, "|"); budget(provider.ID(id)) == budget(p) {
			delete(s.cache, k)
		}
	}
	delete(s.blocked, budget(p))
	s.gen[budget(p)]++
}

// Flush drops the cached pages of every provider on p's budget because a provider parameter
// changed (rating, page size, language, country). The 429 block stays: the key did not change.
func (s *Service) Flush(p provider.ID) {}

// Guard runs an uncached call (the Clips probe) inside p's 429 budget: refused
// while blocked, and a 429 it returns blocks the budget like any other request.
func (s *Service) Guard(p provider.ID, call func() error) error {
	s.mu.Lock()
	err, gen := s.blockedErr(p), s.gen[budget(p)]
	s.mu.Unlock()
	if err != nil {
		return err
	}
	err = call()
	s.noteRateLimit(p, gen, err)
	return err
}

// blockedErr is the rate_limited error while p's budget is blocked, else nil. Caller holds mu.
func (s *Service) blockedErr(p provider.ID) error {
	now := s.now()
	b, ok := s.blocked[budget(p)]
	if !ok || !now.Before(b.until) {
		return nil
	}
	retry := 0
	if b.known {
		retry = int(math.Ceil(b.until.Sub(now).Seconds()))
	}
	return &provider.Error{Kind: provider.RateLimited, Provider: p, RetryAfterSec: retry}
}

// noteRateLimit blocks p's budget when err is a 429 for the key generation gen.
func (s *Service) noteRateLimit(p provider.ID, gen int, err error) {
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != provider.RateLimited {
		return
	}
	wait := max(time.Duration(pe.RetryAfterSec)*time.Second, MinBlock)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gen[budget(p)] == gen {
		s.blocked[budget(p)] = block{until: s.now().Add(wait), known: pe.RetryAfterSec > 0}
	}
}

func (s *Service) get(p provider.ID, rest string, ttl time.Duration, call func(provider.Provider) (provider.Page, error)) (provider.Page, error) {
	pr := s.providers[p]
	if pr == nil {
		return provider.Page{}, ErrUnknownProvider
	}
	key := string(p) + "|" + rest

	s.mu.Lock()
	if err := s.blockedErr(p); err != nil {
		s.mu.Unlock()
		return provider.Page{}, err
	}
	if e, ok := s.cache[key]; ok && s.now().Before(e.expires) {
		s.mu.Unlock()
		return e.page, nil
	}
	gen := s.gen[budget(p)]
	s.mu.Unlock()

	pg, err := call(pr)
	if err != nil {
		s.noteRateLimit(p, gen, err)
		return pg, err // errors and cancellations are never cached
	}
	pg = dedupe(p, pg)
	s.mu.Lock()
	if s.gen[budget(p)] == gen { // the key changed mid-flight: answer the caller, never cache the old key's page
		s.cache[key] = entry{pg, s.now().Add(ttl)}
	}
	s.mu.Unlock()
	return pg, nil
}

// dedupe drops repeated IDs within one page, keeping the first; Klipy pages
// stay exactly as received (its integration rules forbid altering results).
func dedupe(p provider.ID, pg provider.Page) provider.Page {
	if p == provider.Klipy {
		return pg
	}
	seen := make(map[string]bool, len(pg.Items))
	items := make([]provider.Gif, 0, len(pg.Items))
	for _, g := range pg.Items {
		if !seen[g.ID] {
			seen[g.ID] = true
			items = append(items, g)
		}
	}
	pg.Items = items
	return pg
}

// Blocked reports whether p's budget is inside a 429 block.
func (s *Service) Blocked(p provider.ID) bool {
	return s.BlockedErr(p) != nil
}

// BlockedErr is p's rate_limited error (with the seconds left, when known) while its budget is blocked, else nil.
func (s *Service) BlockedErr(p provider.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.blockedErr(p)
}
