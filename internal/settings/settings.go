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
	LastProvider string `json:"lastProvider"`
}

const prefsFile = "settings.json"

// LoadPrefs reads dir/settings.json; a missing file is the first run (zero Prefs, nil).
func LoadPrefs(dir string) (Prefs, error) {
	var p Prefs
	b, err := os.ReadFile(filepath.Join(dir, prefsFile))
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	return p, json.Unmarshal(b, &p)
}

// SavePrefs writes atomically (temp file in the same dir, then rename), mode 0600 in a 0700 dir.
func SavePrefs(dir string, p Prefs) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, prefsFile+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(b)
	if err == nil {
		err = f.Chmod(0o600)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, filepath.Join(dir, prefsFile))
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}

// WriteJSONAtomic is a skeleton (RED).
func WriteJSONAtomic(dir, name string, v any) error { return nil }
