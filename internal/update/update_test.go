package update

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCompare(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.1", -1}, {"1.0.1", "1.1.0", -1}, {"1.1.0", "2.0.0", -1},
		{"1.10.0", "1.9.0", 1}, {"v1.2.3", "1.2.3", 0}, {"2.0.0", "1.99.99", 1},
	} {
		got, err := Compare(c.a, c.b)
		if err != nil || got != c.want {
			t.Errorf("Compare(%q, %q) = %d, %v; want %d", c.a, c.b, got, err, c.want)
		}
	}
}

func TestParseVersion(t *testing.T) {
	if v, err := ParseVersion("v1.10.3"); err != nil || v != [3]int{1, 10, 3} {
		t.Errorf("ParseVersion(v1.10.3) = %v, %v", v, err)
	}
	for _, s := range []string{"", "v", "1.2", "1.2.3.4", "1.2.3-beta.1", "1.2.x", " 1.2.3", "01.2.3"} {
		if _, err := ParseVersion(s); err == nil {
			t.Errorf("ParseVersion(%q) accepted", s)
		}
	}
}

func TestDMGAssetName(t *testing.T) {
	if got := DMGAssetName("1.1.0"); got != "Gifsmith-1.1.0-arm64.dmg" {
		t.Errorf("DMGAssetName = %q", got)
	}
}

const digest = "sha256:" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func asset(name string) map[string]any {
	return map[string]any{
		"name": name, "size": 1234, "state": "uploaded", "digest": digest,
		"browser_download_url": "https://github.com/scottkw/gifsmith/releases/download/v1.1.0/" + name,
	}
}

func release(tag string, assets ...map[string]any) map[string]any {
	if assets == nil {
		assets = []map[string]any{asset(DMGAssetName(strings.TrimPrefix(tag, "v")))}
	}
	return map[string]any{"tag_name": tag, "draft": false, "prerelease": false, "assets": assets}
}

// serve answers the latest-release route with body (ETag "e1", 304 on a match) and records requests.
func serve(t *testing.T, status int, body any) (*Checker, func() []*http.Request) {
	t.Helper()
	var mu sync.Mutex
	var reqs []*http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		reqs = append(reqs, r)
		mu.Unlock()
		if r.Header.Get("If-None-Match") == `"e1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"e1"`)
		w.WriteHeader(status)
		if s, ok := body.(string); ok {
			_, _ = w.Write([]byte(s))
			return
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return &Checker{API: srv.URL, Repo: Repo, HTTP: srv.Client(), Allow: GitHubAllow}, func() []*http.Request {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(reqs)
	}
}

func TestLatestFound(t *testing.T) {
	c, reqs := serve(t, 200, release("v1.1.0"))
	rel, found, err := c.Latest(context.Background())
	if err != nil || !found {
		t.Fatalf("Latest = %v, %v", found, err)
	}
	want := Release{Version: "1.1.0", DMG: Asset{Name: "Gifsmith-1.1.0-arm64.dmg", Size: 1234, SHA256: digest[len("sha256:"):],
		URL: "https://github.com/scottkw/gifsmith/releases/download/v1.1.0/Gifsmith-1.1.0-arm64.dmg"}}
	if rel != want {
		t.Errorf("rel = %+v\nwant %+v", rel, want)
	}
	r := reqs()[0]
	if r.URL.Path != "/repos/scottkw/gifsmith/releases/latest" || r.URL.RawQuery != "" || r.Method != http.MethodGet {
		t.Errorf("request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
	}
	if r.Header.Get("User-Agent") != "Gifsmith" || r.Header.Get("Accept") != "application/vnd.github+json" {
		t.Errorf("headers %v", r.Header)
	}
	for k := range r.Header {
		switch k {
		case "Accept", "User-Agent", "If-None-Match", "Accept-Encoding":
		default:
			t.Errorf("unexpected header %s: %v", k, r.Header[k])
		}
	}
	// The second check sends the ETag and reuses the cached release on 304.
	rel2, found, err := c.Latest(context.Background())
	if err != nil || !found || rel2 != want {
		t.Fatalf("after 304: %+v, %v, %v", rel2, found, err)
	}
	if got := reqs()[1].Header.Get("If-None-Match"); got != `"e1"` {
		t.Errorf("If-None-Match = %q", got)
	}
}

func TestLatestNotFoundAndErrors(t *testing.T) {
	c, _ := serve(t, 404, `{"message":"Not Found"}`)
	if _, found, err := c.Latest(context.Background()); found || err != nil {
		t.Errorf("404: found %v err %v; want false, nil", found, err)
	}
	c, _ = serve(t, 503, "")
	if _, _, err := c.Latest(context.Background()); err == nil {
		t.Error("503: want an error")
	}
	c, _ = serve(t, 200, `{"tag_name":"v1.1.0","body":"`+strings.Repeat("x", 1<<20)+`"}`)
	if _, _, err := c.Latest(context.Background()); err == nil {
		t.Error("body over 1 MiB: want an error")
	}
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	c = &Checker{API: srv.URL, Repo: Repo, HTTP: http.DefaultClient, Allow: GitHubAllow}
	if _, _, err := c.Latest(context.Background()); err == nil {
		t.Error("closed listener: want an error")
	}
}

func TestLatestNotOffered(t *testing.T) {
	draft := release("v1.1.0")
	draft["draft"] = true
	pre := release("v1.1.0")
	pre["prerelease"] = true
	for name, body := range map[string]any{"draft": draft, "prerelease": pre, "v1.1": release("v1.1"), "latest": release("latest")} {
		c, _ := serve(t, 200, body)
		if _, found, err := c.Latest(context.Background()); found || err != nil {
			t.Errorf("%s: found %v err %v; want false, nil", name, found, err)
		}
	}
}

func TestLatestNotInstallable(t *testing.T) {
	name := DMGAssetName("1.1.0")
	mod := func(k string, v any) map[string]any {
		a := asset(name)
		a[k] = v
		return a
	}
	cases := map[string][]map[string]any{
		"missing":     {asset("Gifsmith-1.1.0-x86_64.dmg")},
		"traversal":   {asset("../" + name)},
		"zip":         {asset(name + ".zip")},
		"starter":     {mod("state", "starter")},
		"size 0":      {mod("size", 0)},
		"too large":   {mod("size", MaxDMGBytes+1)},
		"no digest":   {mod("digest", "")},
		"sha1":        {mod("digest", "sha1:0123456789abcdef0123456789abcdef01234567")},
		"bad hex":     {mod("digest", "sha256:XYZ")},
		"other host":  {mod("browser_download_url", "https://evil.example/"+name)},
		"plain http":  {mod("browser_download_url", "http://github.com/scottkw/gifsmith/releases/download/v1.1.0/"+name)},
		"other repo":  {mod("browser_download_url", "https://github.com/other/repo/releases/download/v1.1.0/"+name)},
		"upper hex":   {mod("digest", strings.ToUpper(digest))},
		"empty asset": {},
	}
	for n, assets := range cases {
		c, _ := serve(t, 200, map[string]any{"tag_name": "v1.1.0", "assets": assets})
		if _, _, err := c.Latest(context.Background()); !errors.Is(err, ErrNotInstallable) {
			t.Errorf("%s: err = %v, want ErrNotInstallable", n, err)
		}
	}
}

func TestGitHubAllow(t *testing.T) {
	for raw, want := range map[string]bool{
		"https://github.com/scottkw/gifsmith/releases/download/v1.1.0/Gifsmith-1.1.0-arm64.dmg": true,
		"https://objects.githubusercontent.com/x":                                               true,
		"https://release-assets.githubusercontent.com/x":                                        true,
		"https://github.com/other/repo/releases/download/v1.1.0/x.dmg":                          false,
		"http://github.com/scottkw/gifsmith/releases/download/v1.1.0/x.dmg":                     false,
		"https://github.com.evil.example/scottkw/gifsmith/releases/download/v1.1.0/x.dmg":       false,
		"https://evil.githubusercontent.com/x":                                                  false,
		"https://github.com/scottkw/gifsmith/releases/download/../../other/repo/x.dmg":          false,
		"https://user@github.com/scottkw/gifsmith/releases/download/v1.1.0/x.dmg":               false,
		"http://objects.githubusercontent.com/x":                                                false,
	} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got := GitHubAllow(u); got != want {
			t.Errorf("GitHubAllow(%s) = %v, want %v", raw, got, want)
		}
	}
}

// One request at a time: Latest serializes on its cache, so concurrent calls never race.
func TestLatestConcurrent(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		w.Header().Set("ETag", `"e1"`)
		if r.Header.Get("If-None-Match") == `"e1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		_ = json.NewEncoder(w).Encode(release("v1.1.0"))
	}))
	defer srv.Close()
	c := &Checker{API: srv.URL, Repo: Repo, HTTP: srv.Client(), Allow: GitHubAllow}
	done := make(chan error)
	for range 4 {
		go func() { _, _, err := c.Latest(context.Background()); done <- err }()
	}
	for range 4 {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}
}
