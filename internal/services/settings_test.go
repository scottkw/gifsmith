package services

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
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
	store := settings.NewMemoryStore()
	keys := settings.NewKeys(store)
	keys.Load()
	newClient := func(p provider.ID, key string) provider.Provider {
		k := func() (string, bool) { return key, true }
		if p == provider.Klipy {
			return &klipy.Client{BaseURL: base + "/klipy", Key: k, HTTP: hc}
		}
		return &giphy.Client{BaseURL: base + "/giphy", Key: k, HTTP: hc}
	}
	clipsClient := &giphy.Client{BaseURL: base + "/giphy", Key: keys.Func(provider.GIPHY), HTTP: hc, Clips: true}
	svc := search.New(map[provider.ID]provider.Provider{
		provider.GIPHY: &giphy.Client{BaseURL: base + "/giphy", Key: keys.Func(provider.GIPHY), HTTP: hc},
		provider.Clips: clipsClient,
		provider.Klipy: &klipy.Client{BaseURL: base + "/klipy", Key: keys.Func(provider.Klipy), HTTP: hc},
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
	s := &SettingsService{Keys: keys, Dir: t.TempDir(), Svc: svc, NewClient: newClient, Emit: func(n string) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, n)
	}, Clips: clips}
	return settingsFixture{svc: s, gifs: &GifService{Svc: svc, Keys: keys, Clips: clips}, store: store, hits: &hits, events: &events, clips: clips, probes: &probes}
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
	path := filepath.Join(f.svc.Dir, "settings.json")
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
	if err := os.WriteFile(filepath.Join(f.svc.Dir, "settings.json"), []byte("{nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := f.svc.LastProvider(context.Background()); got != "" {
		t.Errorf("unreadable file: %q, want \"\" (first run)", got)
	}
	if logs.Len() == 0 {
		t.Error("unreadable settings file was not logged")
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
	if len(p.Items) == 0 || !p.Items[0].IsClip || p.Items[0].Provider != provider.Clips || p.Items[0].DownloadURL != "" {
		t.Errorf("clips search = %+v, want clip items with no download URL", p.Items)
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
