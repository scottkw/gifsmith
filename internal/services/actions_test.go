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
	if actionKind(err) != "copy_failed" {
		t.Errorf("err = %v, want kind copy_failed", err)
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
	if actionKind(err) != "not_allowed" || strings.Contains(err.Error(), "evil.example") {
		t.Errorf("err = %v, want kind not_allowed without the URL", err)
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
	if !strings.Contains(r.Message, "Enable Automatic Paste…") || strings.Contains(r.Message, "System Settings") {
		t.Errorf("message = %q, want it to name the Enable Automatic Paste… button and no Settings path", r.Message)
	}
	// App.svelte shows "Enable Automatic Paste…" off this exact JSON key (models.ts Result).
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
	if actionKind(err) != "copy_failed" {
		t.Errorf("err = %v, want kind copy_failed", err)
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
	if _, err := s.CopyText(context.Background(), fx.gif, "link"); !errors.Is(err, ErrBusy) {
		t.Errorf("CopyText err = %v, want ErrBusy", err)
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

func actionKind(err error) string {
	var ae *ActionError
	if errors.As(err, &ae) {
		return ae.Kind
	}
	return ""
}

func TestCopyFailingMediaKinds(t *testing.T) {
	for name, want := range map[string]string{"broken-1": "download_failed", "html-1": "not_media"} {
		f := &platform.Fake{}
		s, fx := newService(t, f)
		g := fx.gif
		g.DownloadURL = strings.Replace(g.DownloadURL, "office-1", name, 1)
		_, err := s.Copy(context.Background(), g)
		if actionKind(err) != want || strings.Contains(err.Error(), "http") {
			t.Errorf("%s: err = %v, want kind %s without a URL", name, err, want)
		}
		if len(f.Calls) != 0 {
			t.Errorf("%s: calls = %q, want none", name, f.Calls)
		}
	}
}

func TestMarshalError(t *testing.T) {
	if got := string(MarshalError(ErrBusy)); got != `{"kind":"busy"}` {
		t.Errorf("MarshalError(ErrBusy) = %s", got)
	}
	if ErrBusy.Error() != "another action is still running" {
		t.Errorf("ErrBusy text = %q; App.svelte compares it", ErrBusy.Error())
	}
	if b := MarshalError(&provider.Error{Kind: provider.Network, Provider: provider.GIPHY}); b != nil {
		t.Errorf("MarshalError(provider error) = %s, want nil (main.go composes provider first)", b)
	}
	if b := MarshalError(errors.New("x")); b != nil {
		t.Errorf("MarshalError(plain) = %s, want nil", b)
	}
}

func TestCopyTextLink(t *testing.T) {
	for _, tc := range []struct {
		p         provider.ID
		raw, want string
	}{
		{provider.GIPHY, "https://media.giphy.com/media/a/giphy.gif?cid=1&rid=2", "https://media.giphy.com/media/a/giphy.gif"},
		// Klipy forbids altering URLs; Finer links have no query to strip.
		{provider.Klipy, "https://static.klipy.com/ii/a/b.gif?x=1#f", "https://static.klipy.com/ii/a/b.gif?x=1#f"},
		{provider.FinerGifs, "https://media.thefinergifs.club/office-1.gif", "https://media.thefinergifs.club/office-1.gif"},
	} {
		f := &platform.Fake{}
		s := &ActionService{Platform: f, Staging: &staging.Store{}} // nil Allow = VendorAllow
		r, err := s.CopyText(context.Background(), provider.Gif{Provider: tc.p, ID: "a", DownloadURL: tc.raw}, "link")
		if err != nil {
			t.Fatalf("%s: %v", tc.p, err)
		}
		if !r.Copied || r.Message != "Copied link" {
			t.Errorf("%s: result = %+v", tc.p, r)
		}
		if want := []string{"CopyText:" + tc.want}; !slices.Equal(f.Calls, want) {
			t.Errorf("%s: calls = %q, want %q", tc.p, f.Calls, want)
		}
	}
}

func TestCopyTextRefusesDisallowedLink(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	g := fx.gif
	g.DownloadURL = "https://evil.example.test/x.gif"
	_, err := s.CopyText(context.Background(), g, "link")
	if actionKind(err) != "not_allowed" || strings.Contains(err.Error(), "evil") {
		t.Errorf("err = %v, want kind not_allowed without the URL", err)
	}
	if len(f.Calls) != 0 {
		t.Errorf("calls = %q, want none", f.Calls)
	}
}

func TestCopyTextCopyError(t *testing.T) {
	f := &platform.Fake{CopyErr: errors.New("pasteboard write failed")}
	s, fx := newService(t, f)
	if _, err := s.CopyText(context.Background(), fx.gif, "link"); actionKind(err) != "copy_failed" {
		t.Errorf("err = %v, want kind copy_failed", err)
	}
}

var (
	giphyGif = provider.Gif{Provider: provider.GIPHY, ID: "a", Title: "Cat", DownloadURL: "https://media.giphy.com/media/a/giphy.gif?cid=1", PageURL: "https://giphy.com/gifs/x"}
	klipyGif = provider.Gif{Provider: provider.Klipy, ID: "k", Title: "K", DownloadURL: "https://static.klipy.com/k.gif"}
	finerGif = provider.Gif{Provider: provider.FinerGifs, ID: "f", Title: "F", DownloadURL: "https://media.thefinergifs.club/f.gif"}
)

// vendorService uses the real vendor allowlist (nil Allow) and never downloads.
func vendorService(f *platform.Fake) *ActionService {
	return &ActionService{Platform: f, Staging: &staging.Store{}, Hide: func() { f.Note("hide") },
		Previous: func() platform.App { return prevApp }, Sleep: func(d time.Duration) { f.Note("sleep:" + d.String()) }}
}

func TestCopyTextMarkdown(t *testing.T) {
	for title, want := range map[string]string{
		"Cat [x]\nnap\\": `![Cat \[x\] nap\\](https://media.giphy.com/media/a/giphy.gif)`,
		" \t":            `![GIF](https://media.giphy.com/media/a/giphy.gif)`,
	} {
		f := &platform.Fake{}
		g := giphyGif
		g.Title = title
		r, err := vendorService(f).CopyText(context.Background(), g, "markdown")
		if err != nil || r.Message != "Copied Markdown" || !r.Copied {
			t.Fatalf("%q: %+v, %v", title, r, err)
		}
		if w := []string{"CopyText:" + want}; !slices.Equal(f.Calls, w) {
			t.Errorf("%q: calls = %q, want %q", title, f.Calls, w)
		}
	}
}

func TestCopyTextPage(t *testing.T) {
	f := &platform.Fake{}
	r, err := vendorService(f).CopyText(context.Background(), giphyGif, "page")
	if err != nil || r.Message != "Copied page link" {
		t.Fatalf("%+v, %v", r, err)
	}
	if w := []string{"CopyText:https://giphy.com/gifs/x"}; !slices.Equal(f.Calls, w) {
		t.Errorf("calls = %q, want %q", f.Calls, w)
	}
	evil := giphyGif
	evil.PageURL = "https://evil.example/gifs/x"
	klipyPage := klipyGif
	klipyPage.PageURL = "https://giphy.com/gifs/x" // only GIPHY and Clips have pages
	for _, g := range []provider.Gif{klipyGif, finerGif, evil, klipyPage} {
		f := &platform.Fake{}
		if _, err := vendorService(f).CopyText(context.Background(), g, "page"); actionKind(err) != "no_page" || len(f.Calls) != 0 {
			t.Errorf("%s %q: err = %v, calls = %q; want no_page and no calls", g.Provider, g.PageURL, err, f.Calls)
		}
	}
}

// D-18: a Clip's link is its GIF rendition with the query removed, never the MP4.
func TestCopyTextClipLink(t *testing.T) {
	clip := provider.Gif{Provider: provider.Clips, ID: "c", IsClip: true, DownloadURL: "https://media.giphy.com/c.mp4", GifURL: "https://media.giphy.com/c/giphy.gif?cid=9"}
	f := &platform.Fake{}
	if _, err := vendorService(f).CopyText(context.Background(), clip, "link"); err != nil {
		t.Fatal(err)
	}
	if w := []string{"CopyText:https://media.giphy.com/c/giphy.gif"}; !slices.Equal(f.Calls, w) {
		t.Errorf("calls = %q, want %q", f.Calls, w)
	}
	clip.GifURL = ""
	f = &platform.Fake{}
	if _, err := vendorService(f).CopyText(context.Background(), clip, "markdown"); actionKind(err) != "no_link" || len(f.Calls) != 0 {
		t.Errorf("clip without GifURL: err = %v, calls = %q; want no_link", err, f.Calls)
	}
}

func TestCopyTextUnknownKind(t *testing.T) {
	f := &platform.Fake{}
	if _, err := vendorService(f).CopyText(context.Background(), giphyGif, "file:///etc/passwd"); actionKind(err) != "not_allowed" || len(f.Calls) != 0 {
		t.Errorf("err = %v, calls = %q; want not_allowed", err, f.Calls)
	}
}

func TestPasteTextTrusted(t *testing.T) {
	for kind, msg := range map[string]string{"link": "Pasted link", "markdown": "Pasted Markdown"} {
		f := &platform.Fake{Trusted: true, Frontmost: prevApp}
		r, err := vendorService(f).PasteText(context.Background(), giphyGif, kind)
		if err != nil || !r.Copied || !r.Pasted || r.Message != msg {
			t.Fatalf("%s: %+v, %v", kind, r, err)
		}
		if len(f.Calls) == 0 || !strings.HasPrefix(f.Calls[0], "CopyText:") {
			t.Fatalf("%s: calls = %q", kind, f.Calls)
		}
		want := []string{"AccessibilityTrusted:false", "hide", "sleep:80ms", "FrontmostApp", "PasteShortcut"}
		if !slices.Equal(f.Calls[1:], want) {
			t.Errorf("%s: calls = %q, want CopyText then %q", kind, f.Calls, want)
		}
	}
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	if _, err := vendorService(f).PasteText(context.Background(), giphyGif, "page"); actionKind(err) != "not_allowed" || len(f.Calls) != 0 {
		t.Errorf("PasteText page: err = %v, calls = %q; want not_allowed", err, f.Calls)
	}
}

func TestPasteTextWithoutAccessibility(t *testing.T) {
	f := &platform.Fake{}
	r, err := vendorService(f).PasteText(context.Background(), giphyGif, "link")
	if err != nil || !r.Copied || r.Pasted || !r.NeedsAccessibility || !strings.Contains(r.Message, "Enable Automatic Paste…") {
		t.Errorf("%+v, %v; want copied only with the T5a message", r, err)
	}
	if slices.Contains(f.Calls, "hide") || slices.Contains(f.Calls, "PasteShortcut") {
		t.Errorf("calls = %q", f.Calls)
	}
}

func TestOpenPage(t *testing.T) {
	f := &platform.Fake{}
	if err := vendorService(f).OpenPage(context.Background(), giphyGif); err != nil {
		t.Fatal(err)
	}
	if w := []string{"OpenURL:https://giphy.com/gifs/x"}; !slices.Equal(f.Calls, w) {
		t.Errorf("calls = %q, want %q", f.Calls, w)
	}
	for _, g := range []provider.Gif{klipyGif, finerGif, {Provider: provider.GIPHY, PageURL: "javascript:alert(1)"}} {
		f := &platform.Fake{}
		if err := vendorService(f).OpenPage(context.Background(), g); actionKind(err) != "no_page" || len(f.Calls) != 0 {
			t.Errorf("%s %q: err = %v, calls = %q; want no_page", g.Provider, g.PageURL, err, f.Calls)
		}
	}
	f = &platform.Fake{OpenErr: errors.New("no browser")}
	if err := vendorService(f).OpenPage(context.Background(), giphyGif); actionKind(err) != "open_failed" {
		t.Errorf("err = %v, want open_failed", err)
	}
}

func TestOpenSource(t *testing.T) {
	f := &platform.Fake{}
	g := giphyGif
	g.SourceURL = "https://example.org/a"
	if err := vendorService(f).OpenSource(context.Background(), g); err != nil {
		t.Fatal(err)
	}
	if w := []string{"OpenURL:https://example.org/a"}; !slices.Equal(f.Calls, w) {
		t.Errorf("calls = %q, want %q", f.Calls, w)
	}
	for _, src := range []string{"", "http://example.org/a", "file:///etc/passwd", "https://"} {
		f := &platform.Fake{}
		g.SourceURL = src
		if err := vendorService(f).OpenSource(context.Background(), g); actionKind(err) != "no_page" || len(f.Calls) != 0 {
			t.Errorf("%q: err = %v, calls = %q; want no_page", src, err, f.Calls)
		}
	}
}

func TestTextActionsRefuseWhileBusy(t *testing.T) {
	f := &platform.Fake{Trusted: true}
	s := vendorService(f)
	s.busy.Lock()
	defer s.busy.Unlock()
	if _, err := s.PasteText(context.Background(), giphyGif, "link"); !errors.Is(err, ErrBusy) {
		t.Errorf("PasteText err = %v", err)
	}
	if err := s.OpenPage(context.Background(), giphyGif); !errors.Is(err, ErrBusy) {
		t.Errorf("OpenPage err = %v", err)
	}
	if err := s.OpenSource(context.Background(), giphyGif); !errors.Is(err, ErrBusy) {
		t.Errorf("OpenSource err = %v", err)
	}
	if len(f.Calls) != 0 {
		t.Errorf("calls = %q", f.Calls)
	}
}
