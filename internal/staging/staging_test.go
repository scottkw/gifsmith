package staging

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"unicode/utf8"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
)

// newStore serves h (fakeapi media by default) and allows only that server's host.
func newStore(t *testing.T, h http.Handler) (*Store, string, *atomic.Int32) {
	t.Helper()
	if h == nil {
		h = fakeapi.Handler(fakeapi.Options{})
	}
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	s := &Store{Dir: t.TempDir(), HTTP: srv.Client(), Allow: func(_ provider.ID, u *url.URL) bool { return u.Host == host }}
	return s, srv.URL, &hits
}

func gif(base, name, title string) provider.Gif {
	return provider.Gif{Provider: provider.FinerGifs, ID: name, Title: title, DownloadURL: base + "/media/" + name + ".gif"}
}

func entries(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if p != dir {
			rel, _ := filepath.Rel(dir, p)
			out = append(out, rel)
		}
		return nil
	})
	return out
}

func noURL(t *testing.T, err error, base string) {
	t.Helper()
	if strings.Contains(err.Error(), base) || strings.Contains(err.Error(), "http") {
		t.Errorf("error text %q contains the URL", err)
	}
}

func TestFetchStagesGIF(t *testing.T) {
	s, base, _ := newStore(t, nil)
	p, err := s.Fetch(context.Background(), gif(base, "office-7", "Office moment 7"))
	if err != nil {
		t.Fatal(err)
	}
	rel, _ := filepath.Rel(s.Dir, p)
	if ok, _ := regexp.MatchString(`^[0-9a-f]{16}/office-moment-7\.gif$`, filepath.ToSlash(rel)); !ok {
		t.Errorf("staged at %q, want <16 hex>/office-moment-7.gif under Dir", rel)
	}
	b, err := os.ReadFile(p)
	if err != nil || !strings.HasPrefix(string(b), "GIF8") {
		t.Errorf("staged bytes start %q (err %v), want GIF8", b[:min(len(b), 6)], err)
	}
	for _, e := range entries(t, s.Dir) {
		if strings.HasSuffix(e, ".part") {
			t.Errorf("leftover %s", e)
		}
	}
}

func TestFetchNotGIF(t *testing.T) {
	s, base, _ := newStore(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>"))
	}))
	_, err := s.Fetch(context.Background(), gif(base, "x", "x"))
	if !errors.Is(err, ErrNotGIF) {
		t.Fatalf("err = %v, want ErrNotGIF", err)
	}
	noURL(t, err, base)
	if e := entries(t, s.Dir); len(e) != 0 {
		t.Errorf("Dir not empty after a rejected download: %q", e)
	}
}

func clip(base, name, title string) provider.Gif {
	return provider.Gif{Provider: provider.Clips, ID: name, Title: title, IsClip: true, DownloadURL: base + "/media/" + name + ".mp4"}
}

// D-15: a Clip is staged as <slug>.mp4 after the ftyp check.
func TestFetchStagesMP4(t *testing.T) {
	s, base, _ := newStore(t, nil)
	p, err := s.Fetch(context.Background(), clip(base, "clip-1-720p", "GIPHY clip 1"))
	if err != nil {
		t.Fatal(err)
	}
	rel, _ := filepath.Rel(s.Dir, p)
	if ok, _ := regexp.MatchString(`^[0-9a-f]{16}/giphy-clip-1\.mp4$`, filepath.ToSlash(rel)); !ok {
		t.Errorf("staged at %q, want <16 hex>/giphy-clip-1.mp4 under Dir", rel)
	}
	if b, err := os.ReadFile(p); err != nil || len(b) < 8 || string(b[4:8]) != "ftyp" {
		t.Errorf("staged bytes %q (err %v), want an ftyp box", b, err)
	}
}

func TestFetchNotMP4(t *testing.T) {
	s, base, _ := newStore(t, nil)
	for _, g := range []provider.Gif{
		clip(base, "html-clip", "x"), // HTML page with 200
		{Provider: provider.Clips, ID: "g", Title: "x", IsClip: true, DownloadURL: base + "/media/office-1.gif"}, // a real GIF
	} {
		_, err := s.Fetch(context.Background(), g)
		if !errors.Is(err, ErrNotMP4) {
			t.Errorf("%s: err = %v, want ErrNotMP4", g.DownloadURL, err)
		}
		if e := entries(t, s.Dir); len(e) != 0 {
			t.Errorf("Dir not empty after a rejected download: %q", e)
		}
	}
	// A GIF item served an MP4 is still not a GIF.
	if _, err := s.Fetch(context.Background(), gif(base, "office-1", "x")); err != nil {
		t.Fatalf("GIF fetch: %v", err)
	}
	g := gif(base, "clip-1", "x")
	g.DownloadURL = base + "/media/clip-1.mp4"
	if _, err := s.Fetch(context.Background(), g); !errors.Is(err, ErrNotGIF) {
		t.Errorf("GIF item with MP4 bytes: err = %v, want ErrNotGIF", err)
	}
}

func TestFetchHTTPError(t *testing.T) {
	s, base, _ := newStore(t, http.NotFoundHandler())
	_, err := s.Fetch(context.Background(), gif(base, "x", "x"))
	if !errors.Is(err, ErrDownload) {
		t.Fatalf("err = %v, want ErrDownload", err)
	}
	noURL(t, err, base)
}

func TestFetchNetworkError(t *testing.T) {
	s := &Store{Dir: t.TempDir(), HTTP: http.DefaultClient, Allow: func(provider.ID, *url.URL) bool { return true }}
	_, err := s.Fetch(context.Background(), gif("http://127.0.0.1:1", "x", "x"))
	if !errors.Is(err, ErrDownload) {
		t.Fatalf("err = %v, want ErrDownload", err)
	}
	noURL(t, err, "127.0.0.1")
}

func TestFetchTooLarge(t *testing.T) {
	s, base, _ := newStore(t, nil)
	s.maxBytes = 64 // fakeapi GIFs are a few hundred bytes
	_, err := s.Fetch(context.Background(), gif(base, "office-1", "big"))
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
	noURL(t, err, base)
	if e := entries(t, s.Dir); len(e) != 0 {
		t.Errorf("Dir not empty after an oversized download: %q", e)
	}
}

func TestFetchDisallowedHost(t *testing.T) {
	s, base, hits := newStore(t, nil)
	s.Allow = nil // VendorAllow: the fixture host is not a vendor host
	_, err := s.Fetch(context.Background(), gif(base, "office-1", "x"))
	if !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("err = %v, want ErrNotAllowed", err)
	}
	noURL(t, err, base)
	if n := hits.Load(); n != 0 {
		t.Errorf("%d requests made to a disallowed host", n)
	}
}

func TestFetchTraversalSlug(t *testing.T) {
	s, base, _ := newStore(t, nil)
	for _, title := range []string{"../../etc/passwd", "", "/", "..", "a/../../b"} {
		p, err := s.Fetch(context.Background(), gif(base, "office-2", title))
		if err != nil {
			t.Fatalf("title %q: %v", title, err)
		}
		rel, err := filepath.Rel(s.Dir, p)
		if err != nil || strings.HasPrefix(rel, "..") || filepath.Dir(filepath.Dir(rel)) != "." {
			t.Errorf("title %q staged at %q, outside <Dir>/<hash>/", title, p)
		}
	}
	if got := slug(""); got != "gif" {
		t.Errorf("slug(\"\") = %q, want gif", got)
	}
	if got := slug("../../etc/passwd"); got != "etc-passwd" {
		t.Errorf("slug(traversal) = %q, want etc-passwd", got)
	}
}

func TestSlugCap(t *testing.T) {
	long := strings.Repeat("é", 100) // 200 bytes
	got := slug(long)
	if len(got) > 80 || !utf8.ValidString(got) || got != strings.Repeat("é", 40) {
		t.Errorf("slug(200-byte title) = %q (%d bytes), want 40 é (80 bytes)", got, len(got))
	}
	if got := slug("Bears. Beets. Battlestar Galactica."); got != "bears-beets-battlestar-galactica" {
		t.Errorf("slug = %q", got)
	}
}

func TestSweep(t *testing.T) {
	s, base, _ := newStore(t, nil)
	old, err := s.Fetch(context.Background(), gif(base, "office-3", "old"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Dir, "sample.gif"), []byte("GIF89a"), 0o644); err != nil { // Phase 1 leftover
		t.Fatal(err)
	}
	keep, err := s.Fetch(context.Background(), gif(base, "office-4", "keep"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Sweep(keep); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("kept file gone: %v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("old staged file still there (err %v)", err)
	}
	rel, _ := filepath.Rel(s.Dir, keep)
	if e := entries(t, s.Dir); len(e) != 2 || e[0] != filepath.Dir(rel) {
		t.Errorf("Dir entries = %q, want only %s", e, filepath.Dir(rel))
	}
}

func TestVendorAllow(t *testing.T) {
	cases := []struct {
		p    provider.ID
		u    string
		want bool
	}{
		{provider.GIPHY, "https://media2.giphy.com/x.gif", true},
		{provider.Clips, "https://media.giphy.com/x.gif", true},
		{provider.GIPHY, "https://giphy.com/x.gif", true},
		{provider.GIPHY, "http://media2.giphy.com/x.gif", false},
		{provider.GIPHY, "https://giphy.com.evil.test/x.gif", false},
		{provider.GIPHY, "https://evilgiphy.com/x.gif", false},
		{provider.GIPHY, "https://media2.giphy.com:8443/x.gif", false},
		{provider.Klipy, "https://static.klipy.com/x.gif", true},
		{provider.Klipy, "https://klipy.com/x.gif", true},
		{provider.Klipy, "https://media2.giphy.com/x.gif", false},
		{provider.FinerGifs, "https://media.thefinergifs.club/a.gif", true},
		{provider.FinerGifs, "https://api.thefinergifs.club/a.gif", false},
		{provider.GIPHY, "https://media.thefinergifs.club/a.gif", false},
		{"", "https://media2.giphy.com/x.gif", false},
	}
	for _, c := range cases {
		u, err := url.Parse(c.u)
		if err != nil {
			t.Fatal(err)
		}
		if got := VendorAllow(c.p, u); got != c.want {
			t.Errorf("VendorAllow(%s, %s) = %v, want %v", c.p, c.u, got, c.want)
		}
	}
}

// A redirect is checked against the allowlist on every hop (WR-03).
func TestFetchRedirectAllowlist(t *testing.T) {
	var offHits atomic.Int32
	off := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { offHits.Add(1) }))
	t.Cleanup(off.Close)
	media := fakeapi.Handler(fakeapi.Options{})
	s, base, _ := newStore(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/media/away.gif":
			http.Redirect(w, r, off.URL+"/x.gif", http.StatusFound)
		case "/media/home.gif":
			http.Redirect(w, r, "/media/office-7.gif", http.StatusFound)
		case "/media/loop.gif":
			http.Redirect(w, r, "/media/loop.gif", http.StatusFound)
		default:
			media.ServeHTTP(w, r)
		}
	}))
	_, err := s.Fetch(context.Background(), gif(base, "away", "Away"))
	if !errors.Is(err, ErrNotAllowed) || offHits.Load() != 0 {
		t.Errorf("redirect off the allowlist: err %v, off-host hits %d; want ErrNotAllowed, 0", err, offHits.Load())
	}
	if _, err := s.Fetch(context.Background(), gif(base, "home", "Home")); err != nil {
		t.Errorf("redirect within the allowlist: %v", err)
	}
	if _, err := s.Fetch(context.Background(), gif(base, "loop", "Loop")); !errors.Is(err, ErrNotAllowed) {
		t.Errorf("redirect loop: %v, want ErrNotAllowed", err)
	}
}

func TestPageAllow(t *testing.T) {
	for _, tc := range []struct {
		p    provider.ID
		u    string
		want bool
	}{
		{provider.GIPHY, "https://giphy.com/gifs/x", true},
		{provider.Clips, "https://giphy.com/clips/x", true},
		{provider.GIPHY, "https://media.giphy.com/x", true},
		{provider.Klipy, "https://klipy.com/x", false},
		{provider.Klipy, "https://giphy.com/gifs/x", false},
		{provider.FinerGifs, "https://giphy.com/gifs/x", false},
		{provider.GIPHY, "http://giphy.com/x", false},
		{provider.GIPHY, "https://giphy.com:8443/x", false},
		{provider.GIPHY, "https://evilgiphy.com/x", false},
		{provider.GIPHY, "https://giphy.com.evil.example/x", false},
		{provider.GIPHY, "file:///etc/passwd", false},
	} {
		u, err := url.Parse(tc.u)
		if err != nil {
			t.Fatal(err)
		}
		if got := PageAllow(tc.p, u); got != tc.want {
			t.Errorf("PageAllow(%s, %s) = %v, want %v", tc.p, tc.u, got, tc.want)
		}
	}
}

func TestClear(t *testing.T) {
	s, base, _ := newStore(t, nil)
	for _, n := range []string{"office-1", "office-2"} {
		if _, err := s.Fetch(context.Background(), gif(base, n, n)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if e := entries(t, s.Dir); len(e) != 0 {
		t.Errorf("Dir holds %q after Clear, want nothing", e)
	}
	if _, err := os.Stat(s.Dir); err != nil {
		t.Errorf("Clear removed Dir itself: %v", err)
	}
	missing := &Store{Dir: filepath.Join(t.TempDir(), "never-created")}
	if err := missing.Clear(); err != nil {
		t.Errorf("Clear on a missing Dir = %v, want nil", err)
	}
}

func TestRemove(t *testing.T) {
	s, base, _ := newStore(t, nil)
	a, err := s.Fetch(context.Background(), gif(base, "office-1", "a"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Fetch(context.Background(), gif(base, "office-2", "b"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(filepath.Dir(a)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(a)); !os.IsNotExist(err) {
		t.Errorf("removed dir still there (err %v)", err)
	}
	if _, err := os.Stat(b); err != nil {
		t.Errorf("other staged file gone: %v", err)
	}
	outside := t.TempDir()
	victim := filepath.Join(outside, "keep.txt")
	if err := os.WriteFile(victim, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{filepath.Join(s.Dir, "..", filepath.Base(outside)), outside, s.Dir, filepath.Join(s.Dir, "../x")} {
		if err := s.Remove(d); !errors.Is(err, ErrNotAllowed) {
			t.Errorf("Remove(%q) = %v, want ErrNotAllowed", d, err)
		}
	}
	if _, err := os.Stat(victim); err != nil {
		t.Errorf("file outside Dir removed: %v", err)
	}
	if _, err := os.Stat(b); err != nil {
		t.Errorf("Remove(Dir) emptied the store: %v", err)
	}
}
