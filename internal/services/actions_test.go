package services

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"io"
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

// clipOf turns the fixture item into a Clip whose download is the fixture's 720p MP4 (D-15).
func clipOf(g provider.Gif, name string) provider.Gif {
	base, _, _ := strings.Cut(g.DownloadURL, "/media/")
	return provider.Gif{Provider: provider.Clips, ID: name, Title: "GIPHY clip 1", IsClip: true, DownloadURL: base + "/media/" + name + ".mp4"}
}

func TestCopyClipMP4(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	r, err := s.Copy(context.Background(), clipOf(fx.gif, "clip-1-720p"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Message != "Copied MP4 — press ⌘V to paste" {
		t.Errorf("result = %+v, want Copied with the T1m message", r)
	}
	if len(f.Calls) != 1 || !strings.HasPrefix(f.Calls[0], "CopyFile:"+fx.dir) || !strings.HasSuffix(f.Calls[0], ".mp4") {
		t.Errorf("calls = %q, want one CopyFile of a staged .mp4", f.Calls)
	}
}

func TestPasteClipMP4(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, fx := newPasteService(t, f)
	r, err := s.Paste(context.Background(), clipOf(fx.gif, "clip-1-720p"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Pasted || r.Message != "Pasted MP4" {
		t.Errorf("result = %+v, want Pasted with message Pasted MP4", r)
	}
	if !slices.Contains(f.Calls, "PasteShortcut") {
		t.Errorf("calls = %q, want a ⌘V", f.Calls)
	}
}

func TestCopyClipFailingMediaKinds(t *testing.T) {
	for name, want := range map[string]string{"broken-clip": "download_failed", "html-clip": "not_media", "office-1": "not_media"} {
		f := &platform.Fake{}
		s, fx := newService(t, f)
		g := clipOf(fx.gif, name)
		if name == "office-1" { // a GIF where an MP4 belongs
			g.DownloadURL = fx.gif.DownloadURL
		}
		_, err := s.Copy(context.Background(), g)
		if actionKind(err) != want {
			t.Errorf("%s: err = %v, want kind %s", name, err, want)
		}
		if len(f.Calls) != 0 {
			t.Errorf("%s: calls = %q, want none", name, f.Calls)
		}
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
	// T5a (P14): the T1 text; the Enable Automatic Paste… button beside it carries the call to action.
	if r.Message != "Copied — press ⌘V to paste" {
		t.Errorf("message = %q, want the T5a text", r.Message)
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
	if err != nil || !r.Copied || r.Pasted || !r.NeedsAccessibility || r.Message != "Copied — press ⌘V to paste" {
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

// timers records After calls instead of starting real timers (no 30 s sleeps in tests).
type timers struct {
	d  []time.Duration
	fn []func()
}

func (tm *timers) after(d time.Duration, fn func()) { tm.d, tm.fn = append(tm.d, d), append(tm.fn, fn) }

// klipyService is a trusted paste service with a recording After and a Klipy fixture item.
func klipyService(t *testing.T, f *platform.Fake) (*ActionService, fixture, *timers, provider.Gif) {
	t.Helper()
	s, fx := newPasteService(t, f)
	tm := &timers{}
	s.After = tm.after
	base, _, _ := strings.Cut(fx.gif.DownloadURL, "/media/")
	k := provider.Gif{Provider: provider.Klipy, ID: "8041071659142945", Title: "Klipy moment 1", DownloadURL: base + "/media/klipy-1-hd.gif"}
	return s, fx, tm, k
}

func stagedPath(t *testing.T, f *platform.Fake) string {
	t.Helper()
	for i := len(f.Calls) - 1; i >= 0; i-- {
		if p, ok := strings.CutPrefix(f.Calls[i], "CopyFile:"); ok {
			return p
		}
	}
	t.Fatalf("no CopyFile in %q", f.Calls)
	return ""
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// D-17: a text copy leaves the clipboard pointing at no file, so every staged file goes.
func TestCopyTextClearsStaging(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	if _, err := s.Copy(context.Background(), fx.gif); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CopyText(context.Background(), fx.gif, "link"); err != nil {
		t.Fatal(err)
	}
	if ents, _ := os.ReadDir(fx.dir); len(ents) != 0 {
		t.Errorf("staging holds %d entries after a text copy, want none", len(ents))
	}
}

func TestCopyTextFailureKeepsStaging(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	if _, err := s.Copy(context.Background(), fx.gif); err != nil {
		t.Fatal(err)
	}
	path := stagedPath(t, f)
	f.CopyErr = errors.New("pasteboard write failed")
	if _, err := s.CopyText(context.Background(), fx.gif, "link"); actionKind(err) != "copy_failed" {
		t.Fatalf("err = %v, want copy_failed", err)
	}
	if !exists(path) {
		t.Error("staged file removed although the text copy failed (the clipboard still points at it)")
	}
}

func TestServiceShutdownClears(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	if _, err := s.Copy(context.Background(), fx.gif); err != nil {
		t.Fatal(err)
	}
	s.busy.Lock() // an action still running must not block quitting
	defer s.busy.Unlock()
	if err := s.ServiceShutdown(); err != nil {
		t.Fatal(err)
	}
	if ents, _ := os.ReadDir(fx.dir); len(ents) != 0 {
		t.Errorf("staging holds %d entries after shutdown, want none", len(ents))
	}
}

// D-19: a pasted Klipy file is deleted KlipyPasteGrace after ⌘V.
func TestKlipyGraceDeletesAfterPaste(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, _, tm, k := klipyService(t, f)
	r, err := s.Paste(context.Background(), k)
	if err != nil || !r.Pasted {
		t.Fatalf("Paste = %+v, %v", r, err)
	}
	path := stagedPath(t, f)
	if len(tm.d) != 1 || tm.d[0] != KlipyPasteGrace || KlipyPasteGrace != 30*time.Second {
		t.Fatalf("After calls = %v, want one of 30s", tm.d)
	}
	if !exists(path) {
		t.Fatal("Klipy file deleted before the grace period")
	}
	tm.fn[0]()
	if exists(filepath.Dir(path)) {
		t.Error("Klipy staged dir still there after the grace timer ran")
	}
}

// A later copy of the same Klipy GIF re-stages the same hash dir; the older timer must not delete it.
func TestKlipyGraceGenerationGuard(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, _, tm, k := klipyService(t, f)
	if _, err := s.Paste(context.Background(), k); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Copy(context.Background(), k); err != nil {
		t.Fatal(err)
	}
	path := stagedPath(t, f)
	if len(tm.fn) != 1 {
		t.Fatalf("After calls = %v, want one (Copy schedules none)", tm.d)
	}
	tm.fn[0]()
	if !exists(path) {
		t.Error("an older Klipy timer deleted a newer staged copy")
	}
}

func TestKlipyGraceOnlyAfterKlipyPaste(t *testing.T) {
	for name, tc := range map[string]struct {
		trusted bool
		paste   bool
		klipy   bool
	}{
		"klipy copy":            {true, false, true},
		"klipy paste untrusted": {false, true, true},
		"giphy-like paste":      {true, true, false},
	} {
		f := &platform.Fake{Trusted: tc.trusted, Frontmost: prevApp}
		s, fx, tm, k := klipyService(t, f)
		g := k
		if !tc.klipy {
			g = fx.gif
		}
		call := s.Copy
		if tc.paste {
			call = s.Paste
		}
		if _, err := call(context.Background(), g); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(tm.d) != 0 {
			t.Errorf("%s: After called %v, want never", name, tm.d)
		}
		if !exists(stagedPath(t, f)) {
			t.Errorf("%s: staged file gone, want it kept until the next clipboard write", name)
		}
	}
}

// downloadService saves into a fresh temp folder, never the user's real Downloads.
func downloadService(t *testing.T, f *platform.Fake) (*ActionService, fixture, string) {
	t.Helper()
	s, fx := newService(t, f)
	dl := filepath.Join(t.TempDir(), "Downloads")
	if err := os.Mkdir(dl, 0o755); err != nil {
		t.Fatal(err)
	}
	s.DownloadsDir = func() string { return dl }
	return s, fx, dl
}

func TestDownloadSavesWithoutStagingOrClipboard(t *testing.T) {
	f := &platform.Fake{}
	s, fx, dl := downloadService(t, f)
	r, err := s.Download(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if r != (DownloadResult{Name: "office-moment-1-office-1.gif", Folder: "Downloads"}) {
		t.Errorf("result = %+v", r)
	}
	if !exists(filepath.Join(dl, r.Name)) {
		t.Errorf("%s not in the download folder", r.Name)
	}
	if len(f.Calls) != 0 {
		t.Errorf("calls = %q, want no pasteboard or other platform call", f.Calls)
	}
	if ents, _ := os.ReadDir(fx.dir); len(ents) != 0 {
		t.Errorf("staging holds %d entries after a download, want none", len(ents))
	}
	r2, err := s.Download(context.Background(), fx.gif)
	if err != nil || r2.Name != "office-moment-1-office-1 (2).gif" {
		t.Errorf("second download = %+v, %v; want the (2) name", r2, err)
	}
}

// The folder name in the result follows whatever folder DownloadsDir supplies (Phase 4 picker).
func TestDownloadFolderIsInjected(t *testing.T) {
	s, fx, _ := downloadService(t, &platform.Fake{})
	chosen := filepath.Join(t.TempDir(), "GIF stash")
	if err := os.Mkdir(chosen, 0o755); err != nil {
		t.Fatal(err)
	}
	s.DownloadsDir = func() string { return chosen }
	r, err := s.Download(context.Background(), fx.gif)
	if err != nil || r.Folder != filepath.Base(chosen) || !exists(filepath.Join(chosen, r.Name)) {
		t.Errorf("Download = %+v, %v; want it saved in and named after %q", r, err, chosen)
	}
}

func TestDownloadErrorKinds(t *testing.T) {
	for name, want := range map[string]string{"broken-1": "download_failed", "html-1": "not_media", "missing-folder": "folder_unwritable", "no-folder": "folder_unwritable"} {
		f := &platform.Fake{}
		s, fx, dl := downloadService(t, f)
		g := fx.gif
		switch name {
		case "missing-folder":
			s.DownloadsDir = func() string { return filepath.Join(dl, "gone") }
		case "no-folder":
			s.DownloadsDir = func() string { return "" }
		default:
			g.DownloadURL = strings.Replace(g.DownloadURL, "office-1", name, 1)
		}
		_, err := s.Download(context.Background(), g)
		if actionKind(err) != want || strings.Contains(err.Error(), "/") {
			t.Errorf("%s: err = %v, want kind %s without a path or URL", name, err, want)
		}
		if ents, _ := os.ReadDir(dl); len(ents) != 0 {
			t.Errorf("%s: download folder holds %d entries", name, len(ents))
		}
	}
}

func TestDownloadRefusesWhileBusy(t *testing.T) {
	f := &platform.Fake{}
	s, fx, dl := downloadService(t, f)
	s.busy.Lock()
	defer s.busy.Unlock()
	if _, err := s.Download(context.Background(), fx.gif); !errors.Is(err, ErrBusy) {
		t.Errorf("Download err = %v, want ErrBusy", err)
	}
	if ents, _ := os.ReadDir(dl); len(ents) != 0 || fx.hits.Load() != 0 {
		t.Errorf("busy download wrote %d entries, made %d requests", len(ents), fx.hits.Load())
	}
}

// D-22: Show in Finder reveals only the file Go saved last; JS never sends a path.
func TestShowDownloadRevealsLastDownload(t *testing.T) {
	f := &platform.Fake{}
	s, fx, dl := downloadService(t, f)
	if _, err := s.Download(context.Background(), fx.gif); err != nil {
		t.Fatal(err)
	}
	r, err := s.Download(context.Background(), fx.gif)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ShowDownload(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := []string{"RevealFile:" + filepath.Join(dl, r.Name)}; !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestShowDownloadNothingToShow(t *testing.T) {
	f := &platform.Fake{}
	s, fx, _ := downloadService(t, f)
	if err := s.ShowDownload(context.Background()); actionKind(err) != "nothing_to_show" {
		t.Errorf("err = %v, want kind nothing_to_show", err)
	}
	g := fx.gif
	g.DownloadURL = strings.Replace(g.DownloadURL, "office-1", "broken-1", 1)
	_, _ = s.Download(context.Background(), g) // a failed download leaves nothing to show either
	if err := s.ShowDownload(context.Background()); actionKind(err) != "nothing_to_show" {
		t.Errorf("after a failed download: err = %v, want kind nothing_to_show", err)
	}
	if len(f.Calls) != 0 {
		t.Errorf("calls = %q, want none", f.Calls)
	}
}

func TestShowDownloadRevealError(t *testing.T) {
	f := &platform.Fake{RevealErr: errors.New("finder said no")}
	s, fx, _ := downloadService(t, f)
	if _, err := s.Download(context.Background(), fx.gif); err != nil {
		t.Fatal(err)
	}
	if err := s.ShowDownload(context.Background()); actionKind(err) != "open_failed" {
		t.Errorf("err = %v, want kind open_failed", err)
	}
}

func TestShowDownloadRefusesWhileBusy(t *testing.T) {
	f := &platform.Fake{}
	s, fx, _ := downloadService(t, f)
	if _, err := s.Download(context.Background(), fx.gif); err != nil {
		t.Fatal(err)
	}
	s.busy.Lock()
	defer s.busy.Unlock()
	if err := s.ShowDownload(context.Background()); !errors.Is(err, ErrBusy) {
		t.Errorf("err = %v, want ErrBusy", err)
	}
	if len(f.Calls) != 0 {
		t.Errorf("calls = %q, want none", f.Calls)
	}
}

// squareOf is the fixture item renamed to a fakeapi media file (wide-*: 48x32, tall-*: 32x48).
func squareOf(g provider.Gif, name string) provider.Gif {
	base, _, _ := strings.Cut(g.DownloadURL, "/media/")
	g.ID, g.Title, g.DownloadURL = name, "Wide moment 1", base+"/media/"+name+".gif"
	return g
}

func TestCopySquare(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	g := squareOf(fx.gif, "wide-1")
	r, err := s.CopySquare(context.Background(), g)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Message != "Copied square GIF — press ⌘V to paste" {
		t.Errorf("result = %+v, want Copied with the T1s message", r)
	}
	path := stagedPath(t, f)
	if len(f.Calls) != 1 || !strings.HasSuffix(path, "-square.gif") {
		t.Fatalf("calls = %q, want one CopyFile of a *-square.gif", f.Calls)
	}
	sum := sha1.Sum([]byte(string(g.Provider) + ":" + g.ID))
	if filepath.Base(filepath.Dir(path)) == hex.EncodeToString(sum[:])[:16] {
		t.Error("square file staged in the original's hash directory")
	}
	if ents, _ := os.ReadDir(fx.dir); len(ents) != 1 || ents[0].Name() != filepath.Base(filepath.Dir(path)) {
		t.Errorf("staging holds %d entries, want only the square file's directory", len(ents))
	}
	fh, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fh.Close() }()
	out, err := gif.DecodeAll(fh)
	if err != nil || out.Config.Width != 32 || out.Config.Height != 32 || len(out.Image) != 2 {
		t.Fatalf("staged square: %v, want a 32x32 two-frame GIF (got %+v)", err, out)
	}
}

func TestCopySquareRefusesClips(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	_, err := s.CopySquare(context.Background(), clipOf(fx.gif, "clip-1-720p"))
	if actionKind(err) != "not_allowed" {
		t.Errorf("err = %v, want kind not_allowed", err)
	}
	if n := fx.hits.Load(); n != 0 || len(f.Calls) != 0 {
		t.Errorf("a Clip square crop made %d requests and calls %q, want none", n, f.Calls)
	}
}

// rawService serves body as the only media file; the item points at it.
func rawService(t *testing.T, f *platform.Fake, body io.Reader) (*ActionService, provider.Gif) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/gif")
		_, _ = io.Copy(w, body)
	}))
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	st := &staging.Store{Dir: t.TempDir(), HTTP: srv.Client(), Allow: func(_ provider.ID, u *url.URL) bool { return u.Host == host }}
	g := provider.Gif{Provider: provider.FinerGifs, ID: "raw-1", Title: "Raw 1", DownloadURL: srv.URL + "/media/raw-1.gif"}
	return &ActionService{Platform: f, Staging: st, Hide: func() { f.Note("hide") }}, g
}

// encodeGIF is a one-frame GIF with a w x h canvas (the frame itself is at most 8x8).
func encodeGIF(t *testing.T, w, h int) []byte {
	t.Helper()
	fr := image.NewPaletted(image.Rect(0, 0, min(w, 8), min(h, 8)), color.Palette{color.Black, color.White})
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, &gif.GIF{Image: []*image.Paletted{fr}, Delay: []int{0}, Config: image.Config{Width: w, Height: h}}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// T-03-18: a claimed size over 15 MB is refused before any request.
func TestCopySquareRefusesClaimedSize(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	g := squareOf(fx.gif, "wide-1")
	g.Size = 16000000
	if SquareMaxBytes != 15<<20 || SquareMaxSide != 4096 {
		t.Fatalf("limits = %d bytes, %d px", SquareMaxBytes, SquareMaxSide)
	}
	_, err := s.CopySquare(context.Background(), g)
	if actionKind(err) != "crop_too_large" {
		t.Errorf("err = %v, want kind crop_too_large", err)
	}
	if n := fx.hits.Load(); n != 0 || len(f.Calls) != 0 {
		t.Errorf("an oversized claim made %d requests and calls %q, want none", n, f.Calls)
	}
}

// The 15 MB cap holds whatever the item claims.
func TestCopySquareDownloadOverCap(t *testing.T) {
	f := &platform.Fake{}
	body := io.MultiReader(bytes.NewReader(encodeGIF(t, 10, 10)), io.LimitReader(zeros{}, SquareMaxBytes))
	s, g := rawService(t, f, body)
	_, err := s.CopySquare(context.Background(), g)
	if actionKind(err) != "crop_too_large" {
		t.Errorf("err = %v, want kind crop_too_large", err)
	}
	if ents, _ := os.ReadDir(s.Staging.Dir); len(ents) != 0 || len(f.Calls) != 0 {
		t.Errorf("staging holds %d entries and calls %q after an oversized download, want none", len(ents), f.Calls)
	}
}

type zeros struct{}

func (zeros) Read(p []byte) (int, error) { clear(p); return len(p), nil }

func TestCopySquareCropFailures(t *testing.T) {
	whole := encodeGIF(t, 20, 10)
	for name, tc := range map[string]struct {
		body []byte
		want string
	}{
		"giant canvas": {encodeGIF(t, 5000, 10), "crop_failed"},
		"tall canvas":  {encodeGIF(t, 10, 4097), "crop_failed"},
		"truncated":    {whole[:len(whole)-6], "crop_failed"},
		"html":         {[]byte("<html>error</html>"), "not_media"},
	} {
		f := &platform.Fake{}
		s, g := rawService(t, f, bytes.NewReader(tc.body))
		_, err := s.CopySquare(context.Background(), g)
		if actionKind(err) != tc.want {
			t.Errorf("%s: err = %v, want kind %s", name, err, tc.want)
		}
		if len(f.Calls) != 0 {
			t.Errorf("%s: calls = %q, want none", name, f.Calls)
		}
	}
}

// bombGIF is a 4096x4096 canvas holding n copies of one uniform full-canvas frame: a few hundred
// KB on the wire that would decode to n x 16 MB of frames.
func bombGIF(t *testing.T, n int) []byte {
	t.Helper()
	fr := image.NewPaletted(image.Rect(0, 0, SquareMaxSide, SquareMaxSide), color.Palette{color.Black, color.White})
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, &gif.GIF{Image: []*image.Paletted{fr}, Delay: []int{0}}); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	head := 13 // header and screen descriptor; Go writes a local color table, not a global one
	if b[10]&0x80 != 0 {
		t.Fatal("encoder wrote a global color table")
	}
	frame := b[head : len(b)-1] // everything up to the trailer
	out := append([]byte{}, b[:head]...)
	for range n {
		out = append(out, frame...)
	}
	return append(out, 0x3B)
}

// CR-01: the size and canvas checks bound the input, not the decoded frames; the frame budget
// refuses a small file whose frames would decode to gigabytes, before DecodeAll runs.
func TestCopySquareDecodeBudget(t *testing.T) {
	n := SquareMaxPixels/(SquareMaxSide*SquareMaxSide) + 4
	body := bombGIF(t, n)
	if len(body) > SquareMaxBytes {
		t.Fatalf("bomb is %d bytes, want under the %d byte cap", len(body), SquareMaxBytes)
	}
	if px, err := framePixels(bytes.NewReader(body)); err != nil || px != int64(n)*(SquareMaxSide*SquareMaxSide+frameOverhead) {
		t.Fatalf("framePixels = %d, %v; want %d frames of 4096x4096", px, err, n)
	}
	f := &platform.Fake{}
	s, g := rawService(t, f, bytes.NewReader(body))
	_, err := s.CopySquare(context.Background(), g)
	if actionKind(err) != "crop_failed" {
		t.Errorf("err = %v, want kind crop_failed", err)
	}
	if len(f.Calls) != 0 {
		t.Errorf("calls = %q, want none", f.Calls)
	}
}

// CR-01 (iteration 2): an extension that framePixels and gif.DecodeAll walk differently would hide
// every following frame from the budget. Each prefix is valid to Go, and its bytes read as a
// terminator plus trailer to a naive sub-block walk.
func TestCopySquareDecodeBudgetExtensions(t *testing.T) {
	n := SquareMaxPixels/(SquareMaxSide*SquareMaxSide) + 4
	bomb := bombGIF(t, n)
	for name, ext := range map[string][]byte{
		"plain text":        append([]byte{0x21, 0x01, 0x00, 0x3B}, append(make([]byte, 11), 0x00)...),   // Go reads 13 raw bytes
		"empty application": append([]byte{0x21, 0xFF, 0x00, 0x3B}, append(make([]byte, 0x3B), 0x00)...), // size 0, then sub-blocks
		"unknown extension": {0x21, 0x02, 0x00, 0x3B},                                                    // Go fails the decode
	} {
		body := append(append(append([]byte{}, bomb[:13]...), ext...), bomb[13:]...)
		px, err := framePixels(bytes.NewReader(body))
		if name == "unknown extension" {
			if err == nil {
				t.Errorf("%s: framePixels = %d, want an error", name, px)
			}
		} else if err != nil || px != int64(n)*(SquareMaxSide*SquareMaxSide+frameOverhead) {
			t.Errorf("%s: framePixels = %d, %v; want %d frames of 4096x4096", name, px, err, n)
		}
		f := &platform.Fake{}
		s, g := rawService(t, f, bytes.NewReader(body))
		if _, err := s.CopySquare(context.Background(), g); actionKind(err) != "crop_failed" {
			t.Errorf("%s: err = %v, want kind crop_failed", name, err)
		}
		if len(f.Calls) != 0 {
			t.Errorf("%s: calls = %q, want none", name, f.Calls)
		}
	}
}

// CR-01 (iteration 3): Go keeps ~4 KB per frame (a cloned palette when a frame is transparent)
// whatever its size, so a file of empty 0x0 transparent frames must still exhaust the budget.
func TestCopySquareDecodeBudgetEmptyFrames(t *testing.T) {
	head := []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff") // 1x1, 2-colour global table
	unit := []byte{0x21, 0xF9, 0x04, 0x01, 0, 0, 0, 0, 0x2C, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02, 0x00}
	n := int(SquareMaxPixels/frameOverhead) + 10
	body := append(append(append([]byte{}, head...), bytes.Repeat(unit, n)...), 0x3B)
	if len(body) > SquareMaxBytes {
		t.Fatalf("file is %d bytes, want under the %d byte cap", len(body), SquareMaxBytes)
	}
	f := &platform.Fake{}
	s, g := rawService(t, f, bytes.NewReader(body))
	if _, err := s.CopySquare(context.Background(), g); actionKind(err) != "crop_failed" {
		t.Errorf("err = %v, want kind crop_failed", err)
	}
	if len(f.Calls) != 0 {
		t.Errorf("calls = %q, want none", f.Calls)
	}
}

func TestPasteSquareTrusted(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, fx := newPasteService(t, f)
	r, err := s.PasteSquare(context.Background(), squareOf(fx.gif, "wide-1"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Pasted || r.Message != "Pasted square GIF" {
		t.Errorf("result = %+v, want Pasted with message Pasted square GIF", r)
	}
	path := stagedPath(t, f)
	want := []string{"CopyFile:" + path, "AccessibilityTrusted:false", "hide", "sleep:80ms", "FrontmostApp", "PasteShortcut"}
	if !strings.HasSuffix(path, "-square.gif") || !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestPasteSquareWithoutAccessibility(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newPasteService(t, f)
	r, err := s.PasteSquare(context.Background(), squareOf(fx.gif, "wide-1"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Pasted || !r.NeedsAccessibility {
		t.Errorf("result = %+v, want the T5a copy-only result", r)
	}
	if slices.Contains(f.Calls, "hide") || slices.Contains(f.Calls, "PasteShortcut") {
		t.Errorf("calls = %q, want no hide and no ⌘V", f.Calls)
	}
}

func TestPasteSquareRefusesClips(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, fx := newPasteService(t, f)
	if _, err := s.PasteSquare(context.Background(), clipOf(fx.gif, "clip-1-720p")); actionKind(err) != "not_allowed" {
		t.Errorf("err = %v, want kind not_allowed", err)
	}
}

// D-19: a pasted Klipy square file is deleted KlipyPasteGrace after ⌘V, like Paste.
func TestPasteSquareKlipyGrace(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s, _, tm, k := klipyService(t, f)
	if r, err := s.PasteSquare(context.Background(), k); err != nil || !r.Pasted {
		t.Fatalf("PasteSquare = %+v, %v", r, err)
	}
	path := stagedPath(t, f)
	if len(tm.d) != 1 || tm.d[0] != KlipyPasteGrace {
		t.Fatalf("After calls = %v, want one of %v", tm.d, KlipyPasteGrace)
	}
	tm.fn[0]()
	if exists(filepath.Dir(path)) {
		t.Error("Klipy square dir still there after the grace timer ran")
	}
}

// recentsRecorder is ActionService.Recents for tests: it records provider:id per Touch.
type recentsRecorder struct {
	keys []string
	err  error
}

func (r *recentsRecorder) Touch(g provider.Gif) error {
	r.keys = append(r.keys, string(g.Provider)+":"+g.ID)
	return r.err
}

// Every successful copy or paste lands in Recents once, including copy-only paste fallbacks (D-07).
func TestRecentsRecordedOnEverySuccessfulCopy(t *testing.T) {
	ctx := context.Background()
	for name, run := range map[string]func(s *ActionService, g provider.Gif) (Result, error){
		"Copy":           func(s *ActionService, g provider.Gif) (Result, error) { return s.Copy(ctx, g) },
		"Paste":          func(s *ActionService, g provider.Gif) (Result, error) { return s.Paste(ctx, g) },
		"CopyText link":  func(s *ActionService, g provider.Gif) (Result, error) { return s.CopyText(ctx, g, "link") },
		"CopyText md":    func(s *ActionService, g provider.Gif) (Result, error) { return s.CopyText(ctx, g, "markdown") },
		"PasteText link": func(s *ActionService, g provider.Gif) (Result, error) { return s.PasteText(ctx, g, "link") },
		"PasteText md":   func(s *ActionService, g provider.Gif) (Result, error) { return s.PasteText(ctx, g, "markdown") },
		"CopySquare": func(s *ActionService, g provider.Gif) (Result, error) {
			return s.CopySquare(ctx, squareOf(g, "wide-1"))
		},
		"PasteSquare": func(s *ActionService, g provider.Gif) (Result, error) {
			return s.PasteSquare(ctx, squareOf(g, "wide-1"))
		},
	} {
		f := &platform.Fake{} // untrusted: pastes fall back to copy only, which still counts
		s, fx := newPasteService(t, f)
		rec := &recentsRecorder{}
		s.Recents = rec
		r, err := run(s, fx.gif)
		if err != nil || !r.Copied {
			t.Fatalf("%s: %+v, %v", name, r, err)
		}
		if len(rec.keys) != 1 {
			t.Errorf("%s: %d Touch calls, want 1", name, len(rec.keys))
		}
	}
	// A trusted paste that posted ⌘V records too.
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	s := vendorService(f)
	rec := &recentsRecorder{}
	s.Recents = rec
	if r, err := s.PasteText(ctx, giphyGif, "link"); err != nil || !r.Pasted || len(rec.keys) != 1 || rec.keys[0] != "giphy:a" {
		t.Errorf("trusted PasteText: %+v, %v, touches %q", r, err, rec.keys)
	}
}

// Page links, downloads, failures and busy refusals never touch Recents.
func TestRecentsNotRecorded(t *testing.T) {
	ctx := context.Background()
	f := &platform.Fake{}
	s := vendorService(f)
	rec := &recentsRecorder{}
	s.Recents = rec
	if _, err := s.CopyText(ctx, giphyGif, "page"); err != nil {
		t.Fatal(err)
	}
	ds, fx, _ := downloadService(t, &platform.Fake{})
	ds.Recents = rec
	if _, err := ds.Download(ctx, fx.gif); err != nil {
		t.Fatal(err)
	}
	broken := fx.gif
	broken.DownloadURL = strings.Replace(broken.DownloadURL, "office-1", "broken-1", 1)
	if _, err := ds.Copy(ctx, broken); err == nil {
		t.Fatal("copy of broken media succeeded")
	}
	ds.busy.Lock()
	if _, err := ds.Copy(ctx, fx.gif); !errors.Is(err, ErrBusy) {
		t.Fatalf("busy copy: %v", err)
	}
	ds.busy.Unlock()
	if len(rec.keys) != 0 {
		t.Errorf("touches %q, want none", rec.keys)
	}
}

// A Recents write failure is logged; the action still reports its result.
func TestRecentsErrorDoesNotFailAction(t *testing.T) {
	f := &platform.Fake{}
	s, fx := newService(t, f)
	rec := &recentsRecorder{err: errors.New("disk full")}
	s.Recents = rec
	r, err := s.Copy(context.Background(), fx.gif)
	if err != nil || !r.Copied || len(rec.keys) != 1 {
		t.Errorf("copy with failing Recents: %+v, %v, touches %q", r, err, rec.keys)
	}
}

func TestDownloadHideFilenames(t *testing.T) {
	s, fx, dl := downloadService(t, &platform.Fake{})
	s.HideFilenames = func() bool { return true }
	for _, want := range []string{"gifsmith-office-1.gif", "gifsmith-office-1 (2).gif"} {
		r, err := s.Download(context.Background(), fx.gif)
		if err != nil || r.Name != want || !exists(filepath.Join(dl, want)) {
			t.Fatalf("Download with hidden names = %+v, %v; want %s saved", r, err, want)
		}
	}
}

func TestDownloadDeletedFolder(t *testing.T) {
	s, fx, dl := downloadService(t, &platform.Fake{})
	if err := os.Remove(dl); err != nil {
		t.Fatal(err)
	}
	r, err := s.Download(context.Background(), fx.gif)
	if actionKind(err) != "folder_unwritable" || r != (DownloadResult{}) {
		t.Errorf("Download into a deleted folder = %+v, %v; want folder_unwritable and an empty result", r, err)
	}
}
