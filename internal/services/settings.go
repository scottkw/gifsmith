package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"gifsmith/internal/platform"
	"gifsmith/internal/provider"
	"gifsmith/internal/search"
	"gifsmith/internal/settings"
)

var errNoKeyProvider = errors.New("this provider does not use a key")

// SettingsService saves verified keys and the last provider. No bound method
// returns a key; errors and log lines name the provider and kind only.
type SettingsService struct {
	Keys       *settings.Keys
	Store      *settings.Store // the only owner of settings.json
	Platform   platform.Platform
	Svc        *search.Service
	NewClient  func(p provider.ID, key string) provider.Provider // a client using exactly this candidate key
	Emit       func(name string, data ...any)
	Clips      *search.Clips
	ApplyTheme func(theme string) // native appearance + window backgrounds; nil in tests
	Autostart  Autostarter
	Executable func() (string, error) // os.Executable; injectable for the not-in-Applications check

	mu sync.Mutex // serializes SaveKey end to end: a verify never interleaves with another save's write
}

// validKey bounds what a key may contain: Klipy puts it in a URL path, GIPHY in a query.
func validKey(k string) bool {
	if n := utf8.RuneCountInString(k); n < 8 || n > 128 {
		return false
	}
	return !strings.ContainsAny(k, "/?#%") && !strings.ContainsFunc(k, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) })
}

func kindOf(err error) string {
	var pe *provider.Error
	if errors.As(err, &pe) {
		return string(pe.Kind)
	}
	return "error"
}

// SaveKey verifies key with a real first-page trending call and stores it only if that succeeds.
// The verified page seeds the trending cache, so verification costs no extra request.
func (s *SettingsService) SaveKey(ctx context.Context, p, key string) error {
	id, ok := provider.ParseID(p)
	if !ok {
		return errUnknownProvider
	}
	if id != provider.GIPHY && id != provider.Klipy {
		return errNoKeyProvider
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	key = strings.TrimSpace(key)
	if !validKey(key) {
		log.Printf("key rejected for %s: %s", id, provider.KeyRejected)
		return &provider.Error{Kind: provider.KeyRejected, Provider: id}
	}
	page, err := s.NewClient(id, key).Trending(ctx, "")
	if err != nil {
		log.Printf("key rejected for %s: %s", id, kindOf(err))
		return err
	}
	if err := s.Keys.Set(id, key); err != nil {
		log.Printf("key not saved for %s: keychain write failed", id)
		return errors.New("couldn't save the key to the macOS Keychain")
	}
	s.Svc.Invalidate(id)
	s.Svc.Seed(id, "", page)
	log.Printf("key saved for %s", id)
	if id == provider.GIPHY {
		// The new key just passed a GIFs call, so its Clips probe is meaningful now.
		s.Clips.Reset()
		s.Clips.Trigger()
	}
	if s.Emit != nil {
		s.Emit("providers:changed")
	}
	return nil
}

// Autostarter is the login item (*application.AutostartManager in the app, a fake in the UAT build).
type Autostarter interface {
	Enable() error
	Disable() error
	IsEnabled() (bool, error)
}

// Snapshot is the settings the webviews render. Never a key.
type Snapshot struct {
	Theme         string `json:"theme"`
	LaunchAtLogin bool   `json:"launchAtLogin"`
}

func (s *SettingsService) Snapshot(ctx context.Context) Snapshot { return snapshotOf(s.Store.Get()) }

func snapshotOf(p settings.Prefs) Snapshot { return Snapshot{Theme: p.Theme} }

var errUnchanged = errors.New("unchanged")

// SetPref sets one allow-listed pref from the webview, validated here; an equal value is a no-op.
// Each write emits settings:changed with the full Snapshot so the panel and Settings converge.
func (s *SettingsService) SetPref(ctx context.Context, name, value string) (Snapshot, error) {
	var set func(*settings.Prefs)
	switch name {
	case "theme":
		set = func(p *settings.Prefs) { p.Theme = value }
	default:
		return s.Snapshot(ctx), errInvalidValue
	}
	p, err := s.Store.Update(func(p *settings.Prefs) error {
		before := *p
		set(p)
		if len(p.Normalize()) > 0 {
			return errInvalidValue
		}
		if *p == before {
			return errUnchanged
		}
		return nil
	})
	switch {
	case errors.Is(err, errUnchanged):
		return snapshotOf(p), nil
	case errors.Is(err, errInvalidValue):
		return snapshotOf(p), err
	case err != nil:
		log.Printf("settings write failed: %s", name)
		return snapshotOf(p), errPrefsWriteFailed
	}
	if name == "theme" && s.ApplyTheme != nil {
		s.ApplyTheme(p.Theme)
	}
	snap := snapshotOf(p)
	if s.Emit != nil {
		s.Emit("settings:changed", snap)
	}
	return snap, nil
}

// LastProvider is the persisted provider id, or "" on first run or an unreadable file.
func (s *SettingsService) LastProvider(ctx context.Context) string {
	lp := s.Store.Get().LastProvider
	if _, ok := provider.ParseID(lp); !ok {
		return ""
	}
	return lp
}

// SetLastProvider persists the provider tab to reopen on next launch (no event: a tab switch).
func (s *SettingsService) SetLastProvider(ctx context.Context, p string) error {
	id, err := parse(p)
	if err != nil {
		return err
	}
	if _, err := s.Store.Update(func(p *settings.Prefs) error { p.LastProvider = string(id); return nil }); err != nil {
		log.Printf("settings write failed: lastProvider")
		return errPrefsWriteFailed
	}
	return nil
}

func (s *SettingsService) SetLaunchAtLogin(ctx context.Context, on bool) (Snapshot, error) {
	return Snapshot{}, nil // RED skeleton
}

func (s *SettingsService) ApplyStartup() {} // RED skeleton

func (s *SettingsService) AccessibilityStatus(ctx context.Context) bool { return false } // RED skeleton
