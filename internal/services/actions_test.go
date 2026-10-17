package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gifsmith/internal/platform"
	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
	"gifsmith/internal/staging"
)

// staged stands in for the per-run staging path in call lists.
const staged = "CopyFile:staged"

type fixture struct {
	gif  provider.Gif
	dir  string
	hits *atomic.Int32
}

func newService(t *testing.T, f *platform.Fake) (*ActionService, fixture) {
	t.Helper()
	var hits atomic.Int32
	h := fakeapi.Handler(fakeapi.Options{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	dir := t.TempDir()
	st := &staging.Store{Dir: dir, HTTP: srv.Client(), Allow: func(_ provider.ID, u *url.URL) bool { return u.Host == host }}
	g := provider.Gif{Provider: provider.FinerGifs, ID: "office-1", Title: "Office moment 1", DownloadURL: srv.URL + "/media/office-1.gif"}
	return &ActionService{Platform: f, Staging: st, Hide: func() { f.Note("hide") }}, fixture{gif: g, dir: dir, hits: &hits}
}

// calls returns f.Calls with a CopyFile of a .gif staged under dir replaced by staged.
func (fx fixture) calls(f *platform.Fake) []string {
	out := slices.Clone(f.Calls)
	for i, c := range out {
		if p, ok := strings.CutPrefix(c, "CopyFile:"); ok && strings.HasPrefix(p, fx.dir+string(filepath.Separator)) && strings.HasSuffix(p, ".gif") {
			out[i] = staged
		}
	}
	return out
}

func TestCopyStagesAndDoesNotHide(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	stray := filepath.Join(fx.dir, "0123456789abcdef")
	if err := os.MkdirAll(stray, 0o755); err != nil { // an older staged copy
		t.Fatal(err)
	}
	r, err := s.Copy(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Message != "Copied — press ⌘V to paste" {
		t.Errorf("result = %+v, want Copied with the UI-SPEC message", r)
	}
	// The frontend hides 400 ms after showing the confirmation; Go must not.
	if want := []string{staged}; !slices.Equal(fx.calls(f), want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
	if _, err := os.Stat(stray); !os.IsNotExist(err) {
		t.Errorf("older staged copy not swept after a successful copy (err %v)", err)
	}
}

func TestCopyErrorKeepsPanelOpen(t *testing.T) {
	f := &platform.Fake{CopyErr: errors.New("pasteboard write failed")}
	s, fx := newService(t, f)
	stray := filepath.Join(fx.dir, "0123456789abcdef")
	if err := os.MkdirAll(stray, 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := s.Copy(context.Background(), fx.gif)
	if !errors.Is(err, f.CopyErr) {
		t.Errorf("err = %v, want %v", err, f.CopyErr)
	}
	if r.Copied {
		t.Error("Copied = true on error")
	}
	if slices.Contains(f.Calls, "hide") {
		t.Errorf("panel hidden on error: calls = %q", f.Calls)
	}
	if _, err := os.Stat(stray); err != nil {
		t.Errorf("previous staged copy removed although the pasteboard write failed: %v", err)
	}
}

func TestCopyClipRefused(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	g := fx.gif
	g.IsClip = true
	if _, err := s.Copy(context.Background(), g); !errors.Is(err, ErrClipsCopy) {
		t.Errorf("Copy err = %v, want ErrClipsCopy", err)
	}
	if _, err := s.Paste(context.Background(), g); !errors.Is(err, ErrClipsCopy) {
		t.Errorf("Paste err = %v, want ErrClipsCopy", err)
	}
	if n := fx.hits.Load(); n != 0 || len(f.Calls) != 0 {
		t.Errorf("Clips copy made %d downloads and calls %q, want none", n, f.Calls)
	}
}

func TestCopyDisallowedURL(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	g := fx.gif
	g.DownloadURL = "https://evil.example.test/x.gif"
	_, err := s.Copy(context.Background(), g)
	if !errors.Is(err, staging.ErrNotAllowed) || strings.Contains(err.Error(), "evil.example") {
		t.Errorf("err = %v, want ErrNotAllowed without the URL", err)
	}
	if n := fx.hits.Load(); n != 0 || len(f.Calls) != 0 {
		t.Errorf("disallowed URL made %d requests and calls %q, want none", n, f.Calls)
	}
}

func TestCopyTwiceKeepsOnlyLatest(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	second := fx.gif
	second.ID, second.Title = "office-2", "Office moment 2"
	second.DownloadURL = strings.Replace(second.DownloadURL, "office-1", "office-2", 1)
	for _, g := range []provider.Gif{fx.gif, second} {
		if _, err := s.Copy(context.Background(), g); err != nil {
			t.Fatal(err)
		}
	}
	if want := []string{staged, staged}; !slices.Equal(fx.calls(f), want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
	last := strings.TrimPrefix(f.Calls[1], "CopyFile:")
	ents, _ := os.ReadDir(fx.dir)
	if len(ents) != 1 || ents[0].Name() != filepath.Base(filepath.Dir(last)) {
		t.Errorf("staging dir holds %d entries, want only the latest copy", len(ents))
	}
}

var prevApp = platform.App{BundleID: "com.apple.TextEdit", Name: "TextEdit", PID: 42}

func newPasteService(t *testing.T, f *platform.Fake) (*ActionService, fixture) {
	t.Helper()
	s, fx := newService(t, f)
	s.Previous = func() platform.App { return prevApp }
	s.Sleep = func(d time.Duration) { f.Note("sleep:" + d.String()) }
	return s, fx
}

func TestPasteWithoutAccessibility(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || !r.NeedsAccessibility || r.Pasted || r.Message == "" {
		t.Errorf("result = %+v, want Copied+NeedsAccessibility with a message", r)
	}
	// Point at the button, not a Settings path (macOS 27 has no Privacy & Security › Accessibility).
	if !strings.Contains(r.Message, "Enable automatic paste…") || strings.Contains(r.Message, "System Settings") {
		t.Errorf("message = %q, want it to name the Enable automatic paste… button and no Settings path", r.Message)
	}
	// App.svelte shows "Enable automatic paste…" off this exact JSON key (models.ts Result).
	if b, _ := json.Marshal(r); !strings.Contains(string(b), `"needsAccessibility":true`) {
		t.Errorf("JSON = %s, want \"needsAccessibility\":true", b)
	}
	if want := []string{staged, "AccessibilityTrusted:false"}; !slices.Equal(fx.calls(f), want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestPasteTrustedOrder(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || !r.Pasted {
		t.Errorf("result = %+v, want Copied+Pasted", r)
	}
	want := []string{staged, "AccessibilityTrusted:false", "hide", "sleep:80ms", "FrontmostApp", "PasteShortcut"}
	if !slices.Equal(fx.calls(f), want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestPasteReactivatesPrevious(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: platform.App{BundleID: "com.other", PID: 7}, ActivateSetsFrontmost: true}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Pasted {
		t.Errorf("result = %+v, want Pasted", r)
	}
	want := []string{staged, "AccessibilityTrusted:false", "hide", "sleep:80ms", "FrontmostApp", "ActivateApp:42", "sleep:80ms", "FrontmostApp", "PasteShortcut"}
	if !slices.Equal(fx.calls(f), want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

// macOS 14+ often declines activation from a background app; ⌘V must not land in the wrong app.
func TestPasteDeclinedReactivationCopiesOnly(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: platform.App{BundleID: "com.other", PID: 7}}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Pasted || !strings.Contains(r.Message, "TextEdit") {
		t.Errorf("result = %+v, want Copied, not Pasted, message naming TextEdit", r)
	}
	if slices.Contains(f.Calls, "PasteShortcut") {
		t.Errorf("⌘V posted after declined activation: calls = %q", f.Calls)
	}
}

// An unknown target (Shell clears prev when its lookup fails) must not get a blind ⌘V.
func TestPasteUnknownTargetCopiesOnly(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, fx := newPasteService(t, f)
	s.Previous = func() platform.App { return platform.App{} }
	r, err := s.Paste(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Pasted || r.Message == "" {
		t.Errorf("result = %+v, want Copied, not Pasted, with a message", r)
	}
	if slices.Contains(f.Calls, "PasteShortcut") {
		t.Errorf("⌘V posted with no remembered target: calls = %q", f.Calls)
	}
}

func TestPasteFrontmostErrorCopiesOnly(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp, FrontmostErr: errors.New("no frontmost app")}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Pasted || r.Message == "" {
		t.Errorf("result = %+v, want Copied, not Pasted, with a message", r)
	}
	if slices.Contains(f.Calls, "PasteShortcut") || slices.Contains(f.Calls, "ActivateApp:42") {
		t.Errorf("acted on a failed frontmost lookup: calls = %q", f.Calls)
	}
}

func TestPasteCopyError(t *testing.T) {
	f := &platform.Fake{Trusted: true, CopyErr: errors.New("pasteboard write failed")}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), fx.gif)
	if !errors.Is(err, f.CopyErr) {
		t.Errorf("err = %v, want %v", err, f.CopyErr)
	}
	if r.Copied || r.Pasted {
		t.Errorf("result = %+v on copy error", r)
	}
	if want := []string{staged}; !slices.Equal(fx.calls(f), want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestPastePostError(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp, PasteErr: errors.New("could not post")}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), fx.gif)
	if err != nil {
		t.Errorf("err = %v, want nil (Wails drops the result on error and the panel is hidden)", err)
	}
	if !r.Copied || r.Pasted || r.Message == "" {
		t.Errorf("result = %+v, want Copied, not Pasted, with a message", r)
	}
}

func TestRequestAccessibilityOpensSettings(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newPasteService(t, f)
	ok, err := s.RequestAccessibility(context.Background())
	if err != nil || ok {
		t.Errorf("got (%v, %v), want (false, nil)", ok, err)
	}
	if want := []string{"AccessibilityTrusted:true", "OpenAccessibilitySettings"}; !slices.Equal(fx.calls(f), want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestActionsRefuseWhileBusy(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, fx := newPasteService(t, f)
	s.busy.Lock() // an action is in flight
	if _, err := s.Copy(context.Background(), fx.gif); !errors.Is(err, ErrBusy) {
		t.Errorf("Copy err = %v, want ErrBusy", err)
	}
	if _, err := s.Paste(context.Background(), fx.gif); !errors.Is(err, ErrBusy) {
		t.Errorf("Paste err = %v, want ErrBusy", err)
	}
	if _, err := s.RequestAccessibility(context.Background()); !errors.Is(err, ErrBusy) {
		t.Errorf("RequestAccessibility err = %v, want ErrBusy", err)
	}
	if len(f.Calls) != 0 || fx.hits.Load() != 0 {
		t.Errorf("calls = %q, downloads = %d, want none while busy", f.Calls, fx.hits.Load())
	}
	s.busy.Unlock()
	if _, err := s.Paste(context.Background(), fx.gif); err != nil {
		t.Errorf("Paste after release: %v", err)
	}
}
