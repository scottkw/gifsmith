package services

import (
	"sync"
	"time"

	"gifsmith/internal/settings"
)

// Registrar binds global shortcuts.
type Registrar interface {
	Register(accel string, cb func()) error
	Unregister(accel string) error
}

const DefaultOpenHotkey = settings.DefaultOpenHotkey

// Hotkeys owns Gifsmith's global shortcuts.
type Hotkeys struct {
	Reg     Registrar
	Actions map[string]func()
	Emit    func(name string, data ...any)
	Now     func() time.Time

	mu sync.Mutex
}

func (h *Hotkeys) Canonical(accel string) (string, error) { return "", nil }
func (h *Hotkeys) Start(b settings.Hotkeys)               {}
func (h *Hotkeys) Bind(slot, accel string) error          { return nil }
func (h *Hotkeys) Suspend()                               {}
func (h *Hotkeys) Resume()                                {}
func (h *Hotkeys) ArmTest(slot string) error              { return nil }
func (h *Hotkeys) Bindings() settings.Hotkeys             { return settings.Hotkeys{} }
