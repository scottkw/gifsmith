package search

import "context"

type ClipsState string

const (
	ClipsNoKey    ClipsState = "no_key"
	ClipsChecking ClipsState = "checking"
	ClipsEnabled  ClipsState = "enabled"
	ClipsDisabled ClipsState = "disabled"
	ClipsUnknown  ClipsState = "unknown"
)

// Clips (skeleton).
type Clips struct {
	Probe    func(ctx context.Context) error
	HasKey   func() bool
	Blocked  func() bool
	OnChange func()
}

func (c *Clips) State() ClipsState                               { return ClipsUnknown }
func (c *Clips) Trigger()                                        {}
func (c *Clips) Recheck(ctx context.Context) (ClipsState, error) { return ClipsUnknown, nil }
func (c *Clips) Reset()                                          {}
