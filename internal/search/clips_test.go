package search

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gifsmith/internal/provider"
)

// probeStub counts probes and answers with err (after release is closed, if set).
type probeStub struct {
	calls   atomic.Int32
	changes atomic.Int32
	mu      sync.Mutex
	err     error
	release chan struct{}
	hasKey  atomic.Bool
	blocked atomic.Bool
}

func (p *probeStub) probe(ctx context.Context) error {
	p.calls.Add(1)
	p.mu.Lock()
	rel, err := p.release, p.err
	p.mu.Unlock()
	if rel != nil {
		<-rel
	}
	return err
}

func (p *probeStub) setErr(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.err = err
}

func newClips() (*Clips, *probeStub) {
	p := &probeStub{}
	p.hasKey.Store(true)
	return &Clips{Probe: p.probe, HasKey: p.hasKey.Load, Blocked: p.blocked.Load, OnChange: func() { p.changes.Add(1) }}, p
}

// settle waits for the async probe to leave "checking".
func settle(t *testing.T, c *Clips) ClipsState {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if s := c.State(); s != ClipsChecking {
			return s
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("clips state still checking after 2 s")
	return ""
}

func TestClipsNoKey(t *testing.T) {
	c, p := newClips()
	p.hasKey.Store(false)
	c.Trigger()
	if s := c.State(); s != ClipsNoKey {
		t.Errorf("state = %s, want no_key", s)
	}
	if n := p.calls.Load(); n != 0 {
		t.Errorf("probed %d times without a key", n)
	}
	p.hasKey.Store(true)
	c.Trigger()
	settle(t, c)
	p.hasKey.Store(false)
	if s := c.State(); s != ClipsNoKey {
		t.Errorf("after the key went away: %s, want no_key regardless of history", s)
	}
}

func TestClipsTriggerOnce(t *testing.T) {
	c, p := newClips()
	if s := c.State(); s != ClipsUnknown {
		t.Errorf("before any probe: %s, want unknown", s)
	}
	p.release = make(chan struct{})
	c.Trigger()
	if s := c.State(); s != ClipsChecking {
		t.Errorf("during probe: %s, want checking", s)
	}
	close(p.release)
	if s := settle(t, c); s != ClipsEnabled {
		t.Errorf("after probe: %s, want enabled", s)
	}
	if n := p.changes.Load(); n != 2 {
		t.Errorf("OnChange called %d times, want 2 (checking, enabled)", n)
	}
	c.Trigger()
	settle(t, c)
	if n := p.calls.Load(); n != 1 {
		t.Errorf("probes = %d after two Triggers, want 1", n)
	}
}

func TestClipsOutcomes(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		want ClipsState
	}{
		"ok":           {nil, ClipsEnabled},
		"key_rejected": {&provider.Error{Kind: provider.KeyRejected, Provider: provider.Clips, Status: 403}, ClipsDisabled},
		"404":          {&provider.Error{Kind: provider.HTTP, Provider: provider.Clips, Status: 404}, ClipsDisabled},
		"rate_limited": {&provider.Error{Kind: provider.RateLimited, Provider: provider.Clips}, ClipsUnknown},
		"network":      {&provider.Error{Kind: provider.Network, Provider: provider.Clips}, ClipsUnknown},
		"timeout":      {&provider.Error{Kind: provider.Timeout, Provider: provider.Clips}, ClipsUnknown},
		"500":          {&provider.Error{Kind: provider.HTTP, Provider: provider.Clips, Status: 500}, ClipsUnknown},
		"canceled":     {context.Canceled, ClipsUnknown},
	} {
		t.Run(name, func(t *testing.T) {
			c, p := newClips()
			p.setErr(tc.err)
			c.Trigger()
			if s := settle(t, c); s != tc.want {
				t.Errorf("state = %s, want %s", s, tc.want)
			}
			if p.calls.Load() != 1 {
				t.Errorf("probes = %d, want 1", p.calls.Load())
			}
		})
	}
}

func TestClipsBlockedSkip(t *testing.T) {
	c, p := newClips()
	p.blocked.Store(true)
	c.Trigger()
	if s := c.State(); s != ClipsUnknown || p.calls.Load() != 0 {
		t.Errorf("blocked: state %s, probes %d; want unknown, 0", s, p.calls.Load())
	}
	p.blocked.Store(false)
	c.Trigger()
	if s := settle(t, c); s != ClipsEnabled || p.calls.Load() != 1 {
		t.Errorf("after the block: state %s, probes %d; want enabled, 1", s, p.calls.Load())
	}
}

func TestClipsRecheck(t *testing.T) {
	c, p := newClips()
	rl := &provider.Error{Kind: provider.RateLimited, Provider: provider.Clips}
	p.setErr(rl)
	c.Trigger()
	if s := settle(t, c); s != ClipsUnknown {
		t.Fatalf("state = %s, want unknown", s)
	}
	s, err := c.Recheck(context.Background())
	if s != ClipsUnknown || !errors.Is(err, rl) || p.calls.Load() != 2 {
		t.Errorf("recheck still failing: %s, %v, probes %d; want unknown, the probe error, 2", s, err, p.calls.Load())
	}
	p.setErr(nil)
	s, err = c.Recheck(context.Background()) // the one re-probe per key is spent: replay its error
	if s != ClipsUnknown || !errors.Is(err, rl) || p.calls.Load() != 2 {
		t.Errorf("second recheck: %s, %v, probes %d; want unknown, the last probe error, no probe", s, err, p.calls.Load())
	}
	c.Reset() // a new key gets its own re-probe
	s, err = c.Recheck(context.Background())
	if s != ClipsEnabled || err != nil || p.calls.Load() != 3 {
		t.Errorf("recheck after Reset: %s, %v, probes %d; want enabled, nil, 3", s, err, p.calls.Load())
	}
	s, err = c.Recheck(context.Background())
	if s != ClipsEnabled || err != nil || p.calls.Load() != 3 {
		t.Errorf("recheck when enabled: %s, %v, probes %d; want enabled, nil, no probe", s, err, p.calls.Load())
	}

	d, dp := newClips()
	dp.setErr(&provider.Error{Kind: provider.KeyRejected, Provider: provider.Clips, Status: 403})
	d.Trigger()
	settle(t, d)
	if s, err := d.Recheck(context.Background()); s != ClipsDisabled || err != nil || dp.calls.Load() != 1 {
		t.Errorf("recheck when disabled: %s, %v, probes %d; want disabled, nil, 1", s, err, dp.calls.Load())
	}

	b, bp := newClips()
	bp.blocked.Store(true)
	s, err = b.Recheck(context.Background())
	var pe *provider.Error
	if s != ClipsUnknown || !errors.As(err, &pe) || pe.Kind != provider.RateLimited || pe.Provider != provider.Clips || bp.calls.Load() != 0 {
		t.Errorf("recheck while blocked: %s, %v, probes %d; want unknown, clips rate_limited, 0", s, err, bp.calls.Load())
	}
}

func TestClipsReset(t *testing.T) {
	c, p := newClips()
	c.Trigger()
	settle(t, c)
	c.Reset()
	if s := c.State(); s != ClipsUnknown {
		t.Errorf("after Reset: %s, want unknown", s)
	}
	p.setErr(&provider.Error{Kind: provider.KeyRejected, Provider: provider.Clips, Status: 403})
	c.Trigger()
	if s := settle(t, c); s != ClipsDisabled || p.calls.Load() != 2 {
		t.Errorf("after Reset + Trigger: %s, probes %d; want disabled, 2", s, p.calls.Load())
	}
}

func TestClipsResetDropsStaleProbe(t *testing.T) {
	c, p := newClips()
	p.release = make(chan struct{})
	p.setErr(&provider.Error{Kind: provider.KeyRejected, Provider: provider.Clips, Status: 403})
	c.Trigger()
	c.Reset() // the key changed while the old key's probe was in flight
	close(p.release)
	time.Sleep(50 * time.Millisecond)
	if s := c.State(); s != ClipsUnknown {
		t.Errorf("stale probe result landed: %s, want unknown", s)
	}
}

func TestBlocked(t *testing.T) {
	s, f, c := setup(provider.GIPHY)
	if s.Blocked(provider.GIPHY) {
		t.Error("blocked before any 429")
	}
	f.err = &provider.Error{Kind: provider.RateLimited, Provider: provider.GIPHY, RetryAfterSec: 120}
	_, _ = s.Trending(ctx, provider.GIPHY, "")
	if !s.Blocked(provider.GIPHY) || s.Blocked(provider.Klipy) {
		t.Error("want GIPHY blocked and Klipy not")
	}
	c.advance(121 * time.Second)
	if s.Blocked(provider.GIPHY) {
		t.Error("still blocked after Retry-After")
	}
}

// A 429 on the probe blocks the shared GIPHY budget, so the next Recheck answers
// rate_limited without probing (CR-01).
func TestClipsProbe429BlocksRecheck(t *testing.T) {
	svc, _, _ := setup(provider.GIPHY)
	var probes atomic.Int32
	c := &Clips{
		Probe: func(context.Context) error {
			probes.Add(1)
			return svc.Guard(provider.Clips, func() error {
				return &provider.Error{Kind: provider.RateLimited, Provider: provider.Clips}
			})
		},
		HasKey:  func() bool { return true },
		Blocked: func() bool { return svc.Blocked(provider.GIPHY) },
	}
	c.Trigger()
	if s := settle(t, c); s != ClipsUnknown {
		t.Fatalf("state = %s, want unknown", s)
	}
	s, err := c.Recheck(context.Background())
	var pe *provider.Error
	if s != ClipsUnknown || !errors.As(err, &pe) || pe.Kind != provider.RateLimited || probes.Load() != 1 {
		t.Errorf("recheck after a probe 429: %s, %v, probes %d; want unknown, rate_limited, 1", s, err, probes.Load())
	}
}

// WR-02: a Recheck before the launch's first GIFs success (offline at login, ⌘2 early)
// spends neither the re-probe nor the GIFs-success probe, and its 401 is not "not approved".
func TestClipsRecheckBeforeGifsSuccess(t *testing.T) {
	c, p := newClips()
	timeout := &provider.Error{Kind: provider.Timeout, Provider: provider.Clips}
	p.setErr(timeout)
	for i := int32(1); i <= 2; i++ {
		if s, err := c.Recheck(context.Background()); s != ClipsUnknown || !errors.Is(err, timeout) || p.calls.Load() != i {
			t.Fatalf("pre-success recheck %d: %s, %v, probes %d; want unknown, the timeout, %d", i, s, err, p.calls.Load(), i)
		}
	}
	p.setErr(nil)
	c.Trigger() // GIPHY GIFs succeeded once the network came back
	if s := settle(t, c); s != ClipsEnabled || p.calls.Load() != 3 {
		t.Errorf("Trigger after pre-success rechecks: %s, probes %d; want enabled, 3", s, p.calls.Load())
	}

	d, dp := newClips()
	rejected := &provider.Error{Kind: provider.KeyRejected, Provider: provider.Clips, Status: 401}
	dp.setErr(rejected)
	if s, err := d.Recheck(context.Background()); s != ClipsUnknown || !errors.Is(err, rejected) {
		t.Errorf("pre-success 401: %s, %v; want unknown and the key_rejected error (a bad key gets 401 too)", s, err)
	}
	d.Trigger()
	if s := settle(t, d); s != ClipsDisabled || dp.calls.Load() != 2 {
		t.Errorf("401 after a GIFs success: %s, probes %d; want disabled, 2", s, dp.calls.Load())
	}

	// A GIFs success while a pre-success Recheck is in flight: no second probe, and the
	// in-flight 401 now counts as "not approved".
	e, ep := newClips()
	ep.setErr(rejected)
	ep.release = make(chan struct{})
	done := make(chan ClipsState)
	go func() {
		s, _ := e.Recheck(context.Background())
		done <- s
	}()
	for e.State() != ClipsChecking {
		time.Sleep(time.Millisecond)
	}
	e.Trigger()
	close(ep.release)
	if s := <-done; s != ClipsDisabled || ep.calls.Load() != 1 {
		t.Errorf("Trigger during a pre-success recheck: %s, probes %d; want disabled, 1", s, ep.calls.Load())
	}
}
