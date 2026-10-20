package services

import (
	"cmp"
	"errors"
	"log"
	"sync"
	"time"

	"gifsmith/internal/settings"
)

// Registrar binds global shortcuts: *application.GlobalShortcutManager in the app, a fake in the
// -tags server build and in tests. It only ever sees canonical accelerators.
type Registrar interface {
	Register(accel string, cb func()) error
	Unregister(accel string) error
}

// DefaultOpenHotkey is ⌃⌥⌘G, canonical.
const DefaultOpenHotkey = settings.DefaultOpenHotkey

const hotkeyTestWindow = 5 * time.Second

var hotkeySlots = []string{"open", "favorites", "recents"}

// Hotkeys owns Gifsmith's global shortcuts, one per slot (open, favorites, recents).
type Hotkeys struct {
	Reg     Registrar
	Actions map[string]func() // slot → what its hotkey does
	Emit    func(name string, data ...any)
	Now     func() time.Time

	mu        sync.Mutex
	bound     settings.Hotkeys     // canonical; what is registered unless suspended
	failed    settings.Hotkeys     // stored combos macOS refused at Start; cleared when the slot is rebound
	armed     map[string]time.Time // slot → Test deadline
	suspended bool
}

func slotOf(b *settings.Hotkeys, slot string) *string {
	switch slot {
	case "open":
		return &b.Open
	case "favorites":
		return &b.Favorites
	case "recents":
		return &b.Recents
	}
	return nil
}

var (
	errHotkeyInvalid    = &SettingsError{Kind: "hotkey_invalid"}
	errHotkeyNoModifier = &SettingsError{Kind: "hotkey_no_modifier"}
	errHotkeySystem     = &SettingsError{Kind: "hotkey_system"}
	errHotkeyFailed     = &SettingsError{Kind: "hotkey_register_failed"}
	errHotkeyFailedNew  = &SettingsError{Kind: "hotkey_register_failed_new"}
	errHotkeyFailedLost = &SettingsError{Kind: "hotkey_register_failed_lost"}
)

// Canonical validates accel (allow-listed keys, at least one of Cmd/Option/Ctrl, not a macOS
// shortcut) and returns Wails' canonical form.
func (h *Hotkeys) Canonical(accel string) (string, error) {
	c, err := settings.CanonicalHotkey(accel)
	switch {
	case errors.Is(err, settings.ErrHotkeyNoModifier):
		return "", errHotkeyNoModifier
	case errors.Is(err, settings.ErrHotkeySystem):
		return "", errHotkeySystem
	case err != nil:
		return "", errHotkeyInvalid
	}
	return c, nil
}

// callback is what Carbon runs for slot: it reports an armed Test, then does the slot's action.
func (h *Hotkeys) callback(slot string) func() {
	return func() {
		h.mu.Lock()
		deadline, armed := h.armed[slot]
		delete(h.armed, slot)
		h.mu.Unlock()
		if armed && h.Now().Before(deadline) && h.Emit != nil {
			h.Emit("hotkey:test", slot)
		}
		if a := h.Actions[slot]; a != nil {
			a()
		}
	}
}

func (h *Hotkeys) register(slot, accel string) error {
	err := h.Reg.Register(accel, h.callback(slot))
	if err != nil {
		log.Printf("hotkey register error: %s: %v", accel, err)
	}
	return err
}

// Start registers the stored bindings (Open falls back to the default) once the app runs.
func (h *Hotkeys) Start(b settings.Hotkeys) {
	h.mu.Lock()
	defer h.mu.Unlock()
	b.Open = cmp.Or(b.Open, DefaultOpenHotkey)
	for _, slot := range hotkeySlots {
		accel := *slotOf(&b, slot)
		if accel == "" {
			continue
		}
		c, err := h.Canonical(accel)
		if err != nil {
			log.Printf("hotkey for %s not registered: %s", slot, err)
			continue
		}
		if h.register(slot, c) != nil {
			*slotOf(&h.failed, slot) = c // not bound; the Hotkeys pane shows it
			continue
		}
		*slotOf(&h.bound, slot) = c
		log.Printf("hotkey registered: %s", c)
	}
}

// Bind rebinds slot to accel ("" clears favorites/recents). It refuses invalid, system and
// duplicate combos before touching the Registrar; a failed registration re-registers the old one.
func (h *Hotkeys) Bind(slot, accel string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	cur := slotOf(&h.bound, slot)
	if cur == nil || (accel == "" && slot == "open") {
		return errHotkeyInvalid
	}
	c := ""
	if accel != "" {
		var err error
		if c, err = h.Canonical(accel); err != nil {
			return err
		}
		for _, other := range hotkeySlots {
			if other != slot && *slotOf(&h.bound, other) == c {
				return &SettingsError{Kind: "hotkey_duplicate", Slot: other}
			}
		}
	}
	old := *cur
	if err := h.rebind(slot, c); err != nil {
		if err == errHotkeyFailedLost {
			*slotOf(&h.failed, slot) = old // still stored, no longer registered
		}
		return err
	}
	*slotOf(&h.failed, slot) = ""
	return nil
}

// Revert puts slot back to old, the value Bindings reported before a Bind the caller could not store.
func (h *Hotkeys) Revert(slot, old string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rebind(slot, old)
}

// rebind moves slot from its current combo to c (canonical, already checked); h.mu is held.
func (h *Hotkeys) rebind(slot, c string) error {
	cur := slotOf(&h.bound, slot)
	old := *cur
	if c == old {
		return nil
	}
	if !h.suspended {
		if old != "" {
			if err := h.Reg.Unregister(old); err != nil {
				log.Printf("hotkey unregister error: %s: %v", old, err)
			}
		}
		if c != "" && h.register(slot, c) != nil {
			if old == "" {
				return errHotkeyFailedNew
			}
			if err := h.Reg.Register(old, h.callback(slot)); err != nil {
				log.Printf("hotkey restore failed: %s: %v", old, err)
				*cur = ""
				return errHotkeyFailedLost
			}
			return errHotkeyFailed
		}
	}
	*cur = c
	return nil
}

// Suspend unregisters every binding (a recorder is capturing keys); idempotent.
func (h *Hotkeys) Suspend() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.suspended {
		return
	}
	h.suspended = true
	for _, slot := range hotkeySlots {
		if c := *slotOf(&h.bound, slot); c != "" {
			if err := h.Reg.Unregister(c); err != nil {
				log.Printf("hotkey unregister error: %s: %v", c, err)
			}
		}
	}
}

// Resume registers every binding again; idempotent.
func (h *Hotkeys) Resume() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.suspended {
		return
	}
	h.suspended = false
	for _, slot := range hotkeySlots {
		if c := *slotOf(&h.bound, slot); c != "" {
			_ = h.register(slot, c) // logged
		}
	}
}

// ArmTest makes slot's next press within 5 s emit hotkey:test.
func (h *Hotkeys) ArmTest(slot string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if p := slotOf(&h.bound, slot); p == nil || *p == "" {
		return errHotkeyInvalid
	}
	if h.armed == nil {
		h.armed = map[string]time.Time{}
	}
	h.armed[slot] = h.Now().Add(hotkeyTestWindow)
	return nil
}

// Bindings are the current canonical bindings.
func (h *Hotkeys) Bindings() settings.Hotkeys {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.bound
}

// State is Bindings plus the stored combos macOS refused, which are not bound.
func (h *Hotkeys) State() (bound, failed settings.Hotkeys) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.bound, h.failed
}
