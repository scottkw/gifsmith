package giphy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
)

func fixedKey(k string) func() (string, bool) { return func() (string, bool) { return k, true } }

// fakeClient points a client at the fixture server's GIPHY routes and counts requests.
func fakeClient(t *testing.T, key string) (*Client, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	h := fakeapi.Handler(fakeapi.Options{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL + "/giphy", Key: fixedKey(key), HTTP: srv.Client()}, &hits
}

// stubClient answers every request with status, headers and body, and records the last query.
func stubClient(t *testing.T, status int, hdr map[string]string, body string) (*Client, *atomic.Int32, func() *url.URL) {
	t.Helper()
	var hits atomic.Int32
	var mu sync.Mutex
	var last *url.URL
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		mu.Lock()
		last = r.URL
		mu.Unlock()
		for k, v := range hdr {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL, Key: fixedKey("uat-key-123"), HTTP: srv.Client()}, &hits, func() *url.URL {
		mu.Lock()
		defer mu.Unlock()
		return last
	}
}

func wantKind(t *testing.T, p provider.Page, err error, kind provider.Kind, status, retry int) {
	t.Helper()
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != kind || pe.Provider != provider.GIPHY || pe.Status != status || pe.RetryAfterSec != retry {
		t.Fatalf("err = %#v, want giphy %s status %d retry %d", err, kind, status, retry)
	}
	if len(p.Items) != 0 {
		t.Errorf("page has %d items alongside an error", len(p.Items))
	}
}

// page builds a GIPHY body with n items and the given total_count (omitted when < 0).
func page(n, total int) string {
	items := make([]string, n)
	for i := range n {
		items[i] = fmt.Sprintf(`{"id":"g%d","slug":"s%d","title":"t%d","images":{"fixed_width":{"url":"https://media.giphy.com/%d/200w.gif","width":200},"original":{"url":"https://media.giphy.com/%d/giphy.gif","width":"480"}}}`, i, i, i, i, i)
	}
	pag := fmt.Sprintf(`{"count":%d,"offset":0}`, n)
	if total >= 0 {
		pag = fmt.Sprintf(`{"total_count":%d,"count":%d,"offset":0}`, total, n)
	}
	return `{"data":[` + strings.Join(items, ",") + `],"pagination":` + pag + `,"meta":{"status":200}}`
}

func TestMapping(t *testing.T) {
	body, err := os.ReadFile("testdata/trending.json")
	if err != nil {
		t.Fatal(err)
	}
	c, _, _ := stubClient(t, 200, nil, string(body))
	p, err := c.Trending(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	want := provider.Gif{
		Provider:    provider.GIPHY,
		ID:          "3o7aD2saalBwwftBIY",
		Title:       "Cat GIF",
		PreviewURL:  "https://media2.giphy.com/media/3o7aD2saalBwwftBIY/200w.gif?cid=abc&rid=200w.gif&ct=g",
		DownloadURL: "https://media2.giphy.com/media/3o7aD2saalBwwftBIY/giphy.gif?cid=abc&rid=giphy.gif&ct=g",
	}
	if len(p.Items) != 1 || p.Items[0] != want {
		t.Fatalf("items = %+v, want [%+v]", p.Items, want)
	}
	if p.Next != "1" {
		t.Errorf("Next = %q, want \"1\" (1 of 5000)", p.Next)
	}
}

func TestTitleFallsBackToSlug(t *testing.T) {
	c, _, _ := stubClient(t, 200, nil, `{"data":[{"id":"a","slug":"cat-a","title":"","images":{"fixed_width":{"url":"https://m/p.gif"},"original":{"url":"https://m/o.gif"}}}],"pagination":{"total_count":1,"count":1,"offset":0}}`)
	p, err := c.Search(context.Background(), "cat", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 1 || p.Items[0].Title != "cat-a" {
		t.Errorf("items = %+v, want title cat-a", p.Items)
	}
}

func TestMissingURLIsHTTPError(t *testing.T) {
	for name, body := range map[string]string{
		"no original": `{"data":[{"id":"a","slug":"a","title":"a","images":{"fixed_width":{"url":"https://m/p.gif"}}}],"pagination":{"total_count":1,"count":1,"offset":0}}`,
		"no preview":  `{"data":[{"id":"a","slug":"a","title":"a","images":{"original":{"url":"https://m/o.gif"}}}],"pagination":{"total_count":1,"count":1,"offset":0}}`,
		"no data":     `{"pagination":{"total_count":1,"count":1,"offset":0},"meta":{"status":200}}`,
	} {
		t.Run(name, func(t *testing.T) {
			c, _, _ := stubClient(t, 200, nil, body)
			p, err := c.Search(context.Background(), "cat", "")
			wantKind(t, p, err, provider.HTTP, 200, 0)
		})
	}
}

func TestFlexInt(t *testing.T) {
	for in, want := range map[string]flexInt{`"200"`: 200, `200`: 200, `""`: 0, `null`: 0} {
		var f flexInt = 7
		if err := json.Unmarshal([]byte(in), &f); err != nil || f != want {
			t.Errorf("%s -> %d, %v; want %d", in, f, err, want)
		}
	}
	var f flexInt
	if err := json.Unmarshal([]byte(`"wide"`), &f); err == nil {
		t.Error(`"wide" decoded without error`)
	}
}

func TestPagingCaps(t *testing.T) {
	for _, tc := range []struct {
		name     string
		trending bool
		cursor   string
		n, total int
		want     string
	}{
		{"search first page", false, "", 24, 5000, "24"},
		{"trending past cap", true, "480", 24, 5000, ""},
		{"search past cap", false, "4980", 24, 5000, ""},
		{"no total, full page", false, "", 24, 0, "24"},
		{"no total, short page", false, "", 10, 0, ""},
		{"total reached", false, "48", 12, 60, ""},
		{"empty page with total", false, "0", 0, 60, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _, _ := stubClient(t, 200, nil, page(tc.n, tc.total))
			var p provider.Page
			var err error
			if tc.trending {
				p, err = c.Trending(context.Background(), tc.cursor)
			} else {
				p, err = c.Search(context.Background(), "cat", tc.cursor)
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(p.Items) != tc.n || p.Next != tc.want {
				t.Errorf("%d items, Next %q; want %d, %q", len(p.Items), p.Next, tc.n, tc.want)
			}
		})
	}
}

func TestBadCursor(t *testing.T) {
	for _, cur := range []string{"x", "-24"} {
		c, hits, _ := stubClient(t, 200, nil, page(1, 1))
		p, err := c.Search(context.Background(), "cat", cur)
		wantKind(t, p, err, provider.HTTP, 0, 0)
		if n := hits.Load(); n != 0 {
			t.Errorf("cursor %q made %d requests, want 0", cur, n)
		}
	}
}

func TestQueryTruncation(t *testing.T) {
	c, _, last := stubClient(t, 200, nil, page(1, 1))
	q := strings.Repeat("é", 30) + strings.Repeat("a", 30) // 60 runes, 90 bytes
	if _, err := c.Search(context.Background(), q, ""); err != nil {
		t.Fatal(err)
	}
	if last() == nil {
		t.Fatal("no request made")
	}
	if got, want := last().Query().Get("q"), strings.Repeat("é", 30)+strings.Repeat("a", 20); got != want {
		t.Errorf("q sent = %q (%d runes), want first 50 runes", got, len([]rune(got)))
	}
}

func TestRequestParams(t *testing.T) {
	c, _, last := stubClient(t, 200, nil, page(1, 1))
	for _, tc := range []struct {
		path string
		call func() error
	}{
		{"/v1/gifs/search", func() error { _, err := c.Search(context.Background(), "cat", "24"); return err }},
		{"/v1/gifs/trending", func() error { _, err := c.Trending(context.Background(), "24"); return err }},
	} {
		if err := tc.call(); err != nil {
			t.Fatal(err)
		}
		u := last()
		if u == nil {
			t.Fatalf("%s: no request made", tc.path)
		}
		v := u.Query()
		if u.Path != tc.path {
			t.Errorf("path = %s, want %s", u.Path, tc.path)
		}
		for k, want := range map[string]string{"api_key": "uat-key-123", "limit": "24", "offset": "24", "rating": "pg-13", "bundle": "messaging_non_clips"} {
			if v.Get(k) != want {
				t.Errorf("%s: %s = %q, want %q", tc.path, k, v.Get(k), want)
			}
		}
		for _, k := range []string{"lang", "random_id"} {
			if v.Has(k) {
				t.Errorf("%s: unexpected param %s", tc.path, k)
			}
		}
	}
	if q := last().Query(); q.Has("q") {
		t.Errorf("trending sent q=%q", q.Get("q"))
	}
}

func TestKeyMissing(t *testing.T) {
	c, hits, _ := stubClient(t, 200, nil, page(1, 1))
	c.Key = func() (string, bool) { return "", false }
	p, err := c.Trending(context.Background(), "")
	wantKind(t, p, err, provider.KeyMissing, 0, 0)
	p, err = c.Search(context.Background(), "cat", "")
	wantKind(t, p, err, provider.KeyMissing, 0, 0)
	if n := hits.Load(); n != 0 {
		t.Errorf("made %d requests without a key, want 0", n)
	}
}

func TestKeyRejected(t *testing.T) {
	c, _ := fakeClient(t, "bad-key-123")
	p, err := c.Trending(context.Background(), "")
	wantKind(t, p, err, provider.KeyRejected, 401, 0)

	body, err := os.ReadFile("testdata/unauthorized.json")
	if err != nil {
		t.Fatal(err)
	}
	s, _, _ := stubClient(t, 401, map[string]string{"Retry-After": "0"}, string(body))
	p, err = s.Search(context.Background(), "cat", "")
	wantKind(t, p, err, provider.KeyRejected, 401, 0)

	s, _, _ = stubClient(t, 403, nil, `{"data":[],"meta":{"status":403}}`)
	p, err = s.Search(context.Background(), "cat", "")
	wantKind(t, p, err, provider.KeyRejected, 403, 0)
}

func TestFixtureStatuses(t *testing.T) {
	c, _ := fakeClient(t, "uat-key-123")
	p, err := c.Search(context.Background(), "ratelimit", "")
	wantKind(t, p, err, provider.RateLimited, 0, 120)
	p, err = c.Search(context.Background(), "boom", "")
	wantKind(t, p, err, provider.HTTP, 500, 0)
	p, err = c.Search(context.Background(), "revoked", "")
	wantKind(t, p, err, provider.KeyRejected, 401, 0)
}

func TestFixturePaging(t *testing.T) {
	c, _ := fakeClient(t, "uat-key-123")
	p, err := c.Trending(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 24 || p.Next != "24" {
		t.Fatalf("first page: %d items, Next %q; want 24, \"24\"", len(p.Items), p.Next)
	}
	g := p.Items[0]
	if !strings.HasSuffix(g.PreviewURL, "/media/giphy-1-200w.gif") || !strings.HasSuffix(g.DownloadURL, "/media/giphy-1.gif") {
		t.Errorf("item 0 = %+v, want fixture media URLs", g)
	}
	p, err = c.Search(context.Background(), "cats", "48")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 12 || p.Next != "" {
		t.Errorf("last page: %d items, Next %q; want 12, \"\"", len(p.Items), p.Next)
	}
	p, err = c.Search(context.Background(), "empty", "")
	if err != nil || p.Items == nil || len(p.Items) != 0 || p.Next != "" {
		t.Errorf("empty: %+v, %v; want non-nil empty page", p, err)
	}
}

func TestRedaction(t *testing.T) {
	const secret = "SECRETKEY123"
	refused := &Client{BaseURL: "http://127.0.0.1:1", Key: fixedKey(secret), HTTP: &http.Client{Timeout: 2 * time.Second}}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(slow.Close)
	timeout := &Client{BaseURL: slow.URL, Key: fixedKey(secret), HTTP: &http.Client{Timeout: 50 * time.Millisecond}}
	for name, c := range map[string]*Client{"refused": refused, "timeout": timeout} {
		_, err := c.Search(context.Background(), "cat", "")
		if err == nil {
			t.Fatalf("%s: no error", name)
		}
		if strings.Contains(err.Error(), secret) || strings.Contains(string(provider.MarshalError(err)), secret) {
			t.Errorf("%s: key leaked: %q / %s", name, err.Error(), provider.MarshalError(err))
		}
		if provider.MarshalError(err) == nil {
			t.Errorf("%s: %v is not a typed provider error", name, err)
		}
	}
}

func clipsBody(withPreview bool) string {
	fw := `"fixed_width":{"url":"https://media.giphy.com/c1/200w.gif"},`
	if !withPreview {
		fw = ""
	}
	return `{"data":[{"id":"c1","slug":"clip-c1","title":"A clip","type":"video","images":{` + fw + `"fixed_height":{"url":"https://m/fh.gif"}},"video":{"assets":{"360p":{"url":"https://m/c1.mp4"}}}}],"pagination":{"total_count":1,"count":1,"offset":0}}`
}

func TestClipsMode(t *testing.T) {
	c, _, last := stubClient(t, 200, nil, clipsBody(true))
	c.Clips = true
	for _, tc := range []struct {
		path string
		call func() (provider.Page, error)
	}{
		{"/v1/clips/search", func() (provider.Page, error) { return c.Search(context.Background(), "cat", "") }},
		{"/v1/clips/trending", func() (provider.Page, error) { return c.Trending(context.Background(), "") }},
	} {
		p, err := tc.call()
		if err != nil {
			t.Fatalf("%s: %v", tc.path, err)
		}
		want := provider.Gif{Provider: provider.Clips, ID: "c1", Title: "A clip", PreviewURL: "https://media.giphy.com/c1/200w.gif", IsClip: true}
		if len(p.Items) != 1 || p.Items[0] != want {
			t.Errorf("%s: items = %+v, want [%+v]", tc.path, p.Items, want)
		}
		u := last()
		if u == nil || u.Path != tc.path {
			t.Fatalf("request = %v, want path %s", u, tc.path)
		}
		for k, want := range map[string]string{"rating": "pg-13", "bundle": "clips_grid_picker", "api_key": "uat-key-123"} {
			if got := u.Query().Get(k); got != want {
				t.Errorf("%s: %s = %q, want %q", tc.path, k, got, want)
			}
		}
	}

	s, _, _ := stubClient(t, 200, nil, clipsBody(false))
	s.Clips = true
	p, err := s.Search(context.Background(), "cat", "")
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != provider.HTTP || pe.Provider != provider.Clips || len(p.Items) != 0 {
		t.Errorf("clip without fixed_width: %+v, %#v; want clips http error", p, err)
	}
}

func probeKind(t *testing.T, err error, kind provider.Kind, status int) {
	t.Helper()
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != kind || pe.Provider != provider.Clips || pe.Status != status {
		t.Fatalf("ProbeClips err = %#v, want clips %s status %d", err, kind, status)
	}
}

func TestProbeClips(t *testing.T) {
	ok, _, last := stubClient(t, 200, nil, clipsBody(true))
	if err := ok.ProbeClips(context.Background()); err != nil {
		t.Fatalf("200: err = %v, want nil", err)
	}
	if u := last(); u == nil || u.Path != "/v1/clips/trending" || u.Query().Get("limit") != "1" || u.Query().Get("rating") != "pg-13" {
		t.Errorf("probe request = %v, want /v1/clips/trending limit=1 rating=pg-13", u)
	}

	fixture, _ := fakeClient(t, "uat-key-123")
	if err := fixture.ProbeClips(context.Background()); err != nil {
		t.Errorf("fixture: err = %v, want nil", err)
	}
	srv := httptest.NewServer(fakeapi.Handler(fakeapi.Options{ClipsDenied: true}))
	t.Cleanup(srv.Close)
	denied := &Client{BaseURL: srv.URL + "/giphy", Key: fixedKey("uat-key-123"), HTTP: srv.Client()}
	probeKind(t, denied.ProbeClips(context.Background()), provider.KeyRejected, 403)

	for _, tc := range []struct {
		status int
		kind   provider.Kind
		want   int
	}{
		{401, provider.KeyRejected, 401},
		{404, provider.HTTP, 404},
		{429, provider.RateLimited, 0},
	} {
		s, _, _ := stubClient(t, tc.status, nil, `{"data":[],"meta":{}}`)
		probeKind(t, s.ProbeClips(context.Background()), tc.kind, tc.want)
	}
}
