package search

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"gifsmith/internal/provider"
)

// fakeProv counts calls and answers with a fixed page or error.
type fakeProv struct {
	mu    sync.Mutex
	calls int
	page  provider.Page
	err   error
}

func (f *fakeProv) answer() (provider.Page, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.page, f.err
}

func (f *fakeProv) Search(context.Context, string, string) (provider.Page, error) { return f.answer() }
func (f *fakeProv) Trending(context.Context, string) (provider.Page, error)       { return f.answer() }

func (f *fakeProv) n() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func page(ids ...string) provider.Page {
	pg := provider.Page{Items: []provider.Gif{}, Next: "n"}
	for _, id := range ids {
		pg.Items = append(pg.Items, provider.Gif{ID: id})
	}
	return pg
}

func ids(pg provider.Page) []string {
	out := []string{}
	for _, g := range pg.Items {
		out = append(out, g.ID)
	}
	return out
}

func setup(p provider.ID) (*Service, *fakeProv, *clock) {
	f := &fakeProv{page: page("a", "b")}
	c := &clock{t: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)}
	return New(map[provider.ID]provider.Provider{p: f}, c.now), f, c
}

var ctx = context.Background()

func TestTrendingTTL(t *testing.T) {
	s, f, c := setup(provider.GIPHY)
	for range 2 {
		if _, err := s.Trending(ctx, provider.GIPHY, ""); err != nil {
			t.Fatal(err)
		}
	}
	if f.n() != 1 {
		t.Fatalf("calls inside TTL = %d, want 1", f.n())
	}
	c.advance(TrendingTTL + time.Second)
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	if f.n() != 2 {
		t.Fatalf("calls after TTL = %d, want 2", f.n())
	}
}

func TestSearchTTL(t *testing.T) {
	s, f, c := setup(provider.GIPHY)
	_, _ = s.Search(ctx, provider.GIPHY, "dwight", "")
	c.advance(4 * time.Minute)
	_, _ = s.Search(ctx, provider.GIPHY, " dwight ", "") // trimmed query shares the entry
	if f.n() != 1 {
		t.Fatalf("calls inside TTL = %d, want 1", f.n())
	}
	c.advance(time.Minute + time.Second)
	_, _ = s.Search(ctx, provider.GIPHY, "dwight", "")
	if f.n() != 2 {
		t.Fatalf("calls after TTL = %d, want 2", f.n())
	}
}

func TestSeparateKeys(t *testing.T) {
	g := &fakeProv{page: page("a")}
	k := &fakeProv{page: page("a")}
	s := New(map[provider.ID]provider.Provider{provider.GIPHY: g, provider.Klipy: k}, nil)
	_, _ = s.Search(ctx, provider.GIPHY, "dwight", "")
	_, _ = s.Search(ctx, provider.GIPHY, "jim", "")
	_, _ = s.Search(ctx, provider.GIPHY, "dwight", "24")
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	_, _ = s.Search(ctx, provider.Klipy, "dwight", "")
	if g.n() != 4 || k.n() != 1 {
		t.Fatalf("giphy calls %d (want 4), klipy calls %d (want 1)", g.n(), k.n())
	}
}

func rateLimited(t *testing.T, err error) *provider.Error {
	t.Helper()
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != provider.RateLimited {
		t.Fatalf("err = %v, want rate_limited", err)
	}
	return pe
}

func TestBlockKnownRetry(t *testing.T) {
	s, f, c := setup(provider.GIPHY)
	f.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY, RetryAfterSec: 120}
	_, err := s.Search(ctx, provider.GIPHY, "dwight", "")
	if pe := rateLimited(t, err); pe.RetryAfterSec != 120 {
		t.Fatalf("first RetryAfterSec = %d, want 120", pe.RetryAfterSec)
	}
	f.err = nil
	c.advance(30 * time.Second)
	_, err = s.Trending(ctx, provider.GIPHY, "") // any request to the provider is blocked
	if pe := rateLimited(t, err); pe.RetryAfterSec != 90 || pe.Provider != provider.GIPHY {
		t.Fatalf("blocked err = %+v, want giphy RetryAfterSec 90", pe)
	}
	if f.n() != 1 {
		t.Fatalf("calls while blocked = %d, want 1", f.n())
	}
	c.advance(91 * time.Second) // +121 s
	if _, err := s.Search(ctx, provider.GIPHY, "dwight", ""); err != nil {
		t.Fatalf("after block: %v", err)
	}
	if f.n() != 2 {
		t.Fatalf("calls after block = %d, want 2", f.n())
	}
}

func TestBlockUnknownRetry(t *testing.T) {
	s, f, c := setup(provider.GIPHY)
	f.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY}
	_, _ = s.Search(ctx, provider.GIPHY, "dwight", "")
	f.err = nil
	c.advance(30 * time.Second)
	_, err := s.Search(ctx, provider.GIPHY, "dwight", "")
	if pe := rateLimited(t, err); pe.RetryAfterSec != 0 {
		t.Fatalf("unknown retry became %d", pe.RetryAfterSec)
	}
	if f.n() != 1 {
		t.Fatalf("calls while blocked = %d, want 1", f.n())
	}
	c.advance(31 * time.Second) // +61 s
	if _, err := s.Search(ctx, provider.GIPHY, "dwight", ""); err != nil {
		t.Fatalf("after MinBlock: %v", err)
	}
	if f.n() != 2 {
		t.Fatalf("calls after MinBlock = %d, want 2", f.n())
	}
}

func TestShortRetryAfterBlocksMinBlock(t *testing.T) {
	s, f, c := setup(provider.GIPHY)
	f.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY, RetryAfterSec: 5}
	_, _ = s.Search(ctx, provider.GIPHY, "dwight", "")
	c.advance(30 * time.Second)
	_, err := s.Search(ctx, provider.GIPHY, "dwight", "")
	if pe := rateLimited(t, err); pe.RetryAfterSec != 30 {
		t.Fatalf("RetryAfterSec = %d, want 30 (MinBlock floor)", pe.RetryAfterSec)
	}
	if f.n() != 1 {
		t.Fatalf("calls = %d, want 1", f.n())
	}
}

func TestNoCacheOnError(t *testing.T) {
	for _, kind := range []provider.Kind{provider.Network, provider.Timeout, provider.HTTP} {
		s, f, _ := setup(provider.FinerGifs)
		f.err = &provider.Error{Kind: kind, Provider: provider.FinerGifs, Status: 500}
		_, _ = s.Search(ctx, provider.FinerGifs, "dwight", "")
		f.err = nil
		if _, err := s.Search(ctx, provider.FinerGifs, "dwight", ""); err != nil {
			t.Fatalf("%s: second call blocked or cached the error: %v", kind, err)
		}
		if f.n() != 2 {
			t.Fatalf("%s: calls = %d, want 2", kind, f.n())
		}
	}
}

func TestCancelNotCachedNotBlocked(t *testing.T) {
	s, f, _ := setup(provider.GIPHY)
	f.err = context.Canceled
	_, err := s.Search(ctx, provider.GIPHY, "dwight", "")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled as is", err)
	}
	f.err = nil
	if _, err := s.Search(ctx, provider.GIPHY, "dwight", ""); err != nil {
		t.Fatalf("after cancel: %v", err)
	}
	if f.n() != 2 {
		t.Fatalf("calls = %d, want 2", f.n())
	}
}

func TestDedupeNonKlipy(t *testing.T) {
	s, f, _ := setup(provider.FinerGifs)
	f.page = page("a", "b", "a", "c")
	pg, _ := s.Search(ctx, provider.FinerGifs, "dwight", "")
	if got := ids(pg); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("finergifs ids = %v, want [a b c]", got)
	}
	if pg.Next != "n" {
		t.Fatalf("Next = %q, lost by dedupe", pg.Next)
	}
	if got := ids(f.page); !slices.Equal(got, []string{"a", "b", "a", "c"}) {
		t.Fatalf("dedupe mutated the provider's slice: %v", got)
	}
}

func TestNoDedupeKlipy(t *testing.T) {
	s, f, _ := setup(provider.Klipy)
	f.page = page("a", "b", "a", "c")
	pg, _ := s.Trending(ctx, provider.Klipy, "")
	if got := ids(pg); !slices.Equal(got, []string{"a", "b", "a", "c"}) {
		t.Fatalf("klipy ids = %v, want unchanged [a b a c]", got)
	}
}

func TestSeedInvalidate(t *testing.T) {
	s, f, _ := setup(provider.GIPHY)
	s.Seed(provider.GIPHY, "", page("seeded"))
	pg, err := s.Trending(ctx, provider.GIPHY, "")
	if err != nil || !slices.Equal(ids(pg), []string{"seeded"}) || f.n() != 0 {
		t.Fatalf("seeded trending = %v, %v, calls %d; want [seeded], nil, 0", ids(pg), err, f.n())
	}
	f.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY, RetryAfterSec: 600}
	_, _ = s.Search(ctx, provider.GIPHY, "dwight", "")
	f.err = nil
	s.Invalidate(provider.GIPHY)
	pg, err = s.Trending(ctx, provider.GIPHY, "")
	if err != nil || !slices.Equal(ids(pg), []string{"a", "b"}) || f.n() != 2 {
		t.Fatalf("after Invalidate = %v, %v, calls %d; want [a b], nil, 2", ids(pg), err, f.n())
	}
}

func TestUnknownProvider(t *testing.T) {
	s, _, _ := setup(provider.GIPHY)
	if _, err := s.Search(ctx, provider.Klipy, "dwight", ""); err == nil {
		t.Fatal("want error for unconfigured provider")
	}
	if _, err := s.Trending(ctx, provider.ID("nope"), ""); err == nil {
		t.Fatal("want error for unknown provider")
	}
}

func TestConcurrent(t *testing.T) {
	f := &fakeProv{page: page("a", "a", "b")}
	s := New(map[provider.ID]provider.Provider{provider.GIPHY: f}, nil)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			q := []string{"x", "y", "z"}[i%3]
			_, _ = s.Search(ctx, provider.GIPHY, q, "")
			_, _ = s.Trending(ctx, provider.GIPHY, "")
			if i%7 == 0 {
				s.Invalidate(provider.GIPHY)
			}
		})
	}
	wg.Wait()
}

// GIPHY GIFs and Clips share one key, so one 429 budget and one Invalidate (CR-01).
func TestClipsShareGiphyBudget(t *testing.T) {
	g, cl := &fakeProv{page: page("a")}, &fakeProv{page: page("c")}
	c := &clock{t: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)}
	s := New(map[provider.ID]provider.Provider{provider.GIPHY: g, provider.Clips: cl}, c.now)

	g.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY, RetryAfterSec: 120}
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	_, err := s.Search(ctx, provider.Clips, "dwight", "")
	if pe := rateLimited(t, err); pe.Provider != provider.Clips || pe.RetryAfterSec != 120 || cl.n() != 0 {
		t.Fatalf("clips after a GIPHY 429: %+v, calls %d; want clips rate_limited 120, 0 calls", pe, cl.n())
	}
	if !s.Blocked(provider.Clips) {
		t.Error("Blocked(clips) = false inside the GIPHY block")
	}
	g.err = nil
	c.advance(121 * time.Second)

	cl.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.Clips}
	_, _ = s.Trending(ctx, provider.Clips, "")
	if _, err := s.Trending(ctx, provider.GIPHY, ""); g.n() != 1 {
		t.Fatalf("GIPHY after a Clips 429: %v, calls %d; want blocked, 1 call", err, g.n())
	}

	err = s.Guard(provider.Clips, func() error { t.Fatal("Guard called into a block"); return nil })
	_ = rateLimited(t, err)

	cl.err = nil
	s.Invalidate(provider.GIPHY)
	if _, err := s.Trending(ctx, provider.Clips, ""); err != nil {
		t.Fatalf("clips after Invalidate(giphy): %v", err)
	}
	s.Invalidate(provider.GIPHY)
	_, _ = s.Trending(ctx, provider.Clips, "")
	if cl.n() != 3 {
		t.Errorf("clips calls = %d, want 3 (Invalidate(giphy) drops cached clips pages)", cl.n())
	}

	s2, _, _ := setup(provider.GIPHY)
	_ = s2.Guard(provider.Clips, func() error { return &provider.Error{Kind: provider.RateLimited, Provider: provider.Clips} })
	if !s2.Blocked(provider.GIPHY) {
		t.Error("a Guard 429 did not block the GIPHY budget")
	}
}

// gatedProv blocks each call until release, then answers with page or err.
type gatedProv struct {
	started chan struct{}
	release chan struct{}
	page    provider.Page
	err     error
}

func (g *gatedProv) answer() (provider.Page, error) {
	g.started <- struct{}{}
	<-g.release
	return g.page, g.err
}

func (g *gatedProv) Search(context.Context, string, string) (provider.Page, error) { return g.answer() }
func (g *gatedProv) Trending(context.Context, string) (provider.Page, error)       { return g.answer() }

// A request made with the old key that lands after Invalidate must not restore
// its 429 block or overwrite the new key's seeded page (WR-01).
func TestInvalidateDropsInFlightResult(t *testing.T) {
	for name, tc := range map[string]struct {
		page provider.Page
		err  error
	}{
		"429":     {err: &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY, RetryAfterSec: 600}},
		"success": {page: page("old")},
	} {
		t.Run(name, func(t *testing.T) {
			g := &gatedProv{started: make(chan struct{}), release: make(chan struct{}), page: tc.page, err: tc.err}
			s := New(map[provider.ID]provider.Provider{provider.GIPHY: g}, nil)
			done := make(chan struct{})
			go func() {
				defer close(done)
				_, _ = s.Trending(ctx, provider.GIPHY, "")
			}()
			<-g.started
			s.Invalidate(provider.GIPHY) // SaveKey: new key stored, then Invalidate + Seed
			s.Seed(provider.GIPHY, "", page("new"))
			close(g.release)
			<-done
			if s.Blocked(provider.GIPHY) {
				t.Error("the old key's 429 blocked the new key")
			}
			pg, err := s.Trending(ctx, provider.GIPHY, "")
			if err != nil || !slices.Equal(ids(pg), []string{"new"}) {
				t.Errorf("trending after the late answer = %v, %v; want the seeded [new]", ids(pg), err)
			}
		})
	}
}

// Flush (a provider parameter changed) drops cached pages of the budget but keeps a 429 block.
func TestFlushKeepsRateLimitBlock(t *testing.T) {
	g, cl := &fakeProv{page: page("a")}, &fakeProv{page: page("c")}
	c := &clock{t: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)}
	s := New(map[provider.ID]provider.Provider{provider.GIPHY: g, provider.Clips: cl}, c.now)
	_, _ = s.Trending(ctx, provider.Clips, "")
	g.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY, RetryAfterSec: 120}
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	g.err = nil
	s.Flush(provider.GIPHY)
	_, err := s.Search(ctx, provider.GIPHY, "cat", "")
	if pe := rateLimited(t, err); pe.RetryAfterSec != 120 || g.n() != 1 {
		t.Fatalf("after Flush: %+v, calls %d; want still rate_limited, no new call", pe, g.n())
	}
	c.advance(121 * time.Second)
	_, _ = s.Trending(ctx, provider.Clips, "")
	if cl.n() != 2 {
		t.Errorf("clips calls = %d, want 2 (Flush(giphy) drops the cached Clips page)", cl.n())
	}
	s.Invalidate(provider.GIPHY) // a key change still lifts a block
	g.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY}
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	g.err = nil
	s.Invalidate(provider.GIPHY)
	if _, err := s.Trending(ctx, provider.GIPHY, ""); err != nil {
		t.Errorf("Invalidate did not lift the block: %v", err)
	}
}

func TestFlushKeepsOtherProviders(t *testing.T) {
	g, k := &fakeProv{page: page("a")}, &fakeProv{page: page("k")}
	s := New(map[provider.ID]provider.Provider{provider.GIPHY: g, provider.Klipy: k}, nil)
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	_, _ = s.Trending(ctx, provider.Klipy, "")
	s.Flush(provider.GIPHY)
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	_, _ = s.Trending(ctx, provider.Klipy, "")
	if g.n() != 2 || k.n() != 1 {
		t.Errorf("calls giphy %d klipy %d; want 2 and 1", g.n(), k.n())
	}
}
