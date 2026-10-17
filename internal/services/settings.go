package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"gifsmith/internal/provider"
	"gifsmith/internal/search"
	"gifsmith/internal/settings"
)

var errNoKeyProvider = errors.New("this provider does not use a key")

// SettingsService saves verified keys and the last provider. No bound method
// returns a key; errors and log lines name the provider and kind only.
type SettingsService struct {
	Keys      *settings.Keys
	Dir       string // holds settings.json
	Svc       *search.Service
	NewClient func(p provider.ID, key string) provider.Provider // a client using exactly this candidate key
	Emit      func(name string)
	Clips     *search.Clips

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

// LastProvider is the persisted provider id, or "" on first run or an unreadable file.
func (s *SettingsService) LastProvider(ctx context.Context) string {
	prefs, err := settings.LoadPrefs(s.Dir)
	if err != nil {
		log.Printf("settings unreadable, treating as first run: %v", err)
		return ""
	}
	if _, ok := provider.ParseID(prefs.LastProvider); !ok {
		return ""
	}
	return prefs.LastProvider
}

// SetLastProvider persists the provider tab to reopen on next launch.
func (s *SettingsService) SetLastProvider(ctx context.Context, p string) error {
	id, err := parse(p)
	if err != nil {
		return err
	}
	// No lock: each write is its own temp file + rename, so the last one wins whole.
	return settings.SavePrefs(s.Dir, settings.Prefs{LastProvider: string(id)})
}
