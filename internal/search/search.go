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
}

func New(providers map[provider.ID]provider.Provider, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{providers: providers, now: now, cache: map[string]entry{}, blocked: map[provider.ID]block{}}
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

// Invalidate drops the provider's cached pages and its 429 block (the key changed).
func (s *Service) Invalidate(p provider.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := string(p) + "|"
	for k := range s.cache {
		if strings.HasPrefix(k, prefix) {
			delete(s.cache, k)
		}
	}
	delete(s.blocked, p)
}

func (s *Service) get(p provider.ID, rest string, ttl time.Duration, call func(provider.Provider) (provider.Page, error)) (provider.Page, error) {
	pr := s.providers[p]
	if pr == nil {
		return provider.Page{}, ErrUnknownProvider
	}
	key := string(p) + "|" + rest

	s.mu.Lock()
	now := s.now()
	if b, ok := s.blocked[p]; ok && now.Before(b.until) {
		s.mu.Unlock()
		retry := 0
		if b.known {
			retry = int(math.Ceil(b.until.Sub(now).Seconds()))
		}
		return provider.Page{}, &provider.Error{Kind: provider.RateLimited, Provider: p, RetryAfterSec: retry}
	}
	if e, ok := s.cache[key]; ok && now.Before(e.expires) {
		s.mu.Unlock()
		return e.page, nil
	}
	s.mu.Unlock()

	pg, err := call(pr)
	if err != nil {
		var pe *provider.Error
		if errors.As(err, &pe) && pe.Kind == provider.RateLimited {
			wait := max(time.Duration(pe.RetryAfterSec)*time.Second, MinBlock)
			s.mu.Lock()
			s.blocked[p] = block{until: s.now().Add(wait), known: pe.RetryAfterSec > 0}
			s.mu.Unlock()
		}
		return pg, err // errors and cancellations are never cached
	}
	pg = dedupe(p, pg)
	s.mu.Lock()
	s.cache[key] = entry{pg, s.now().Add(ttl)}
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

// Blocked reports whether p is inside a 429 block (skeleton).
func (s *Service) Blocked(p provider.ID) bool { return false }
