package search

import (
	"context"
	"errors"
	"sync"
	"time"

	"gifsmith/internal/provider"
)

type ClipsState string

const (
	ClipsNoKey    ClipsState = "no_key"
	ClipsChecking ClipsState = "checking"
	ClipsEnabled  ClipsState = "enabled"
	ClipsDisabled ClipsState = "disabled"
	ClipsUnknown  ClipsState = "unknown"
)

const probeTimeout = 15 * time.Second

// Clips remembers, for this process only, whether the user's GIPHY key is
// approved for Clips. It spends at most one probe per launch (per key), and
// only when the caller already saw a GIFs success with that key.
type Clips struct {
	Probe    func(ctx context.Context) error
	HasKey   func() bool
	Blocked  func() bool // GIPHY is inside a 429 block: never probe into it
	OnChange func()      // optional; called after every state change

	mu     sync.Mutex
	state  ClipsState // "" until a probe starts: reported as unknown
	probed bool       // a probe ran or is running since the last Reset
	gen    int        // bumped by Reset so a probe for an old key cannot land
}

func (c *Clips) State() ClipsState {
	if !c.HasKey() {
		return ClipsNoKey
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == "" {
		return ClipsUnknown
	}
	return c.state
}

// Trigger starts the launch's single probe in the background; later calls are no-ops.
func (c *Clips) Trigger() {
	gen, ok := c.start(false)
	if !ok {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
		defer cancel()
		c.finish(gen, c.Probe(ctx))
	}()
}

// Recheck probes again only when the result is unknown (the user chose Clips).
// It returns the new state, plus the probe error when the state is still unknown.
func (c *Clips) Recheck(ctx context.Context) (ClipsState, error) {
	if c.HasKey() && c.Blocked() && c.State() == ClipsUnknown {
		return ClipsUnknown, &provider.Error{Kind: provider.RateLimited, Provider: provider.Clips}
	}
	gen, ok := c.start(true)
	if !ok {
		return c.State(), nil
	}
	err := c.Probe(ctx)
	s := c.finish(gen, err)
	if s != ClipsUnknown {
		return s, nil
	}
	return s, err
}

// Reset forgets the result (the GIPHY key changed). It does not notify:
// SaveKey emits its own change event right after.
func (c *Clips) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state, c.probed = "", false
	c.gen++
}

// start moves to checking if a probe may run: a key, no 429 block, and either
// no probe yet (Trigger) or an unknown result (Recheck).
func (c *Clips) start(recheck bool) (int, bool) {
	if !c.HasKey() || c.Blocked() {
		return 0, false
	}
	c.mu.Lock()
	unknown := c.state == "" || c.state == ClipsUnknown
	if (recheck && !unknown) || (!recheck && c.probed) {
		c.mu.Unlock()
		return 0, false
	}
	c.probed = true
	c.state = ClipsChecking
	gen := c.gen
	c.mu.Unlock()
	c.changed()
	return gen, true
}

func (c *Clips) finish(gen int, err error) ClipsState {
	s := classify(err)
	c.mu.Lock()
	if gen != c.gen {
		c.mu.Unlock()
		return ClipsUnknown
	}
	c.state = s
	c.mu.Unlock()
	c.changed()
	return s
}

// classify: 401/403 (key_rejected) or 404 after a GIFs success means "not approved";
// rate limits, network, timeouts and anything else leave the answer unknown.
func classify(err error) ClipsState {
	if err == nil {
		return ClipsEnabled
	}
	var pe *provider.Error
	if errors.As(err, &pe) && (pe.Kind == provider.KeyRejected || (pe.Kind == provider.HTTP && pe.Status == 404)) {
		return ClipsDisabled
	}
	return ClipsUnknown
}

func (c *Clips) changed() {
	if c.OnChange != nil {
		c.OnChange()
	}
}
