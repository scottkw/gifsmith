package settings

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gifsmith/internal/provider"

	"github.com/zalando/go-keyring"
)

// countingStore records Get calls so tests can prove keys are read once.
type countingStore struct {
	*MemoryStore
	gets int
}

func (c *countingStore) Get(a string) (string, error) {
	c.gets++
	return c.MemoryStore.Get(a)
}

func TestMemoryStore(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.Get("giphy"); !errors.Is(err, ErrNoKey) {
		t.Fatalf("missing account: err = %v, want ErrNoKey", err)
	}
	if err := s.Set("giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if v, err := s.Get("giphy"); err != nil || v != "uat-key-123" {
		t.Errorf("Get = %q, %v", v, err)
	}
}

func TestKeysLoadOnce(t *testing.T) {
	cs := &countingStore{MemoryStore: NewMemoryStore()}
	_ = cs.Set("klipy", "uat-key-123")
	k := NewKeys(cs)
	k.Load()
	for range 3 {
		k.Get(provider.GIPHY)
		k.Get(provider.Klipy)
	}
	if cs.gets != 2 {
		t.Errorf("store Get called %d times, want 2 (once per account at Load)", cs.gets)
	}
	if v, ok := k.Get(provider.Klipy); !ok || v != "uat-key-123" {
		t.Errorf("klipy = %q, %v", v, ok)
	}
	if _, ok := k.Get(provider.GIPHY); ok {
		t.Error("giphy configured without a stored key")
	}
}

func TestKeysSetAndFunc(t *testing.T) {
	s := NewMemoryStore()
	k := NewKeys(s)
	k.Load()
	f := k.Func(provider.GIPHY)
	if _, ok := f(); ok {
		t.Fatal("Func reports a key before Set")
	}
	if err := k.Set(provider.GIPHY, "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if v, ok := f(); !ok || v != "uat-key-123" {
		t.Errorf("Func after Set = %q, %v", v, ok)
	}
	if v, _ := s.Get("giphy"); v != "uat-key-123" {
		t.Errorf("store holds %q, want the key under account giphy", v)
	}
}

// failingStore fails every write, and every read with a non-ErrNoKey error.
type failingStore struct{}

func (failingStore) Get(string) (string, error) { return "", errors.New("boom") }
func (failingStore) Set(string, string) error   { return errors.New("boom") }

func TestKeysStoreFailures(t *testing.T) {
	k := NewKeys(failingStore{})
	k.Load() // logged, not fatal
	if err := k.Set(provider.GIPHY, "uat-key-123"); err == nil {
		t.Fatal("Set succeeded on a failing store")
	}
	if _, ok := k.Get(provider.GIPHY); ok {
		t.Error("cache updated although the store write failed")
	}
}

func TestKeyringStoreErrors(t *testing.T) {
	keyring.MockInit() // in-memory provider: never touches the login Keychain
	s := KeyringStore{Service: "gifsmith-test"}
	if _, err := s.Get("giphy"); !errors.Is(err, ErrNoKey) {
		t.Fatalf("missing item: err = %v, want ErrNoKey", err)
	}
	if err := s.Set("giphy", "SECRETKEY123"); err != nil {
		t.Fatal(err)
	}
	if v, err := s.Get("giphy"); err != nil || v != "SECRETKEY123" {
		t.Fatalf("Get = %q, %v", v, err)
	}
	keyring.MockInitWithError(errors.New("security: SECRETKEY123 exit status 45"))
	_, gerr := s.Get("giphy")
	serr := s.Set("giphy", "SECRETKEY123")
	for _, err := range []error{gerr, serr} {
		if err == nil || errors.Is(err, ErrNoKey) || strings.Contains(err.Error(), "SECRETKEY123") || strings.Contains(err.Error(), "security") {
			t.Errorf("err = %v, want a fixed keychain failure text", err)
		}
	}
}

func TestPrefsRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "com.kenscott.gifsmith")
	p, err := LoadPrefs(dir)
	if err != nil || p != (Prefs{}) {
		t.Fatalf("missing dir: %+v, %v; want zero Prefs, nil", p, err)
	}
	if err := SavePrefs(dir, Prefs{LastProvider: "klipy"}); err != nil {
		t.Fatal(err)
	}
	p, err = LoadPrefs(dir)
	if err != nil || p.LastProvider != "klipy" {
		t.Fatalf("LoadPrefs = %+v, %v", p, err)
	}
	fi, err := os.Stat(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("settings.json mode %v, want 0600", fi.Mode().Perm())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("dir holds %d entries, want only settings.json (no temp files)", len(entries))
	}
}

func TestPrefsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte("{nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPrefs(dir); err == nil {
		t.Error("invalid JSON loaded without error")
	}
}
