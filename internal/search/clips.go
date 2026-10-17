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
	Blocked  func() error // GIPHY's rate_limited error while its 429 block lasts, else nil: never probe into it
	OnChange func()       // optional; called after every state change

	mu        sync.Mutex
	state     ClipsState // "" until a probe starts: reported as unknown
	triggered bool       // Trigger's probe ran or is running since the last Reset
	proven    bool       // a GIFs call succeeded with this key, so a 401 from the probe means "not approved"
	rechecked bool       // Recheck spent its one re-probe since the last Reset (only a probe after proven, and not a 429)
	lastErr   error      // the last probe's error, replayed once the re-probe is spent
	gen       int        // bumped by Reset so a probe for an old key cannot land
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

// Trigger runs after a GIFs success with the key. It starts the launch's single
// probe in the background; later calls (or calls while a probe runs) only note the success.
func (c *Clips) Trigger() {
	c.mu.Lock()
	c.proven = true
	c.mu.Unlock()
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

// Recheck probes again only when the result is unknown (the user chose Clips),
// at most once per key per launch. It returns the new state, plus the probe
// error (or, once the re-probe is spent, the last one) when still unknown.
func (c *Clips) Recheck(ctx context.Context) (ClipsState, error) {
	if c.HasKey() && c.State() == ClipsUnknown {
		if err := c.Blocked(); err != nil {
			return ClipsUnknown, err
		}
	}
	gen, ok := c.start(true)
	if !ok {
		c.mu.Lock()
		err := c.lastErr
		c.mu.Unlock()
		s := c.State()
		if s != ClipsUnknown {
			err = nil
		}
		return s, err
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
	c.state, c.triggered, c.proven, c.rechecked, c.lastErr = "", false, false, false, nil
	c.gen++
}

// start moves to checking if a probe may run: a key, no 429 block, and either
// no Trigger probe yet and none running (Trigger) or an unknown result and no
// re-probe yet (Recheck). A Recheck before any GIFs success never spends either.
func (c *Clips) start(recheck bool) (int, bool) {
	if !c.HasKey() || c.Blocked() != nil {
		return 0, false
	}
	c.mu.Lock()
	unknown := c.state == "" || c.state == ClipsUnknown
	if (recheck && (!unknown || c.rechecked)) || (!recheck && (c.triggered || c.state == ClipsChecking)) {
		c.mu.Unlock()
		return 0, false
	}
	c.triggered = c.triggered || !recheck
	c.rechecked = c.rechecked || (recheck && c.proven)
	c.state = ClipsChecking
	gen := c.gen
	c.mu.Unlock()
	c.changed()
	return gen, true
}

func (c *Clips) finish(gen int, err error) ClipsState {
	c.mu.Lock()
	if gen != c.gen {
		c.mu.Unlock()
		return ClipsUnknown
	}
	s := classify(err, c.proven)
	c.state, c.lastErr = s, err
	var pe *provider.Error
	if errors.As(err, &pe) && pe.Kind == provider.RateLimited {
		c.rechecked = false // a 429 says nothing about Clips, and the block already spaces the next probe
	}
	c.mu.Unlock()
	c.changed()
	return s
}

// classify: 401/403 (key_rejected) or 404 after a GIFs success means "not approved"
// (before one, a bad key gets the same answer); rate limits, network, timeouts and
// anything else leave the answer unknown.
func classify(err error, proven bool) ClipsState {
	if err == nil {
		return ClipsEnabled
	}
	var pe *provider.Error
	if proven && errors.As(err, &pe) && (pe.Kind == provider.KeyRejected || (pe.Kind == provider.HTTP && pe.Status == 404)) {
		return ClipsDisabled
	}
	return ClipsUnknown
}

func (c *Clips) changed() {
	if c.OnChange != nil {
		c.OnChange()
	}
}
