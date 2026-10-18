package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func readPrefsFile(t *testing.T, dir string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("settings.json = %q: %v", b, err)
	}
	return m
}

func TestStoreFirstRunAndWrite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "com.kenscott.gifsmith")
	s := OpenStore(dir)
	if got := s.Get(); got != (Prefs{}) {
		t.Fatalf("first run Get = %+v, want zero Prefs", got)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("OpenStore created %s (err %v); first run writes nothing", dir, err)
	}
	p, err := s.Update(func(p *Prefs) error { p.Theme = "dark"; return nil })
	if err != nil || p.Theme != "dark" || s.Get().Theme != "dark" {
		t.Fatalf("Update = %+v, %v; Get = %+v", p, err, s.Get())
	}
	m := readPrefsFile(t, dir)
	if m["theme"] != "dark" {
		t.Errorf("settings.json theme = %v, want dark", m["theme"])
	}
	if _, ok := m["lastProvider"]; !ok {
		t.Errorf("settings.json lacks lastProvider: %v", m)
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
	if got := OpenStore(dir).Get(); got.Theme != "dark" {
		t.Errorf("reopened Get = %+v, want theme dark", got)
	}
}

// Each Update changes only what its mutator touches: the old SavePrefs(Prefs{...}) erased the rest.
func TestStoreUpdateKeepsOtherFields(t *testing.T) {
	dir := t.TempDir()
	s := OpenStore(dir)
	if _, err := s.Update(func(p *Prefs) error { p.LastProvider = "klipy"; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(func(p *Prefs) error { p.Theme = "light"; return nil }); err != nil {
		t.Fatal(err)
	}
	m := readPrefsFile(t, dir)
	if m["lastProvider"] != "klipy" || m["theme"] != "light" {
		t.Errorf("settings.json = %v, want lastProvider klipy and theme light", m)
	}
}

func TestStoreUnknownEnumResets(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"lastProvider":"giphy","theme":"neon"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got := OpenStore(dir).Get()
	if got.Theme != "" || got.LastProvider != "giphy" {
		t.Errorf("Get = %+v, want theme reset to \"\" and lastProvider giphy kept", got)
	}
	if !strings.Contains(buf.String(), "theme") {
		t.Errorf("log %q does not name the reset field", buf.String())
	}
}

func TestStoreUnreadableIsFirstRun(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte("{nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := OpenStore(dir).Get(); got != (Prefs{}) {
		t.Errorf("unreadable file: Get = %+v, want zero Prefs", got)
	}
	if buf.Len() == 0 {
		t.Error("unreadable settings file was not logged")
	}
}

func TestStoreWriteFailureKeepsPrefs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(dir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := OpenStore(dir) // the "dir" is a file: every write fails
	if _, err := s.Update(func(p *Prefs) error { p.Theme = "dark"; return nil }); err == nil {
		t.Fatal("Update into a file-as-dir succeeded")
	}
	if got := s.Get(); got != (Prefs{}) {
		t.Errorf("Get after a failed write = %+v, want unchanged zero Prefs", got)
	}
	mutErr := errors.New("refused")
	if _, err := OpenStore(t.TempDir()).Update(func(p *Prefs) error { p.Theme = "dark"; return mutErr }); !errors.Is(err, mutErr) {
		t.Errorf("mutator error = %v, want it returned", err)
	}
}

func TestStoreLaunchAtLoginRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if _, err := OpenStore(dir).Update(func(p *Prefs) error { p.LaunchAtLogin = true; return nil }); err != nil {
		t.Fatal(err)
	}
	if m := readPrefsFile(t, dir); m["launchAtLogin"] != true {
		t.Errorf("settings.json = %v, want launchAtLogin true", m)
	}
	if !OpenStore(dir).Get().LaunchAtLogin {
		t.Error("reopened store lost launchAtLogin")
	}
}

func TestStoreConcurrentUpdates(t *testing.T) {
	dir := t.TempDir()
	s := OpenStore(dir)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			if i%2 == 0 {
				_, err = s.Update(func(p *Prefs) error { p.LastProvider = "klipy"; return nil })
			} else {
				_, err = s.Update(func(p *Prefs) error { p.Theme = "dark"; return nil })
			}
			if err != nil {
				t.Error(err)
			}
			_ = s.Get()
		}()
	}
	wg.Wait()
	m := readPrefsFile(t, dir)
	if m["lastProvider"] != "klipy" || m["theme"] != "dark" {
		t.Errorf("settings.json = %v, want both fields after concurrent updates", m)
	}
}

func TestWriteJSONAtomicModes(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := WriteJSONAtomic(dir, "list.json", map[string]int{"version": 1}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "list.json"))
	if err != nil || string(b) != `{"version":1}` {
		t.Fatalf("list.json = %q, %v", b, err)
	}
	if fi, _ := os.Stat(filepath.Join(dir, "list.json")); fi.Mode().Perm() != 0o600 {
		t.Errorf("file mode %v, want 0600", fi.Mode().Perm())
	}
	if fi, _ := os.Stat(dir); fi.Mode().Perm() != 0o700 {
		t.Errorf("dir mode %v, want 0700", fi.Mode().Perm())
	}
}

// A failed write keeps the previous file byte for byte and leaves no temp file behind.
func TestWriteJSONAtomicFailureKeepsPrevious(t *testing.T) {
	dir := t.TempDir()
	if err := WriteJSONAtomic(dir, "list.json", []int{1}); err != nil {
		t.Fatal(err)
	}
	// Rename onto a non-empty directory fails after the temp file was written.
	if err := os.MkdirAll(filepath.Join(dir, "busy.json", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSONAtomic(dir, "busy.json", []int{2}); err == nil {
		t.Error("write over a directory succeeded")
	}
	// An unwritable dir fails before anything is created.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	if err := WriteJSONAtomic(dir, "list.json", []int{3}); err == nil {
		t.Error("write into a read-only dir succeeded")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "list.json")); string(b) != "[1]" {
		t.Errorf("list.json = %q, want the previous [1]", b)
	}
	ents, _ := os.ReadDir(dir)
	if len(ents) != 2 {
		t.Errorf("dir holds %d entries, want list.json and busy.json only (no temp files)", len(ents))
	}
}

func TestStoreNormalizeSearchPrefs(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"giphyRating":"x","pageSize":30,"giphyLang":"xx","klipyCountry":"zz"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got := OpenStore(dir).Get()
	if got.GiphyRating != "" || got.PageSize != 0 || got.GiphyLang != "" || got.KlipyCountry != "" {
		t.Errorf("Get = %+v, want every search pref reset", got)
	}
	for _, f := range []string{"giphyRating", "pageSize", "giphyLang", "klipyCountry"} {
		if !strings.Contains(buf.String(), "reset "+f) {
			t.Errorf("log %q does not name %s", buf.String(), f)
		}
	}
	ok := Prefs{GiphyRating: "r", PageSize: 48, GiphyLang: "zh-TW", KlipyCountry: "de"}
	if p := ok; len(p.Normalize()) != 0 || p != ok {
		t.Errorf("valid prefs %+v were reset", ok)
	}
}
