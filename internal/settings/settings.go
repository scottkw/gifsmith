// Package settings holds API keys (macOS Keychain behind SecretStore) and the
// non-secret prefs file. No error or log line here ever carries a key.
package settings

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gifsmith/internal/provider"

	"github.com/zalando/go-keyring"
)

// KeychainService is the Keychain item service; accounts are "giphy" and "klipy".
const KeychainService = "com.kenscott.gifsmith"

var ErrNoKey = errors.New("no key stored")

type SecretStore interface {
	Get(account string) (string, error) // ErrNoKey when absent
	Set(account, secret string) error
}

// KeyringStore is the macOS Keychain via go-keyring (/usr/bin/security, secret on stdin).
type KeyringStore struct{ Service string }

// Errors are fixed text: the security tool's output could echo the secret.
func (s KeyringStore) Get(account string) (string, error) {
	v, err := keyring.Get(s.Service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNoKey
	}
	if err != nil {
		return "", errors.New("keychain read failed")
	}
	return v, nil
}

func (s KeyringStore) Set(account, secret string) error {
	if err := keyring.Set(s.Service, account, secret); err != nil {
		return errors.New("keychain write failed")
	}
	return nil
}

// MemoryStore is the SecretStore for tests and the -tags server UAT build.
type MemoryStore struct {
	mu sync.Mutex
	m  map[string]string
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{m: map[string]string{}} }

func (s *MemoryStore) Get(account string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.m[account]
	if !ok {
		return "", ErrNoKey
	}
	return v, nil
}

func (s *MemoryStore) Set(account, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[account] = secret
	return nil
}

// Keys caches the stored keys: each Keychain read spawns /usr/bin/security, so read once at startup.
type Keys struct {
	store SecretStore
	mu    sync.RWMutex
	m     map[provider.ID]string
}

func NewKeys(s SecretStore) *Keys { return &Keys{store: s, m: map[provider.ID]string{}} }

func (k *Keys) Load() {
	for _, p := range []provider.ID{provider.GIPHY, provider.Klipy} {
		v, err := k.store.Get(string(p))
		if err != nil {
			if !errors.Is(err, ErrNoKey) {
				log.Printf("keychain read failed for %s", p)
			}
			continue
		}
		k.mu.Lock()
		k.m[p] = v
		k.mu.Unlock()
	}
}

func (k *Keys) Get(p provider.ID) (string, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	v, ok := k.m[p]
	return v, ok
}

// Set writes the store first; the cache changes only when the write succeeded.
func (k *Keys) Set(p provider.ID, key string) error {
	if err := k.store.Set(string(p), key); err != nil {
		return err
	}
	k.mu.Lock()
	k.m[p] = key
	k.mu.Unlock()
	return nil
}

// Func is a provider client's Key function; it sees later Sets.
func (k *Keys) Func(p provider.ID) func() (string, bool) {
	return func() (string, bool) { return k.Get(p) }
}

// Prefs is everything in settings.json. Never add a secret here.
type Prefs struct {
	LastProvider  string `json:"lastProvider"`
	Theme         string `json:"theme,omitempty"` // "" = system | "light" | "dark"
	LaunchAtLogin bool   `json:"launchAtLogin,omitempty"`
	GiphyRating   string `json:"giphyRating,omitempty"`  // "" = pg-13
	PageSize      int    `json:"pageSize,omitempty"`     // 0 = 24
	GiphyLang     string `json:"giphyLang,omitempty"`    // "" = match system
	KlipyCountry  string `json:"klipyCountry,omitempty"` // "" = match system
}

// Normalize resets invalid values to their declared defaults and returns the reset field names.
func (p *Prefs) Normalize() []string {
	var reset []string
	switch p.Theme {
	case "", "light", "dark":
	default:
		p.Theme = ""
		reset = append(reset, "theme")
	}
	return reset
}

const prefsFile = "settings.json"

// Store is the single owner of settings.json: every write is a load-modify-save under one mutex,
// so concurrent writers (panel and Settings) never erase each other's fields.
type Store struct {
	dir string
	mu  sync.Mutex
	p   Prefs
}

// OpenStore reads dir/settings.json once. A missing file is the first run; an unreadable one is
// treated as the first run too, logged. Unknown values reset to their defaults, logged by name.
func OpenStore(dir string) *Store {
	s := &Store{dir: dir}
	b, err := os.ReadFile(filepath.Join(dir, prefsFile))
	if err == nil {
		err = json.Unmarshal(b, &s.p)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("settings unreadable, treating as first run: %v", err)
		s.p = Prefs{}
	}
	for _, f := range s.p.Normalize() {
		log.Printf("settings: reset %s to default", f)
	}
	return s
}

func (s *Store) Get() Prefs {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.p
}

// Update runs f on a copy, normalizes and writes it atomically, and keeps it only when the write
// succeeded; on any error the in-memory prefs are unchanged and the current prefs are returned.
func (s *Store) Update(f func(*Prefs) error) (Prefs, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.p
	if err := f(&next); err != nil {
		return s.p, err
	}
	next.Normalize()
	if err := WriteJSONAtomic(s.dir, prefsFile, next); err != nil {
		return s.p, err
	}
	s.p = next
	return next, nil
}

// WriteJSONAtomic writes v as JSON to dir/name: temp file in the same dir, Sync, then rename,
// mode 0600 in a 0700 dir. On any error the previous file is untouched and the temp file removed.
func WriteJSONAtomic(dir, name string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, name+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(b)
	if err == nil {
		err = f.Chmod(0o600)
	}
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, filepath.Join(dir, name))
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}
