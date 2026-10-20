// Package settings holds API keys (macOS Keychain behind SecretStore) and the
// non-secret prefs file. No error or log line here ever carries a key.
package settings

import (
	"cmp"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/giphy"
	"gifsmith/internal/provider/klipy"

	"github.com/zalando/go-keyring"
)

// KeychainService is the Keychain item service; accounts are "giphy" and "klipy".
const KeychainService = "com.kenscott.gifsmith"

var ErrNoKey = errors.New("no key stored")

type SecretStore interface {
	Get(account string) (string, error) // ErrNoKey when absent
	Set(account, secret string) error
	Delete(account string) error // nil when absent
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

// Delete treats an absent item as deleted.
func (s KeyringStore) Delete(account string) error {
	if err := keyring.Delete(s.Service, account); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return errors.New("keychain delete failed")
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

func (s *MemoryStore) Delete(account string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, account)
	return nil
}

// Keys caches the stored keys: each Keychain read spawns /usr/bin/security, so read once at startup.
type Keys struct {
	store    SecretStore
	mu       sync.RWMutex
	m        map[provider.ID]string
	rejected map[provider.ID]bool // the vendor refused the stored key; cleared by Set, Delete, ClearRejected
}

func NewKeys(s SecretStore) *Keys {
	return &Keys{store: s, m: map[provider.ID]string{}, rejected: map[provider.ID]bool{}}
}

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
	delete(k.rejected, p)
	k.mu.Unlock()
	return nil
}

// Delete removes p's key from the store, then the cache; a store error leaves both unchanged.
func (k *Keys) Delete(p provider.ID) error {
	if err := k.store.Delete(string(p)); err != nil {
		return err
	}
	k.mu.Lock()
	delete(k.m, p)
	delete(k.rejected, p)
	k.mu.Unlock()
	return nil
}

// MarkRejected records that the vendor refused p's stored key (a failed Test or search).
func (k *Keys) MarkRejected(p provider.ID) { k.setRejected(p, true) }

// ClearRejected records that p's stored key worked again (a passing Test).
func (k *Keys) ClearRejected(p provider.ID) { k.setRejected(p, false) }

func (k *Keys) setRejected(p provider.ID, on bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if _, ok := k.m[p]; ok && on {
		k.rejected[p] = true
	} else {
		delete(k.rejected, p)
	}
}

// Rejected reports whether p's stored key was refused since it was last set.
func (k *Keys) Rejected(p provider.ID) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.rejected[p]
}

// Func is a provider client's Key function; it sees later Sets.
func (k *Keys) Func(p provider.ID) func() (string, bool) {
	return func() (string, bool) { return k.Get(p) }
}

// Prefs is everything in settings.json. Never add a secret here.
type Prefs struct {
	LastProvider  string  `json:"lastProvider"`
	Theme         string  `json:"theme,omitempty"` // "" = system | "light" | "dark"
	LaunchAtLogin bool    `json:"launchAtLogin,omitempty"`
	GiphyRating   string  `json:"giphyRating,omitempty"`  // "" = pg-13
	PageSize      int     `json:"pageSize,omitempty"`     // 0 = 24
	GiphyLang     string  `json:"giphyLang,omitempty"`    // "" = match system
	KlipyCountry  string  `json:"klipyCountry,omitempty"` // "" = match system
	Hotkeys       Hotkeys `json:"hotkeys,omitzero"`
	DownloadDir   string  `json:"downloadDir,omitempty"` // "" = the default Downloads folder; else absolute
	HideFilenames bool    `json:"hideFilenames,omitempty"`
	// ↵ and ⌘↵ actions, one of Actions; "" = DefaultPrimary / DefaultSecondary. Never equal.
	PrimaryAction   string `json:"primaryAction,omitempty"`
	SecondaryAction string `json:"secondaryAction,omitempty"`
	GridSearch      string `json:"gridSearch,omitempty"`   // "" = medium | small | large; also Favorites and Recents
	GridTrending    string `json:"gridTrending,omitempty"` // "" = medium | small | large
	// The zero value means automatic update checks are on (D-28); SetPref "updateCheck" stores it inverted.
	UpdateCheckOff bool `json:"updateCheckOff,omitempty"`
}

// Actions are the ↵/⌘↵ choices in Settings order (no square actions).
var Actions = []string{"copy", "paste", "copyLink", "pasteLink", "copyMarkdown", "pasteMarkdown", "download"}

// GridSizes are the panel grid sizes; "" stores the default, medium.
var GridSizes = []string{"small", "medium", "large"}

const (
	DefaultPrimary   = "copy"
	DefaultSecondary = "paste"
)

// Hotkeys are the global shortcuts as canonical Wails accelerators ("Cmd+Ctrl+Option+G").
type Hotkeys struct {
	Open      string `json:"open,omitempty"`      // "" = DefaultOpenHotkey
	Favorites string `json:"favorites,omitempty"` // "" = not set
	Recents   string `json:"recents,omitempty"`   // "" = not set
}

// DefaultOpenHotkey is ⌃⌥⌘G in canonical form.
const DefaultOpenHotkey = "Cmd+Ctrl+Option+G"

var (
	ErrHotkeyInvalid    = errors.New("not a shortcut Gifsmith can register")
	ErrHotkeyNoModifier = errors.New("the shortcut has no Cmd, Option or Ctrl")
	ErrHotkeySystem     = errors.New("the shortcut is a macOS shortcut")
)

var hotkeyMods = map[string]string{"cmd": "Cmd", "ctrl": "Ctrl", "option": "Option", "shift": "Shift"}

// hotkeyNamed are the non-character keys the recorder offers (all bindable by Wails on macOS).
var hotkeyNamed = []string{"space", "up", "down", "left", "right", "return", "tab"}

// systemHotkeys are macOS shortcuts Gifsmith refuses to take over (UI-SPEC H3), canonical.
var systemHotkeys = []string{"Cmd+Q", "Cmd+W", "Cmd+H", "Cmd+M", "Cmd+TAB", "Cmd+SPACE", "Ctrl+SPACE", "Cmd+Shift+3", "Cmd+Shift+4", "Cmd+Shift+5"}

func hotkeyKey(k string) bool {
	k = strings.ToLower(k)
	if len(k) == 1 {
		c := k[0]
		return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.IndexByte("=-][';\\,/.`", c) >= 0
	}
	if n, err := strconv.Atoi(strings.TrimPrefix(k, "f")); err == nil && k == "f"+strconv.Itoa(n) && n >= 1 && n <= 20 {
		return true
	}
	return slices.Contains(hotkeyNamed, k)
}

// CanonicalHotkey validates accel against an allow-list and returns it in Wails' canonical form
// (modifiers sorted Cmd, Ctrl, Option, Shift, then the upper-cased key), the form Wails compares.
func CanonicalHotkey(accel string) (string, error) {
	parts := strings.Split(accel, "+")
	key := parts[len(parts)-1]
	if !hotkeyKey(key) {
		return "", ErrHotkeyInvalid
	}
	var mods []string
	for _, m := range parts[:len(parts)-1] {
		name, ok := hotkeyMods[strings.ToLower(m)]
		if !ok {
			return "", ErrHotkeyInvalid
		}
		if !slices.Contains(mods, name) {
			mods = append(mods, name)
		}
	}
	slices.Sort(mods)
	if !slices.ContainsFunc(mods, func(m string) bool { return m != "Shift" }) {
		return "", ErrHotkeyNoModifier
	}
	canon := strings.Join(append(mods, strings.ToUpper(key)), "+")
	if slices.Contains(systemHotkeys, canon) {
		return "", ErrHotkeySystem
	}
	return canon, nil
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
	if p.GiphyRating != "" && !slices.Contains(provider.Ratings, p.GiphyRating) {
		p.GiphyRating = ""
		reset = append(reset, "giphyRating")
	}
	if p.PageSize != 0 && !slices.Contains(provider.PageSizes, p.PageSize) {
		p.PageSize = 0
		reset = append(reset, "pageSize")
	}
	if p.GiphyLang != "" && !slices.Contains(giphy.Languages, p.GiphyLang) {
		p.GiphyLang = ""
		reset = append(reset, "giphyLang")
	}
	if p.KlipyCountry != "" && !slices.Contains(klipy.Countries, p.KlipyCountry) {
		p.KlipyCountry = ""
		reset = append(reset, "klipyCountry")
	}
	for _, f := range []struct {
		name string
		v    *string
	}{{"primaryAction", &p.PrimaryAction}, {"secondaryAction", &p.SecondaryAction}} {
		if *f.v != "" && !slices.Contains(Actions, *f.v) {
			*f.v = ""
			reset = append(reset, f.name)
		}
	}
	// ↵ and ⌘↵ never run the same action: the secondary gives way first, then the primary.
	if cmp.Or(p.PrimaryAction, DefaultPrimary) == cmp.Or(p.SecondaryAction, DefaultSecondary) {
		if p.SecondaryAction != "" {
			p.SecondaryAction = ""
			reset = append(reset, "secondaryAction")
		} else {
			p.PrimaryAction = ""
			reset = append(reset, "primaryAction")
		}
	}
	for _, f := range []struct {
		name string
		v    *string
	}{{"gridSearch", &p.GridSearch}, {"gridTrending", &p.GridTrending}} {
		if *f.v != "" && !slices.Contains(GridSizes, *f.v) {
			*f.v = ""
			reset = append(reset, f.name)
		}
	}
	if p.DownloadDir != "" && !filepath.IsAbs(p.DownloadDir) { // only Go's picker path is stored, always absolute
		p.DownloadDir = ""
		reset = append(reset, "downloadDir")
	}
	return append(reset, p.Hotkeys.normalize()...)
}

// normalize canonicalises each binding and resets one that is invalid, a macOS shortcut or a
// duplicate of an earlier slot (Open, which defaults to ⌃⌥⌘G, then Favorites, then Recents).
func (h *Hotkeys) normalize() []string {
	var reset, taken []string
	for _, f := range []struct {
		name string
		v    *string
		def  string
	}{{"hotkeys.open", &h.Open, DefaultOpenHotkey}, {"hotkeys.favorites", &h.Favorites, ""}, {"hotkeys.recents", &h.Recents, ""}} {
		if *f.v != "" {
			c, err := CanonicalHotkey(*f.v)
			if err != nil || slices.Contains(taken, c) {
				c = ""
				reset = append(reset, f.name)
			}
			*f.v = c
		}
		if eff := cmp.Or(*f.v, f.def); eff != "" {
			taken = append(taken, eff)
		}
	}
	if h.Open == DefaultOpenHotkey {
		h.Open = ""
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
