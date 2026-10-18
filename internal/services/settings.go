package services

import (
	"context"
	"errors"
	"log"
	"slices"
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
	GiphyRating   string `json:"giphyRating"`  // resolved: never ""
	PageSize      int    `json:"pageSize"`     // resolved: never 0
	GiphyLang     string `json:"giphyLang"`    // "" = Match System
	KlipyCountry  string `json:"klipyCountry"` // "" = Match System
}

// Choices are the allowed Search-pane values, with what Match System resolves to now.
type Choices struct {
	Ratings        []string `json:"ratings"`
	PageSizes      []int    `json:"pageSizes"`
	Languages      []string `json:"languages"`
	Countries      []string `json:"countries"`
	SystemLanguage string   `json:"systemLanguage"`
	SystemCountry  string   `json:"systemCountry"`
}

// Choices lists the Search-pane options; the webview names them.
func (s *SettingsService) Choices(ctx context.Context) Choices { return Choices{} }

// Params are the provider parameters for the next request: the defaults overlaid with the
// stored choices. Clients call it on every request.
func (s *SettingsService) Params() provider.Params { return s.paramsOf(s.Store.Get()) }

func (s *SettingsService) paramsOf(p settings.Prefs) provider.Params {
	prm := provider.DefaultParams()
	if p.GiphyRating != "" {
		prm.Rating = p.GiphyRating
	}
	return prm
}

func (s *SettingsService) Snapshot(ctx context.Context) Snapshot { return s.snapshotOf(s.Store.Get()) }

// snapshotOf reports launchAtLogin from the OS, not the pref: the checkbox shows what will happen.
func (s *SettingsService) snapshotOf(p settings.Prefs) Snapshot {
	on, err := s.Autostart.IsEnabled()
	if err != nil {
		log.Printf("autostart: status failed")
	}
	prm := s.paramsOf(p)
	return Snapshot{Theme: p.Theme, LaunchAtLogin: on && err == nil, GiphyRating: prm.Rating}
}

var errUnchanged = errors.New("unchanged")

// SetPref sets one allow-listed pref from the webview, validated here; an equal value is a no-op.
// Each write emits settings:changed with the full Snapshot so the panel and Settings converge.
// A provider parameter also flushes the affected providers' cached pages (never their 429 block).
func (s *SettingsService) SetPref(ctx context.Context, name, value string) (Snapshot, error) {
	var set func(*settings.Prefs)
	var flush []provider.ID
	switch name {
	case "theme":
		set = func(p *settings.Prefs) { p.Theme = value }
	case "giphyRating":
		if !slices.Contains(provider.Ratings, value) {
			return s.Snapshot(ctx), errInvalidValue
		}
		set = func(p *settings.Prefs) { p.GiphyRating = defaultAsEmpty(value, provider.DefaultParams().Rating) }
		flush = []provider.ID{provider.GIPHY} // Clips share GIPHY's budget, so their pages go too
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
		return s.snapshotOf(p), nil
	case errors.Is(err, errInvalidValue):
		return s.snapshotOf(p), err
	case err != nil:
		log.Printf("settings write failed: %s", name)
		return s.snapshotOf(p), errPrefsWriteFailed
	}
	if name == "theme" && s.ApplyTheme != nil {
		s.ApplyTheme(p.Theme)
	}
	for _, id := range flush {
		s.Svc.Flush(id)
	}
	snap := s.snapshotOf(p)
	if s.Emit != nil {
		s.Emit("settings:changed", snap)
	}
	return snap, nil
}

// defaultAsEmpty stores the default as "", so choosing it when nothing is stored is a no-op.
func defaultAsEmpty(v, def string) string {
	if v == def {
		return ""
	}
	return v
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

// SetLaunchAtLogin registers or removes the login item now and persists the choice for ApplyStartup.
func (s *SettingsService) SetLaunchAtLogin(ctx context.Context, on bool) (Snapshot, error) {
	if on {
		// A copy run from the DMG or a translocated path would register a login item that goes away.
		exe, err := s.Executable()
		if err == nil && (strings.Contains(exe, "/AppTranslocation/") || strings.HasPrefix(exe, "/Volumes/")) {
			log.Printf("autostart: not enabled, the app is not in Applications")
			return s.Snapshot(ctx), errNotInApplications
		}
	}
	op, apply := "disable", s.Autostart.Disable
	if on {
		op, apply = "enable", s.Autostart.Enable
	}
	if err := apply(); err != nil {
		log.Printf("autostart: %s failed", op)
		return s.Snapshot(ctx), errAutostartFailed
	}
	p, err := s.Store.Update(func(p *settings.Prefs) error { p.LaunchAtLogin = on; return nil })
	if err != nil {
		log.Printf("settings write failed: launchAtLogin")
		return s.snapshotOf(p), errPrefsWriteFailed
	}
	snap := s.snapshotOf(p)
	if s.Emit != nil {
		s.Emit("settings:changed", snap)
	}
	return snap, nil
}

// ApplyStartup runs once at launch: when the pref is on and the OS reports the login item off,
// register it again. It never disables. Risk (A4): IsEnabled also reports "off" for an item the user
// switched off in System Settings, so this re-registers it; verify natively what register does then.
func (s *SettingsService) ApplyStartup() {
	if !s.Store.Get().LaunchAtLogin {
		return
	}
	on, err := s.Autostart.IsEnabled()
	switch {
	case err != nil:
		log.Printf("autostart: status failed at startup")
	case on:
		log.Printf("autostart: login item registered")
	default:
		if err := s.Autostart.Enable(); err != nil {
			log.Printf("autostart: re-enable at startup failed")
			return
		}
		log.Printf("autostart: re-enabled at startup")
	}
}

// AccessibilityStatus reports the Automatic paste grant without prompting.
func (s *SettingsService) AccessibilityStatus(ctx context.Context) bool {
	return s.Platform.AccessibilityTrusted(false)
}
