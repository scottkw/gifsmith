package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gifsmith/internal/platform"
	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
	"gifsmith/internal/provider/finergifs"
	"gifsmith/internal/provider/giphy"
	"gifsmith/internal/provider/klipy"
	"gifsmith/internal/search"
	"gifsmith/internal/settings"
)

type settingsFixture struct {
	svc    *SettingsService
	gifs   *GifService
	store  *settings.MemoryStore
	hits   *atomic.Int32 // fixture requests, Clips routes excluded (the probe is async)
	events *[]string
	data   *[]any // the payload of each event (nil when none), parallel to events
	dir    string // holds settings.json
	clips  *search.Clips
	probes *atomic.Int32
}

// newSettings wires SettingsService and GifService the way main.go does, against fakeapi.
func newSettings(t *testing.T, base string) settingsFixture {
	t.Helper()
	var hits atomic.Int32
	if base == "" {
		h := fakeapi.Handler(fakeapi.Options{})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/clips/") {
				hits.Add(1)
			}
			h.ServeHTTP(w, r)
		}))
		t.Cleanup(srv.Close)
		base = srv.URL
	}
	hc := &http.Client{Timeout: 2 * time.Second}
	var s *SettingsService // set below; clients read its Params per request, as in main.go
	params := func() provider.Params { return s.Params() }
	store := settings.NewMemoryStore()
	keys := settings.NewKeys(store)
	keys.Load()
	newClient := func(p provider.ID, key string) provider.Provider {
		k := func() (string, bool) { return key, true }
		if p == provider.Klipy {
			return &klipy.Client{BaseURL: base + "/klipy", Key: k, HTTP: hc, Params: params}
		}
		return &giphy.Client{BaseURL: base + "/giphy", Key: k, HTTP: hc, Params: params}
	}
	clipsClient := &giphy.Client{BaseURL: base + "/giphy", Key: keys.Func(provider.GIPHY), HTTP: hc, Clips: true, Params: params}
	svc := search.New(map[provider.ID]provider.Provider{
		provider.GIPHY:     &giphy.Client{BaseURL: base + "/giphy", Key: keys.Func(provider.GIPHY), HTTP: hc, Params: params},
		provider.Clips:     clipsClient,
		provider.Klipy:     &klipy.Client{BaseURL: base + "/klipy", Key: keys.Func(provider.Klipy), HTTP: hc, Params: params},
		provider.FinerGifs: &finergifs.Client{BaseURL: base + "/finer", MediaBase: base + "/media", HTTP: hc, Params: params},
	}, nil)
	var probes atomic.Int32
	clips := &search.Clips{
		Probe: func(ctx context.Context) error {
			probes.Add(1)
			return svc.Guard(provider.Clips, func() error { return clipsClient.ProbeClips(ctx) }) // as main.go
		},
		HasKey:  func() bool { _, ok := keys.Get(provider.GIPHY); return ok },
		Blocked: func() error { return svc.BlockedErr(provider.Clips) }, // as main.go
	}
	var mu sync.Mutex
	events := []string{}
	data := []any{}
	dir := t.TempDir()
	s = &SettingsService{Keys: keys, Store: settings.OpenStore(dir), Platform: &platform.Fake{}, Svc: svc, NewClient: newClient, Emit: func(n string, d ...any) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, n)
		if len(d) > 0 {
			data = append(data, d[0])
		} else {
			data = append(data, nil)
		}
	}, Clips: clips, Autostart: &fakeAutostart{}, Executable: func() (string, error) {
		return "/Applications/Gifsmith.app/Contents/MacOS/gifsmith", nil
	}}
	return settingsFixture{svc: s, gifs: &GifService{Svc: svc, Keys: keys, Clips: clips}, store: store, hits: &hits, events: &events, data: &data, dir: dir, clips: clips, probes: &probes}
}

func wantProviderErr(t *testing.T, err error, kind provider.Kind, p provider.ID, status int) {
	t.Helper()
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != kind || pe.Provider != p || pe.Status != status {
		t.Fatalf("err = %#v, want %s %s status %d", err, p, kind, status)
	}
}

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

func TestSaveKeyVerifiesAndSeeds(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "  uat-key-123  "); err != nil {
		t.Fatal(err)
	}
	if v, err := f.store.Get("giphy"); err != nil || v != "uat-key-123" {
		t.Errorf("store = %q, %v; want the trimmed key", v, err)
	}
	if got := *f.events; len(got) != 1 || got[0] != "providers:changed" {
		t.Errorf("events = %v, want [providers:changed]", got)
	}
	before := f.hits.Load()
	p, err := f.gifs.Trending(ctx, "giphy", "")
	if err != nil {
		t.Fatal(err)
	}
	if n := f.hits.Load() - before; n != 0 {
		t.Errorf("Trending after SaveKey made %d requests, want 0 (seeded)", n)
	}
	if len(p.Items) != 24 || !strings.Contains(p.Items[0].PreviewURL, "/media/giphy-") {
		t.Errorf("seeded page = %d items, first %+v", len(p.Items), p.Items)
	}
}

func TestSaveKeyKlipy(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	err := f.svc.SaveKey(ctx, "klipy", "bad-key-123")
	wantProviderErr(t, err, provider.KeyRejected, provider.Klipy, 404)
	if err := f.svc.SaveKey(ctx, "klipy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	before := f.hits.Load()
	p, err := f.gifs.Trending(ctx, "klipy", "")
	if err != nil {
		t.Fatal(err)
	}
	if n := f.hits.Load() - before; n != 0 {
		t.Errorf("Klipy Trending after SaveKey made %d requests, want 0 (seeded)", n)
	}
	dups := 0
	for _, g := range p.Items {
		if g.ID == "klipy-dup" {
			dups++
		}
	}
	if dups != 2 {
		t.Errorf("klipy-dup appears %d times in the seeded page, want 2 (never deduped)", dups)
	}
}

func TestSaveKeyRejectedKeepsOldKey(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	*f.events = (*f.events)[:0]
	err := f.svc.SaveKey(ctx, "giphy", "bad-key-123")
	wantProviderErr(t, err, provider.KeyRejected, provider.GIPHY, 401)
	if v, _ := f.store.Get("giphy"); v != "uat-key-123" {
		t.Errorf("store = %q after a failed verify, want the old key", v)
	}
	if v, _ := f.svc.Keys.Get(provider.GIPHY); v != "uat-key-123" {
		t.Errorf("key cache = %q after a failed verify, want the old key", v)
	}
	if len(*f.events) != 0 {
		t.Errorf("events = %v after a failed verify, want none", *f.events)
	}
}

func TestSaveKeyOverwrites(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	for _, k := range []string{"uat-key-123", "uat-key-456"} {
		if err := f.svc.SaveKey(ctx, "giphy", k); err != nil {
			t.Fatal(err)
		}
	}
	if v, _ := f.store.Get("giphy"); v != "uat-key-456" {
		t.Errorf("store = %q, want the newer key", v)
	}
}

func TestSaveKeySanitizer(t *testing.T) {
	f := newSettings(t, "")
	for _, k := range []string{"short", strings.Repeat("k", 129), "has space", "abcd/efgh", "abcd?efgh", "abcd#efgh", "abcd%efgh", "abcd\x01efgh", "abcd efgh", ""} {
		err := f.svc.SaveKey(context.Background(), "giphy", k)
		wantProviderErr(t, err, provider.KeyRejected, provider.GIPHY, 0)
	}
	if n := f.hits.Load(); n != 0 {
		t.Errorf("sanitizer let %d requests through, want 0", n)
	}
	if err := f.svc.SaveKey(context.Background(), "giphy", strings.Repeat("k", 128)); err != nil {
		t.Errorf("128-char key: %v", err)
	}
}

func TestSaveKeyRateLimitedAndNetwork(t *testing.T) {
	f := newSettings(t, "")
	err := f.svc.SaveKey(context.Background(), "giphy", "ratelimit-key")
	wantProviderErr(t, err, provider.RateLimited, provider.GIPHY, 0)

	n := newSettings(t, "http://127.0.0.1:1")
	err = n.svc.SaveKey(context.Background(), "giphy", "uat-key-123")
	wantProviderErr(t, err, provider.Network, provider.GIPHY, 0)
	if _, ok := n.svc.Keys.Get(provider.GIPHY); ok {
		t.Error("key stored although verification failed")
	}
}

func TestSaveKeyUnsupportedProvider(t *testing.T) {
	f := newSettings(t, "")
	for _, p := range []string{"finergifs", "clips", "nope"} {
		err := f.svc.SaveKey(context.Background(), p, "uat-key-123")
		var pe *provider.Error
		if err == nil || errors.As(err, &pe) {
			t.Errorf("SaveKey(%s) = %v, want a plain error", p, err)
		}
	}
	if n := f.hits.Load(); n != 0 {
		t.Errorf("made %d requests, want 0", n)
	}
}

// slowProvider verifies "good" keys after a delay and rejects anything else; it tracks overlap.
type slowProvider struct {
	key     string
	active  *atomic.Int32
	overlap *atomic.Bool
}

func (s slowProvider) Search(ctx context.Context, q, c string) (provider.Page, error) {
	return s.Trending(ctx, c)
}

func (s slowProvider) Trending(ctx context.Context, cursor string) (provider.Page, error) {
	if s.active.Add(1) > 1 {
		s.overlap.Store(true)
	}
	defer s.active.Add(-1)
	if !strings.HasPrefix(s.key, "good") {
		return provider.Page{}, &provider.Error{Kind: provider.KeyRejected, Provider: provider.GIPHY, Status: 401}
	}
	time.Sleep(50 * time.Millisecond)
	return provider.Page{Items: []provider.Gif{}}, nil
}

func TestSaveKeyConcurrent(t *testing.T) {
	var active atomic.Int32
	var overlap atomic.Bool
	f := newSettings(t, "")
	f.svc.NewClient = func(p provider.ID, key string) provider.Provider {
		return slowProvider{key: key, active: &active, overlap: &overlap}
	}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, k := range []string{"good-key-123", "bad-key-1234"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = f.svc.SaveKey(context.Background(), "giphy", k)
		}()
	}
	wg.Wait()
	if errs[0] != nil || errs[1] == nil {
		t.Errorf("errs = %v, want good ok and bad rejected", errs)
	}
	if v, _ := f.store.Get("giphy"); v != "good-key-123" {
		t.Errorf("store = %q, want the good key", v)
	}
	if overlap.Load() {
		t.Error("two verifications ran at once; SaveKey is not serialized")
	}
}

func TestSaveKeyNeverLogsOrReturnsKey(t *testing.T) {
	logs := captureLog(t)
	f := newSettings(t, "")
	ctx := context.Background()
	var errs []error
	for _, tc := range []struct{ p, k string }{{"giphy", "bad-key-123"}, {"klipy", "bad-key-123"}, {"giphy", "uat-key-123"}, {"klipy", "uat-key-123"}, {"giphy", "abcd/efgh-uat-key-123"}} {
		errs = append(errs, f.svc.SaveKey(ctx, tc.p, tc.k))
	}
	n := newSettings(t, "http://127.0.0.1:1")
	errs = append(errs, n.svc.SaveKey(ctx, "klipy", "uat-key-123"))
	out := logs.String()
	if !strings.Contains(out, "key saved for giphy") || !strings.Contains(out, "key rejected for giphy") {
		t.Errorf("log lacks the save/reject lines: %q", out)
	}
	for _, secret := range []string{"uat-key-123", "bad-key-123"} {
		if strings.Contains(out, secret) {
			t.Errorf("log contains %s: %q", secret, out)
		}
		for _, err := range errs {
			if err != nil && (strings.Contains(err.Error(), secret) || strings.Contains(string(provider.MarshalError(err)), secret)) {
				t.Errorf("error leaks %s: %v", secret, err)
			}
		}
	}
}

// keychainDown fails every write, like a locked Keychain.
type keychainDown struct{ *settings.MemoryStore }

func (keychainDown) Set(string, string) error { return errors.New("keychain write failed") }
func (keychainDown) Delete(string) error      { return errors.New("keychain delete failed") }

func TestSaveKeyKeychainFailure(t *testing.T) {
	f := newSettings(t, "")
	f.svc.Keys = settings.NewKeys(keychainDown{settings.NewMemoryStore()})
	err := f.svc.SaveKey(context.Background(), "giphy", "uat-key-123")
	if err == nil || err.Error() != "couldn't save the key to the macOS Keychain" {
		t.Fatalf("err = %v, want the fixed Keychain failure text", err)
	}
	if len(*f.events) != 0 {
		t.Errorf("events = %v, want none", *f.events)
	}
}

func TestProviders(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	want := []ProviderStatus{
		{ID: "giphy", Configured: false, Clips: "no_key"},
		{ID: "clips", Configured: false, Clips: "no_key"},
		{ID: "klipy", Configured: false, Clips: "no_key"},
		{ID: "finergifs", Configured: true, Clips: "no_key"},
	}
	if got := f.gifs.Providers(ctx); !reflect.DeepEqual(got, want) {
		t.Errorf("no keys: %+v, want %+v", got, want)
	}
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	got := f.gifs.Providers(ctx)
	if len(got) != 4 {
		t.Fatalf("after GIPHY save: %+v, want 4 slots", got)
	}
	if !got[0].Configured || !got[1].Configured || got[2].Configured {
		t.Errorf("after GIPHY save: %+v", got)
	}
	if s := waitClips(t, f.clips); s != search.ClipsEnabled {
		t.Errorf("clips after GIPHY save settled at %s, want enabled", s)
	}
	if got := f.gifs.Providers(ctx); got[1].Clips != "enabled" {
		t.Errorf("Providers clips = %q, want enabled", got[1].Clips)
	}
	// The struct crossing the Wails boundary carries no key material.
	typ := reflect.TypeFor[ProviderStatus]()
	for i := range typ.NumField() {
		if k := typ.Field(i).Type.Kind(); k != reflect.String && k != reflect.Bool {
			t.Errorf("ProviderStatus.%s is %v", typ.Field(i).Name, k)
		}
	}
	if typ.NumField() != 3 {
		t.Errorf("ProviderStatus has %d fields, want id, configured, clips", typ.NumField())
	}
}

func TestLastProvider(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if got := f.svc.LastProvider(ctx); got != "" {
		t.Errorf("first run: %q, want \"\"", got)
	}
	if err := f.svc.SetLastProvider(ctx, "klipy"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(f.dir, "settings.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.svc.SetLastProvider(ctx, "bogus"); err == nil {
		t.Error("SetLastProvider(bogus) succeeded")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Errorf("settings.json changed on a rejected provider: %s", after)
	}
	if got := f.svc.LastProvider(ctx); got != "klipy" {
		t.Errorf("LastProvider = %q, want klipy", got)
	}
}

func TestLastProviderUnreadable(t *testing.T) {
	logs := captureLog(t)
	f := newSettings(t, "")
	if err := os.WriteFile(filepath.Join(f.dir, "settings.json"), []byte("{nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	f.svc.Store = settings.OpenStore(f.dir)
	if got := f.svc.LastProvider(context.Background()); got != "" {
		t.Errorf("unreadable file: %q, want \"\" (first run)", got)
	}
	if logs.Len() == 0 {
		t.Error("unreadable settings file was not logged")
	}
}

func TestSetPrefTheme(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	var applied []string
	f.svc.ApplyTheme = func(th string) { applied = append(applied, th) }
	snap, err := f.svc.SetPref(ctx, "theme", "dark")
	if err != nil || snap.Theme != "dark" {
		t.Fatalf("SetPref(theme, dark) = %+v, %v", snap, err)
	}
	if len(applied) != 1 || applied[0] != "dark" {
		t.Errorf("ApplyTheme calls = %v, want [dark]", applied)
	}
	if got := *f.events; len(got) != 1 || got[0] != "settings:changed" {
		t.Fatalf("events = %v, want [settings:changed]", got)
	}
	if got, ok := (*f.data)[0].(Snapshot); !ok || got != snap {
		t.Errorf("settings:changed payload = %#v, want the returned Snapshot %#v", (*f.data)[0], snap)
	}
	if got := f.svc.Snapshot(ctx); got.Theme != "dark" {
		t.Errorf("Snapshot = %+v, want theme dark", got)
	}
	b, err := os.ReadFile(filepath.Join(f.dir, "settings.json"))
	if err != nil || !strings.Contains(string(b), `"theme":"dark"`) {
		t.Errorf("settings.json = %s, %v", b, err)
	}
	// Equal value: no write, no event, no ApplyTheme.
	if err := os.Remove(filepath.Join(f.dir, "settings.json")); err != nil {
		t.Fatal(err)
	}
	snap, err = f.svc.SetPref(ctx, "theme", "dark")
	if err != nil || snap.Theme != "dark" {
		t.Errorf("SetPref(theme, dark) again = %+v, %v", snap, err)
	}
	if _, err := os.Stat(filepath.Join(f.dir, "settings.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("an unchanged value was written (stat err %v)", err)
	}
	if len(*f.events) != 1 || len(applied) != 1 {
		t.Errorf("unchanged value: events %v, ApplyTheme %v; want no new calls", *f.events, applied)
	}
}

func wantSettingsErr(t *testing.T, err error, kind string) {
	t.Helper()
	var se *SettingsError
	if !errors.As(err, &se) || se.Kind != kind {
		t.Fatalf("err = %#v, want *SettingsError %s", err, kind)
	}
}

func TestSetPrefRejectsInvalid(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	applied := 0
	f.svc.ApplyTheme = func(string) { applied++ }
	for _, tc := range [][2]string{{"theme", "neon"}, {"bogus", "x"}, {"lastProvider", "giphy"}} {
		_, err := f.svc.SetPref(ctx, tc[0], tc[1])
		wantSettingsErr(t, err, "invalid_value")
	}
	if _, err := os.Stat(filepath.Join(f.dir, "settings.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a refused value wrote settings.json (stat err %v)", err)
	}
	if len(*f.events) != 0 || applied != 0 {
		t.Errorf("refused values: events %v, ApplyTheme %d; want none", *f.events, applied)
	}
}

func TestSetPrefWriteFailure(t *testing.T) {
	logs := captureLog(t)
	f := newSettings(t, "")
	file := filepath.Join(f.dir, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	f.svc.Store = settings.OpenStore(file)
	_, err := f.svc.SetPref(context.Background(), "theme", "dark")
	wantSettingsErr(t, err, "prefs_write_failed")
	if len(*f.events) != 0 {
		t.Errorf("events = %v after a failed write, want none", *f.events)
	}
	if !strings.Contains(logs.String(), "settings write failed: theme") {
		t.Errorf("log = %q, want the failed field named", logs.String())
	}
}

// The overwrite bug: a tab switch used to write Prefs{LastProvider} wholesale and erase the theme.
func TestSetLastProviderKeepsTheme(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if _, err := f.svc.SetPref(ctx, "theme", "dark"); err != nil {
		t.Fatal(err)
	}
	*f.events = (*f.events)[:0]
	if err := f.svc.SetLastProvider(ctx, "klipy"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(f.dir, "settings.json"))
	if !strings.Contains(string(b), `"lastProvider":"klipy"`) || !strings.Contains(string(b), `"theme":"dark"`) {
		t.Errorf("settings.json = %s, want lastProvider klipy and theme dark", b)
	}
	if got := f.svc.LastProvider(ctx); got != "klipy" {
		t.Errorf("LastProvider = %q, want klipy", got)
	}
	if len(*f.events) != 0 {
		t.Errorf("a tab switch emitted %v, want nothing", *f.events)
	}
}

func TestSetPrefConcurrentWithLastProvider(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := range 40 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if i%2 == 0 {
				_ = f.svc.SetLastProvider(ctx, "klipy")
			} else {
				_, _ = f.svc.SetPref(ctx, "theme", []string{"light", "dark"}[i%4/2])
			}
		}()
	}
	wg.Wait()
	b, _ := os.ReadFile(filepath.Join(f.dir, "settings.json"))
	if !strings.Contains(string(b), `"lastProvider":"klipy"`) || !strings.Contains(string(b), `"theme":"`) {
		t.Errorf("settings.json = %s, want both fields", b)
	}
}

func TestMarshalSettingsError(t *testing.T) {
	if got := string(MarshalError(&SettingsError{Kind: "invalid_value"})); got != `{"kind":"invalid_value"}` {
		t.Errorf("MarshalError(SettingsError) = %s", got)
	}
	if got := string(MarshalError(ErrBusy)); got != `{"kind":"busy"}` {
		t.Errorf("MarshalError(ErrBusy) = %s", got)
	}
	if MarshalError(errors.New("plain")) != nil {
		t.Error("MarshalError(plain error) is not nil")
	}
}

// fakeAutostart is the in-memory login item: Enable/Disable flip enabled unless told to fail.
type fakeAutostart struct {
	mu                    sync.Mutex
	enabled               bool
	enableErr, disableErr error
	enables, disables     int
}

func (a *fakeAutostart) Enable() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enables++
	if a.enableErr != nil {
		return a.enableErr
	}
	a.enabled = true
	return nil
}

func (a *fakeAutostart) Disable() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.disables++
	if a.disableErr != nil {
		return a.disableErr
	}
	a.enabled = false
	return nil
}

func (a *fakeAutostart) IsEnabled() (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.enabled, nil
}

func TestSetLaunchAtLogin(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	as := f.svc.Autostart.(*fakeAutostart)
	snap, err := f.svc.SetLaunchAtLogin(ctx, true)
	if err != nil || !snap.LaunchAtLogin || as.enables != 1 || !f.svc.Store.Get().LaunchAtLogin {
		t.Fatalf("on: snap %+v, err %v, enables %d, pref %v", snap, err, as.enables, f.svc.Store.Get().LaunchAtLogin)
	}
	if got := *f.events; len(got) != 1 || got[0] != "settings:changed" {
		t.Errorf("events = %v, want [settings:changed]", got)
	}
	if got := f.svc.Snapshot(ctx); !got.LaunchAtLogin {
		t.Errorf("Snapshot = %+v, want launchAtLogin from the OS (true)", got)
	}
	// Idempotent: on again succeeds and the pref stays on.
	if snap, err := f.svc.SetLaunchAtLogin(ctx, true); err != nil || !snap.LaunchAtLogin || !f.svc.Store.Get().LaunchAtLogin {
		t.Errorf("on twice: %+v, %v", snap, err)
	}
	snap, err = f.svc.SetLaunchAtLogin(ctx, false)
	if err != nil || snap.LaunchAtLogin || as.disables != 1 || f.svc.Store.Get().LaunchAtLogin {
		t.Errorf("off: snap %+v, err %v, disables %d", snap, err, as.disables)
	}
}

// The checkbox shows the OS truth: a user who switched Gifsmith off in System Settings sees it off.
func TestSnapshotLaunchAtLoginIsOSState(t *testing.T) {
	f := newSettings(t, "")
	if _, err := f.svc.Store.Update(func(p *settings.Prefs) error { p.LaunchAtLogin = true; return nil }); err != nil {
		t.Fatal(err)
	}
	if f.svc.Snapshot(context.Background()).LaunchAtLogin {
		t.Error("Snapshot reports the pref, want the OS state (off)")
	}
}

func TestSetLaunchAtLoginNotInApplications(t *testing.T) {
	logs := captureLog(t)
	for _, exe := range []string{
		"/private/var/folders/x/AppTranslocation/abc/d/Gifsmith.app/Contents/MacOS/gifsmith",
		"/Volumes/Gifsmith/Gifsmith.app/Contents/MacOS/gifsmith",
	} {
		f := newSettings(t, "")
		f.svc.Executable = func() (string, error) { return exe, nil }
		_, err := f.svc.SetLaunchAtLogin(context.Background(), true)
		wantSettingsErr(t, err, "not_in_applications")
		if as := f.svc.Autostart.(*fakeAutostart); as.enables != 0 || f.svc.Store.Get().LaunchAtLogin || len(*f.events) != 0 {
			t.Errorf("%s: enables %d, pref %v, events %v; want nothing changed", exe, as.enables, f.svc.Store.Get().LaunchAtLogin, *f.events)
		}
	}
	if strings.Contains(logs.String(), "/Volumes/") || strings.Contains(logs.String(), "AppTranslocation/abc") {
		t.Errorf("log carries the path: %q", logs.String())
	}
}

func TestSetLaunchAtLoginFailures(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	as := f.svc.Autostart.(*fakeAutostart)
	as.enableErr = errors.New("smappservice: /Applications/Gifsmith.app denied")
	_, err := f.svc.SetLaunchAtLogin(ctx, true)
	wantSettingsErr(t, err, "autostart_failed")
	if f.svc.Store.Get().LaunchAtLogin || len(*f.events) != 0 {
		t.Errorf("failed Enable: pref %v, events %v; want unchanged", f.svc.Store.Get().LaunchAtLogin, *f.events)
	}
	as.disableErr = errors.New("boom")
	_, err = f.svc.SetLaunchAtLogin(ctx, false)
	wantSettingsErr(t, err, "autostart_failed")
}

func TestApplyStartup(t *testing.T) {
	for _, tc := range []struct {
		name          string
		pref, enabled bool
		wantEnables   int
	}{
		{"pref on, OS off", true, false, 1},
		{"pref on, OS on", true, true, 0},
		{"pref off, OS on", false, true, 0}, // never disables on launch
		{"pref off, OS off", false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newSettings(t, "")
			as := f.svc.Autostart.(*fakeAutostart)
			as.enabled = tc.enabled
			if _, err := f.svc.Store.Update(func(p *settings.Prefs) error { p.LaunchAtLogin = tc.pref; return nil }); err != nil {
				t.Fatal(err)
			}
			f.svc.ApplyStartup()
			if as.enables != tc.wantEnables || as.disables != 0 {
				t.Errorf("enables %d, disables %d; want %d, 0", as.enables, as.disables, tc.wantEnables)
			}
		})
	}
}

func TestAccessibilityStatus(t *testing.T) {
	f := newSettings(t, "")
	fake := &platform.Fake{}
	f.svc.Platform = fake
	if f.svc.AccessibilityStatus(context.Background()) {
		t.Error("AccessibilityStatus = true with an untrusted Fake")
	}
	fake.Trusted = true
	if !f.svc.AccessibilityStatus(context.Background()) {
		t.Error("AccessibilityStatus = false with a trusted Fake")
	}
	for _, c := range fake.Calls {
		if c != "AccessibilityTrusted:false" {
			t.Errorf("call %q, want only AccessibilityTrusted:false (never prompts)", c)
		}
	}
}

// waitClips waits for an async Clips probe to settle.
func waitClips(t *testing.T, c *search.Clips) search.ClipsState {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if s := c.State(); s != search.ClipsChecking && s != search.ClipsUnknown {
			return s
		}
		time.Sleep(5 * time.Millisecond)
	}
	return c.State()
}

func TestProvidersClipsChecking(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	release := make(chan struct{})
	probe := f.clips.Probe
	f.clips.Probe = func(ctx context.Context) error { <-release; return probe(ctx) }
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if got := f.gifs.Providers(ctx); got[1].Clips != "checking" {
		t.Errorf("during a slow probe: clips = %q, want checking", got[1].Clips)
	}
	close(release)
	if s := waitClips(t, f.clips); s != search.ClipsEnabled {
		t.Errorf("after the probe: %s, want enabled", s)
	}
}

func TestSaveKeyTriggersClips(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "bad-key-123"); err == nil {
		t.Fatal("bad key saved")
	}
	time.Sleep(50 * time.Millisecond)
	if n := f.probes.Load(); n != 0 {
		t.Errorf("failed SaveKey probed Clips %d times, want 0", n)
	}
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if s := waitClips(t, f.clips); s != search.ClipsEnabled || f.probes.Load() != 1 {
		t.Errorf("after save: %s, probes %d; want enabled, 1", s, f.probes.Load())
	}
	// A new GIPHY key resets the result and probes again.
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-456"); err != nil {
		t.Fatal(err)
	}
	if s := waitClips(t, f.clips); s != search.ClipsEnabled || f.probes.Load() != 2 {
		t.Errorf("after a second save: %s, probes %d; want enabled, 2", s, f.probes.Load())
	}
	// A Klipy save leaves Clips alone.
	if err := f.svc.SaveKey(ctx, "klipy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if n := f.probes.Load(); n != 2 {
		t.Errorf("Klipy save probed Clips (probes %d, want 2)", n)
	}
}

func TestGiphySuccessTriggersClipsOnce(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.store.Set("giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	f.svc.Keys.Load()
	if _, err := f.gifs.Search(ctx, "giphy", "boom", ""); err == nil {
		t.Fatal("boom succeeded")
	}
	time.Sleep(50 * time.Millisecond)
	if n := f.probes.Load(); n != 0 {
		t.Errorf("a failed GIPHY call probed Clips (%d)", n)
	}
	for range 2 {
		if _, err := f.gifs.Trending(ctx, "giphy", ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.gifs.Search(ctx, "giphy", "cats", ""); err != nil {
		t.Fatal(err)
	}
	if s := waitClips(t, f.clips); s != search.ClipsEnabled || f.probes.Load() != 1 {
		t.Errorf("after GIPHY successes: %s, probes %d; want enabled, 1", s, f.probes.Load())
	}
}

func TestClipsRefusedUnlessEnabled(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	for _, call := range []func() (provider.Page, error){
		func() (provider.Page, error) { return f.gifs.Search(ctx, "clips", "cat", "") },
		func() (provider.Page, error) { return f.gifs.Trending(ctx, "clips", "") },
	} {
		p, err := call()
		wantProviderErr(t, err, provider.ClipsDisabled, provider.Clips, 0)
		if len(p.Items) != 0 {
			t.Errorf("refused call returned %d items", len(p.Items))
		}
	}
	if n := f.hits.Load(); n != 0 {
		t.Errorf("refused Clips calls made %d fixture requests", n)
	}
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	waitClips(t, f.clips)
	p, err := f.gifs.Search(ctx, "clips", "cat", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) == 0 || !p.Items[0].IsClip || p.Items[0].Provider != provider.Clips || !strings.HasSuffix(p.Items[0].DownloadURL, "-720p.mp4") {
		t.Errorf("clips search = %+v, want clip items downloading the 720p MP4", p.Items)
	}
}

func TestClipsDeniedFixture(t *testing.T) {
	h := fakeapi.Handler(fakeapi.Options{ClipsDenied: true})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	f := newSettings(t, srv.URL)
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if s := waitClips(t, f.clips); s != search.ClipsDisabled {
		t.Errorf("denied: %s, want disabled", s)
	}
	if s, err := f.gifs.CheckClips(ctx); s != "disabled" || err != nil {
		t.Errorf("CheckClips when disabled = %q, %v; want disabled, nil", s, err)
	}
	_, err := f.gifs.Trending(ctx, "clips", "")
	wantProviderErr(t, err, provider.ClipsDisabled, provider.Clips, 0)
}

func TestCheckClipsUnknown(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.store.Set("giphy", "ratelimit-key"); err != nil {
		t.Fatal(err)
	}
	f.svc.Keys.Load()
	s, err := f.gifs.CheckClips(ctx)
	wantProviderErr(t, err, provider.RateLimited, provider.Clips, 0)
	if s != "unknown" || f.probes.Load() != 1 {
		t.Errorf("CheckClips = %q, probes %d; want unknown, 1", s, f.probes.Load())
	}
	// The probe's 429 blocks the shared GIPHY budget: no second probe into it.
	s, err = f.gifs.CheckClips(ctx)
	wantProviderErr(t, err, provider.RateLimited, provider.Clips, 0)
	if s != "unknown" || f.probes.Load() != 1 || !f.svc.Svc.Blocked(provider.GIPHY) {
		t.Errorf("CheckClips while blocked = %q, probes %d; want unknown, 1, GIPHY blocked", s, f.probes.Load())
	}
	if err := f.store.Set("giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	f.svc.Keys.Load()
	f.svc.Svc.Invalidate(provider.GIPHY) // what SaveKey does for a new key
	f.clips.Reset()
	if s, err := f.gifs.CheckClips(ctx); s != "enabled" || err != nil || f.probes.Load() != 2 {
		t.Errorf("CheckClips retry = %q, %v, probes %d; want enabled, nil, 2", s, err, f.probes.Load())
	}
}

// SRCH-09 tracer: a rating change is stored, flushes GIPHY's cached pages and reaches the next request.
func TestSetPrefGiphyRating(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil { // seeds GIPHY trending
		t.Fatal(err)
	}
	*f.events = (*f.events)[:0]
	*f.data = (*f.data)[:0]
	snap, err := f.svc.SetPref(ctx, "giphyRating", "g")
	if err != nil || snap.GiphyRating != "g" {
		t.Fatalf("SetPref(giphyRating, g) = %+v, %v", snap, err)
	}
	if got := f.svc.Store.Get().GiphyRating; got != "g" {
		t.Errorf("stored rating = %q, want g", got)
	}
	if got := *f.events; len(got) != 1 || got[0] != "settings:changed" {
		t.Fatalf("events = %v, want [settings:changed]", got)
	}
	if got, ok := (*f.data)[0].(Snapshot); !ok || got.GiphyRating != "g" {
		t.Errorf("payload = %#v, want giphyRating g", (*f.data)[0])
	}
	if p := f.svc.Params(); p.Rating != "g" {
		t.Errorf("Params().Rating = %q, want g", p.Rating)
	}
	before := f.hits.Load()
	p, err := f.gifs.Trending(ctx, "giphy", "")
	if err != nil {
		t.Fatal(err)
	}
	if n := f.hits.Load() - before; n != 1 {
		t.Errorf("Trending after a rating change made %d requests, want 1 (cache flushed)", n)
	}
	if !strings.HasPrefix(p.Items[0].Title, "[rating=g] ") {
		t.Errorf("first title %q: the request did not carry rating=g", p.Items[0].Title)
	}
	// The same value again: no write, no flush, no event.
	if err := os.Remove(filepath.Join(f.dir, "settings.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.SetPref(ctx, "giphyRating", "g"); err != nil {
		t.Fatal(err)
	}
	before = f.hits.Load()
	_, _ = f.gifs.Trending(ctx, "giphy", "")
	if n := f.hits.Load() - before; n != 0 || len(*f.events) != 1 {
		t.Errorf("unchanged rating: %d requests, events %v; want cached, no event", n, *f.events)
	}
	if _, err := os.Stat(filepath.Join(f.dir, "settings.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("an unchanged rating was written (stat err %v)", err)
	}
}

func TestSetPrefGiphyRatingInvalid(t *testing.T) {
	f := newSettings(t, "")
	for _, v := range []string{"", "nc-17", "PG"} {
		_, err := f.svc.SetPref(context.Background(), "giphyRating", v)
		wantSettingsErr(t, err, "invalid_value")
	}
	if len(*f.events) != 0 {
		t.Errorf("events = %v, want none", *f.events)
	}
}

func TestParamsDefaultRating(t *testing.T) {
	f := newSettings(t, "")
	if p := f.svc.Params(); p.Rating != "pg-13" || p.PageSize != 24 {
		t.Errorf("Params() = %+v, want rating pg-13 and page size 24", p)
	}
	if s := f.svc.Snapshot(context.Background()); s.GiphyRating != "pg-13" {
		t.Errorf("Snapshot.GiphyRating = %q, want pg-13 (resolved)", s.GiphyRating)
	}
}

func TestSetPrefPageSize(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if err := f.svc.SaveKey(ctx, "klipy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.gifs.Search(ctx, "finergifs", "dwight", ""); err != nil {
		t.Fatal(err)
	}
	snap, err := f.svc.SetPref(ctx, "pageSize", "36")
	if err != nil || snap.PageSize != 36 || f.svc.Params().PageSize != 36 {
		t.Fatalf("SetPref(pageSize, 36) = %+v, %v", snap, err)
	}
	for _, c := range []struct {
		p    string
		call func() (provider.Page, error)
	}{
		{"giphy", func() (provider.Page, error) { return f.gifs.Trending(ctx, "giphy", "") }},
		{"klipy", func() (provider.Page, error) { return f.gifs.Trending(ctx, "klipy", "") }},
		{"finergifs", func() (provider.Page, error) { return f.gifs.Search(ctx, "finergifs", "dwight", "") }},
	} {
		before := f.hits.Load()
		pg, err := c.call()
		if err != nil || f.hits.Load()-before != 1 || len(pg.Items) != 36 {
			t.Errorf("%s after pageSize 36: %d items, %d requests, %v; want 36 fresh items", c.p, len(pg.Items), f.hits.Load()-before, err)
		}
	}
	for _, v := range []string{"25", "0", "23", "49", "50", "abc", "", "036"} {
		_, err := f.svc.SetPref(ctx, "pageSize", v)
		wantSettingsErr(t, err, "invalid_value")
	}
	for _, v := range []string{"24", "48"} {
		if snap, err := f.svc.SetPref(ctx, "pageSize", v); err != nil || strconv.Itoa(snap.PageSize) != v {
			t.Errorf("SetPref(pageSize, %s) = %+v, %v", v, snap, err)
		}
	}
}

func TestSetPrefLangCountry(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	for _, p := range []string{"giphy", "klipy"} {
		if err := f.svc.SaveKey(ctx, p, "uat-key-123"); err != nil {
			t.Fatal(err)
		}
	}
	cached := func(p string) bool {
		before := f.hits.Load()
		_, _ = f.gifs.Trending(ctx, p, "")
		return f.hits.Load() == before
	}
	if snap, err := f.svc.SetPref(ctx, "giphyLang", "de"); err != nil || snap.GiphyLang != "de" || f.svc.Params().Lang != "de" {
		t.Fatalf("SetPref(giphyLang, de) = %+v, %v", snap, err)
	}
	if cached("giphy") || !cached("klipy") {
		t.Error("giphyLang must flush GIPHY only")
	}
	if snap, err := f.svc.SetPref(ctx, "klipyCountry", "de"); err != nil || snap.KlipyCountry != "de" || f.svc.Params().Country != "de" {
		t.Fatalf("SetPref(klipyCountry, de) = %+v, %v", snap, err)
	}
	if !cached("giphy") || cached("klipy") {
		t.Error("klipyCountry must flush Klipy only")
	}
	pg, err := f.gifs.Search(ctx, "giphy", "cat", "")
	if err != nil || !strings.HasPrefix(pg.Items[0].Title, "[lang=de] ") {
		t.Errorf("GIPHY search title %q, %v; want the lang=de tag", pg.Items[0].Title, err)
	}
	for _, tc := range [][2]string{{"giphyLang", "xx"}, {"giphyLang", "DE"}, {"klipyCountry", "zz"}, {"klipyCountry", "US"}} {
		_, err := f.svc.SetPref(ctx, tc[0], tc[1])
		wantSettingsErr(t, err, "invalid_value")
	}
	for _, name := range []string{"giphyLang", "klipyCountry"} {
		if snap, err := f.svc.SetPref(ctx, name, ""); err != nil || snap.GiphyLang+snap.KlipyCountry != map[string]string{"giphyLang": "de", "klipyCountry": ""}[name] {
			t.Errorf("SetPref(%s, \"\") = %+v, %v; want Match System", name, snap, err)
		}
	}
}

func TestParamsSystemLocale(t *testing.T) {
	f := newSettings(t, "")
	fake := &platform.Fake{}
	f.svc.Platform = fake
	for _, tc := range []struct{ lang, region, wantLang, wantCountry string }{
		{"pt-BR", "BR", "pt", "br"},
		{"zh-Hant-TW", "TW", "zh-TW", "tw"},
		{"zh-Hans-CN", "CN", "zh-CN", "cn"},
		{"zh-HK", "HK", "zh-TW", "hk"},
		{"nb-NO", "NO", "no", "no"},
		{"fil-PH", "PH", "tl", "ph"},
		{"iw-IL", "IL", "he", "il"},
		{"de-DE", "DE", "de", "de"},
		{"xx-YY", "ZZ", "en", "us"},
		{"", "", "en", "us"},
	} {
		fake.Lang, fake.Region = tc.lang, tc.region
		p := f.svc.Params()
		if p.Lang != tc.wantLang || p.Country != tc.wantCountry {
			t.Errorf("system %q/%q: Params lang %q country %q; want %q %q", tc.lang, tc.region, p.Lang, p.Country, tc.wantLang, tc.wantCountry)
		}
	}
	fake.Lang, fake.Region = "de-DE", "DE"
	c := f.svc.Choices(context.Background())
	if strings.Join(c.Ratings, " ") != "g pg pg-13 r" || fmt.Sprint(c.PageSizes) != "[24 36 48]" ||
		!slices.Equal(c.Languages, giphy.Languages) || !slices.Equal(c.Countries, klipy.Countries) ||
		c.SystemLanguage != "de" || c.SystemCountry != "de" {
		t.Errorf("Choices = %+v", c)
	}
	if s := f.svc.Snapshot(context.Background()); s.PageSize != 24 || s.GiphyLang != "" || s.KlipyCountry != "" {
		t.Errorf("Snapshot = %+v, want pageSize 24 and Match System", s)
	}
}

func TestKeyStatus(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	want := []KeyStatus{{"giphy", "saved", "-123"}, {"klipy", "not_set", ""}}
	if got := f.svc.KeyStatus(ctx); !reflect.DeepEqual(got, want) {
		t.Errorf("KeyStatus = %+v, want %+v", got, want)
	}
	f.svc.Keys.MarkRejected(provider.GIPHY)
	if got := f.svc.KeyStatus(ctx); got[0] != (KeyStatus{"giphy", "rejected", "-123"}) {
		t.Errorf("after MarkRejected: %+v", got[0])
	}
}

func TestTestKeyUncached(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	f.svc.Keys.MarkRejected(provider.GIPHY)
	before := f.hits.Load()
	for range 2 {
		if err := f.svc.TestKey(ctx, "giphy"); err != nil {
			t.Fatal(err)
		}
	}
	if n := f.hits.Load() - before; n != 2 {
		t.Errorf("two Tests made %d requests, want 2 (never from the trending cache)", n)
	}
	if f.svc.Keys.Rejected(provider.GIPHY) {
		t.Error("a passing Test did not clear Rejected")
	}
	before = f.hits.Load()
	if _, err := f.gifs.Trending(ctx, "giphy", ""); err != nil || f.hits.Load() != before {
		t.Errorf("Trending after Test: %v, %d requests; want the seeded page", err, f.hits.Load()-before)
	}
}

func TestTestKeyRejected(t *testing.T) {
	logs := captureLog(t)
	f := newSettings(t, "")
	if err := f.svc.Keys.Set(provider.GIPHY, "bad-key-123"); err != nil { // stored, then revoked at GIPHY
		t.Fatal(err)
	}
	err := f.svc.TestKey(context.Background(), "giphy")
	wantProviderErr(t, err, provider.KeyRejected, provider.GIPHY, 401)
	if !f.svc.Keys.Rejected(provider.GIPHY) {
		t.Error("key_rejected did not mark GIPHY rejected")
	}
	if got := *f.events; len(got) != 1 || got[0] != "providers:changed" {
		t.Errorf("events = %v, want [providers:changed]", got)
	}
	if strings.Contains(logs.String(), "bad-key-123") || strings.Contains(logs.String(), "http") && strings.Contains(logs.String(), "api_key") {
		t.Errorf("log leaks key material or a URL: %q", logs.String())
	}
}

func TestTestKeyInsideBudget(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.Keys.Set(provider.GIPHY, "ratelimit-key-1"); err != nil { // fixture answers 429
		t.Fatal(err)
	}
	err := f.svc.TestKey(ctx, "giphy")
	wantProviderErr(t, err, provider.RateLimited, provider.GIPHY, 429)
	before := f.hits.Load()
	err = f.svc.TestKey(ctx, "giphy")
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != provider.RateLimited || f.hits.Load() != before {
		t.Errorf("Test while blocked: %v, %d requests; want rate_limited and no request", err, f.hits.Load()-before)
	}
}

func TestTestKeyRefusals(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.TestKey(ctx, "finergifs"); !errors.Is(err, errNoKeyProvider) {
		t.Errorf("finergifs: %v, want errNoKeyProvider", err)
	}
	if err := f.svc.TestKey(ctx, "bogus"); !errors.Is(err, errUnknownProvider) {
		t.Errorf("bogus: %v, want errUnknownProvider", err)
	}
	wantProviderErr(t, f.svc.TestKey(ctx, "giphy"), provider.KeyMissing, provider.GIPHY, 0)
	if f.hits.Load() != 0 {
		t.Errorf("refusals made %d requests", f.hits.Load())
	}
}

func TestRemoveKey(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	waitClips(t, f.clips)
	*f.events = (*f.events)[:0]
	if err := f.svc.RemoveKey(ctx, "giphy"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.store.Get("giphy"); !errors.Is(err, settings.ErrNoKey) {
		t.Errorf("Keychain item still there: %v", err)
	}
	if _, ok := f.svc.Keys.Get(provider.GIPHY); ok {
		t.Error("Keys still has the GIPHY key")
	}
	if _, err := f.gifs.Trending(ctx, "giphy", ""); err == nil {
		t.Error("the seeded GIPHY page survived RemoveKey (no Invalidate)")
	}
	if s := f.clips.State(); s != search.ClipsNoKey {
		t.Errorf("Clips state %s, want no_key", s)
	}
	if got := *f.events; len(got) != 1 || got[0] != "providers:changed" {
		t.Errorf("events = %v, want [providers:changed]", got)
	}
	if err := f.svc.RemoveKey(ctx, "giphy"); err != nil {
		t.Errorf("second RemoveKey = %v, want nil", err)
	}
	if err := f.svc.RemoveKey(ctx, "finergifs"); !errors.Is(err, errNoKeyProvider) {
		t.Errorf("finergifs: %v", err)
	}
}

func TestRemoveKeyKeychainFailure(t *testing.T) {
	logs := captureLog(t)
	f := newSettings(t, "")
	ms := settings.NewMemoryStore()
	_ = ms.Set("klipy", "uat-key-123")
	f.svc.Keys = settings.NewKeys(keychainDown{ms})
	f.svc.Keys.Load()
	err := f.svc.RemoveKey(context.Background(), "klipy")
	wantSettingsErr(t, err, "keychain_delete_failed")
	if _, ok := f.svc.Keys.Get(provider.Klipy); !ok {
		t.Error("the key went from the cache although the Keychain delete failed")
	}
	if !strings.Contains(logs.String(), "key not removed for klipy: keychain delete failed") || len(*f.events) != 0 {
		t.Errorf("log %q, events %v", logs.String(), *f.events)
	}
}

// A key_rejected from search (GIFs or Clips) marks the key rejected, so the panel and Settings agree.
func TestSearchKeyRejectedMarks(t *testing.T) {
	f := newSettings(t, "")
	ctx := context.Background()
	if err := f.svc.SaveKey(ctx, "giphy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if s := waitClips(t, f.clips); s != search.ClipsEnabled {
		t.Fatalf("clips %s, want enabled", s)
	}
	if _, err := f.gifs.Search(ctx, "clips", "revoked", ""); err == nil {
		t.Fatal("clips search with revoked succeeded")
	}
	if !f.svc.Keys.Rejected(provider.GIPHY) {
		t.Error("a Clips key_rejected did not mark GIPHY rejected")
	}
	if err := f.svc.SaveKey(ctx, "klipy", "uat-key-123"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.gifs.Search(ctx, "klipy", "revoked", ""); err == nil || !f.svc.Keys.Rejected(provider.Klipy) {
		t.Errorf("klipy revoked: %v, rejected %v", err, f.svc.Keys.Rejected(provider.Klipy))
	}
}
