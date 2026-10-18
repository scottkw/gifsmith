package klipy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
)

func fixedKey(k string) func() (string, bool) { return func() (string, bool) { return k, true } }

func fakeClient(t *testing.T, key string) (*Client, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	h := fakeapi.Handler(fakeapi.Options{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL + "/klipy", Key: fixedKey(key), HTTP: srv.Client()}, &hits
}

// stubClient answers every request with status and body and records the last request URI.
func stubClient(t *testing.T, status int, body string) (*Client, *atomic.Int32, func() string) {
	t.Helper()
	var hits atomic.Int32
	var mu sync.Mutex
	var last string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		mu.Lock()
		last = r.RequestURI
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL, Key: fixedKey("uat-key-123"), HTTP: srv.Client()}, &hits, func() string {
		mu.Lock()
		defer mu.Unlock()
		return last
	}
}

func wantKind(t *testing.T, p provider.Page, err error, kind provider.Kind, status, retry int) {
	t.Helper()
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != kind || pe.Provider != provider.Klipy || pe.Status != status || pe.RetryAfterSec != retry {
		t.Fatalf("err = %#v, want klipy %s status %d retry %d", err, kind, status, retry)
	}
	if len(p.Items) != 0 {
		t.Errorf("page has %d items alongside an error", len(p.Items))
	}
}

// item is one Klipy gif item; empty sm/hd/md leave that size out.
func item(slug, sm, md, hd string) string {
	var sizes []string
	for name, u := range map[string]string{"sm": sm, "md": md, "hd": hd} {
		if u != "" {
			sizes = append(sizes, fmt.Sprintf(`%q:{"gif":{"url":%q,"width":220}}`, name, u))
		}
	}
	return fmt.Sprintf(`{"id":8041071659142944,"slug":%q,"title":"","type":"gif","file":{%s}}`, slug, strings.Join(sizes, ","))
}

func body(page int, hasNext bool, items ...string) string {
	return fmt.Sprintf(`{"result":true,"data":{"data":[%s],"current_page":%d,"per_page":24,"has_next":%t}}`, strings.Join(items, ","), page, hasNext)
}

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

func TestMapping(t *testing.T) {
	b, err := os.ReadFile("testdata/search.json")
	if err != nil {
		t.Fatal(err)
	}
	logs := captureLog(t)
	c, _, _ := stubClient(t, 200, string(b))
	p, err := c.Search(context.Background(), "hello", "")
	if err != nil {
		t.Fatal(err)
	}
	want := provider.Gif{
		Provider:    provider.Klipy,
		ID:          "hello-hi-662",
		Title:       "Hello",
		PreviewURL:  "https://static.klipy.com/ii/935d7ab9d8c6202580a668421940ec81/14/af/y6iepZM7.gif",
		DownloadURL: "https://static.klipy.com/ii/935d7ab9d8c6202580a668421940ec81/14/af/um0L4dFH.gif",
		Width:       498,
		Height:      498,
		Size:        4001918,
	}
	ad := provider.Gif{Provider: provider.Klipy, ID: "sponsored-1", Title: "Sponsored", PreviewURL: "https://ads.example.test/sm.gif", DownloadURL: "https://ads.example.test/hd.gif"}
	if len(p.Items) != 2 || !reflect.DeepEqual(p.Items[0], want) || !reflect.DeepEqual(p.Items[1], ad) {
		t.Fatalf("items = %+v, want [%+v %+v] (an ad with GIF renditions stays in place; one without is skipped)", p.Items, want, ad)
	}
	if p.Next != "2" {
		t.Errorf("Next = %q, want \"2\"", p.Next)
	}
	if !strings.Contains(logs.String(), "klipy: skipped 1 non-gif items") {
		t.Errorf("log = %q, want the skipped-count line", logs.String())
	}
	if strings.Contains(logs.String(), "http") {
		t.Errorf("log line carries a URL: %q", logs.String())
	}
}

func TestTitleFallsBackToSlug(t *testing.T) {
	c, _, _ := stubClient(t, 200, body(1, false, item("cat-1", "https://s/sm.gif", "", "https://s/hd.gif")))
	p, err := c.Search(context.Background(), "cat", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 1 || p.Items[0].Title != "cat-1" {
		t.Errorf("items = %+v, want title cat-1", p.Items)
	}
}

// WR-02: a long slug (the id, and the title when there is none) is clamped to the library limits.
func TestLongSlugClamped(t *testing.T) {
	slug := strings.Repeat("é", provider.MaxTitleRunes+50)
	c, _, _ := stubClient(t, 200, body(1, false, item(slug, "https://s/sm.gif", "", "https://s/hd.gif")))
	p, err := c.Search(context.Background(), "cat", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 1 || p.Items[0].ID != slug[:provider.MaxIDBytes] || p.Items[0].Title != string([]rune(slug)[:provider.MaxTitleRunes]) {
		t.Errorf("items = %+v, want id and title clamped", p.Items)
	}
}

func TestMDFallback(t *testing.T) {
	c, _, _ := stubClient(t, 200, body(1, false, item("a", "https://s/sm.gif", "https://s/md.gif", "")))
	p, err := c.Search(context.Background(), "cat", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 1 || p.Items[0].DownloadURL != "https://s/md.gif" {
		t.Errorf("items = %+v, want md.gif download", p.Items)
	}
}

func TestMissingURLIsHTTPError(t *testing.T) {
	for name, b := range map[string]string{
		"no hd or md":  body(1, false, item("ok", "https://s/sm.gif", "", "https://s/hd.gif"), item("a", "https://s/sm.gif", "", "")),
		"no sm":        body(1, false, item("a", "", "", "https://s/hd.gif")),
		"result false": `{"result":false,"data":{"data":[],"current_page":1,"has_next":false}}`,
		"no data":      `{"result":true}`,
		"no inner":     `{"result":true,"data":{"current_page":1,"has_next":false}}`,
	} {
		t.Run(name, func(t *testing.T) {
			c, _, _ := stubClient(t, 200, b)
			p, err := c.Search(context.Background(), "cat", "")
			wantKind(t, p, err, provider.HTTP, 200, 0)
		})
	}
}

func TestNoDedupe(t *testing.T) {
	c, _, _ := stubClient(t, 200, body(1, false,
		item("dup", "https://s/1.gif", "", "https://s/1h.gif"),
		item("other", "https://s/2.gif", "", "https://s/2h.gif"),
		item("dup", "https://s/3.gif", "", "https://s/3h.gif")))
	p, err := c.Trending(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, g := range p.Items {
		got = append(got, g.ID+" "+g.PreviewURL)
	}
	want := []string{"dup https://s/1.gif", "other https://s/2.gif", "dup https://s/3.gif"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("items = %v, want %v (order and duplicates kept)", got, want)
	}
}

func TestPaging(t *testing.T) {
	one := item("a", "https://s/sm.gif", "", "https://s/hd.gif")
	c, _, last := stubClient(t, 200, body(2, true, one))
	p, err := c.Search(context.Background(), "cat", "2")
	if err != nil {
		t.Fatal(err)
	}
	if p.Next != "3" {
		t.Errorf("has_next, page 2: Next = %q, want \"3\"", p.Next)
	}
	if !strings.Contains(last(), "page=2") {
		t.Errorf("request %s, want page=2", last())
	}

	c, _, last = stubClient(t, 200, body(1, false, one))
	p, err = c.Search(context.Background(), "cat", "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Next != "" {
		t.Errorf("no has_next: Next = %q, want \"\"", p.Next)
	}
	if !strings.Contains(last(), "page=1") {
		t.Errorf("empty cursor sent %s, want page=1", last())
	}

	// A page with nothing renderable ends paging: an empty page with a cursor makes the grid fetch again at once.
	c, _, _ = stubClient(t, 200, `{"result":true,"data":{"data":[{"slug":"ad","type":"ad","content":"<div></div>"}],"current_page":1,"has_next":true}}`)
	p, err = c.Search(context.Background(), "cat", "")
	if err != nil || len(p.Items) != 0 || p.Next != "" {
		t.Errorf("page of only unrenderable items = %+v, %v; want no items and no Next", p, err)
	}

	for _, cur := range []string{"x", "0", "-1"} {
		c, hits, _ := stubClient(t, 200, body(1, false, one))
		p, err := c.Search(context.Background(), "cat", cur)
		wantKind(t, p, err, provider.HTTP, 0, 0)
		if n := hits.Load(); n != 0 {
			t.Errorf("cursor %q made %d requests, want 0", cur, n)
		}
	}
}

func TestRequest(t *testing.T) {
	c, _, last := stubClient(t, 200, body(1, false))
	c.Key = fixedKey("abc;def12")
	if _, err := c.Search(context.Background(), "hi there", ""); err != nil {
		t.Fatal(err)
	}
	uri := last()
	if !strings.HasPrefix(uri, "/api/v1/abc%3Bdef12/gifs/search?") {
		t.Errorf("search URI = %s, want the path-escaped key", uri)
	}
	for _, want := range []string{"q=hi+there", "per_page=24", "content_filter=medium", "page=1"} {
		if !strings.Contains(uri, want) {
			t.Errorf("search URI %s lacks %s", uri, want)
		}
	}
	if _, err := c.Trending(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	uri = last()
	if !strings.HasPrefix(uri, "/api/v1/abc%3Bdef12/gifs/trending?") || strings.Contains(uri, "q=") {
		t.Errorf("trending URI = %s", uri)
	}
	for _, banned := range []string{"customer_id", "locale"} {
		if strings.Contains(uri, banned) {
			t.Errorf("trending URI %s sends %s", uri, banned)
		}
	}
}

func TestKeyValidation(t *testing.T) {
	for _, k := range []string{"a/b", "a b", "a?b", "a#b", "a\x01b", "a\tb"} {
		c, hits, _ := stubClient(t, 200, body(1, false))
		c.Key = fixedKey(k)
		p, err := c.Trending(context.Background(), "")
		wantKind(t, p, err, provider.KeyRejected, 0, 0)
		if n := hits.Load(); n != 0 {
			t.Errorf("key %q made %d requests, want 0", k, n)
		}
	}
	c, hits, _ := stubClient(t, 200, body(1, false))
	c.Key = func() (string, bool) { return "", false }
	p, err := c.Search(context.Background(), "cat", "")
	wantKind(t, p, err, provider.KeyMissing, 0, 0)
	if n := hits.Load(); n != 0 {
		t.Errorf("made %d requests without a key, want 0", n)
	}
}

func TestKeyRejected(t *testing.T) {
	c, _ := fakeClient(t, "bad-key-123")
	p, err := c.Trending(context.Background(), "")
	wantKind(t, p, err, provider.KeyRejected, 404, 0)

	b, err := os.ReadFile("testdata/invalid_key.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, status := range []int{401, 403, 404} {
		s, _, _ := stubClient(t, status, string(b))
		p, err := s.Search(context.Background(), "cat", "")
		wantKind(t, p, err, provider.KeyRejected, status, 0)
	}
}

func TestFixtureStatuses(t *testing.T) {
	c, _ := fakeClient(t, "uat-key-123")
	p, err := c.Search(context.Background(), "ratelimit", "")
	wantKind(t, p, err, provider.RateLimited, 0, 120)
	p, err = c.Search(context.Background(), "boom", "")
	wantKind(t, p, err, provider.HTTP, 500, 0)
	p, err = c.Search(context.Background(), "revoked", "")
	wantKind(t, p, err, provider.KeyRejected, 404, 0)
}

func TestFixturePaging(t *testing.T) {
	c, _ := fakeClient(t, "uat-key-123")
	p, err := c.Trending(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 24 || p.Next != "2" {
		t.Fatalf("first page: %d items, Next %q; want 24, \"2\"", len(p.Items), p.Next)
	}
	dups := 0
	for _, g := range p.Items {
		if g.ID == "klipy-dup" {
			dups++
		}
	}
	if dups != 2 {
		t.Errorf("klipy-dup appears %d times on page 1, want 2", dups)
	}
	if g := p.Items[0]; !strings.HasSuffix(g.PreviewURL, "/media/klipy-1-sm.gif") || !strings.HasSuffix(g.DownloadURL, "/media/klipy-1-hd.gif") {
		t.Errorf("item 0 = %+v, want fixture media URLs", g)
	}
	p, err = c.Search(context.Background(), "cats", "3")
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

// tagged is one Klipy item with the given raw tags value, hd and md sizes.
func tagged(t *testing.T, tags string) provider.Gif {
	t.Helper()
	it := `{"slug":"a","title":"A","type":"gif","tags":` + tags + `,"file":{"sm":{"gif":{"url":"https://s/sm.gif"}},` +
		`"md":{"gif":{"url":"https://s/md.gif","width":"200","height":150,"size":"bad"}},` +
		`"hd":{"gif":{"url":"https://s/hd.gif","width":498,"height":"300","size":4001918}}}}`
	c, _, _ := stubClient(t, 200, body(1, false, it))
	p, err := c.Search(context.Background(), "cat", "")
	if err != nil || len(p.Items) != 1 {
		t.Fatalf("tags %s: %+v, %v", tags, p, err)
	}
	return p.Items[0]
}

func TestTagsTolerant(t *testing.T) {
	long := strings.Repeat("é", 150)
	many := make([]string, 60)
	for i := range many {
		many[i] = fmt.Sprintf("%q", fmt.Sprint("t", i))
	}
	for tags, want := range map[string][]string{
		`["funny",{"name":"x"},"",5,"cat"]`: {"funny", "cat"},
		`[]`:                                nil,
		`"funny"`:                           nil,
		`{"a":1}`:                           nil,
		`null`:                              nil,
		`["` + long + `"]`:                  {strings.Repeat("é", 100)},
	} {
		if got := tagged(t, tags).Tags; !reflect.DeepEqual(got, want) {
			t.Errorf("tags %.40s -> %q, want %q", tags, got, want)
		}
	}
	if got := tagged(t, "["+strings.Join(many, ",")+"]").Tags; len(got) != 50 || got[49] != "t49" {
		t.Errorf("60 tags -> %d kept, want the first 50", len(got))
	}
}

func TestDimensionsFollowDownload(t *testing.T) {
	g := tagged(t, `[]`)
	if g.Width != 498 || g.Height != 300 || g.Size != 4001918 || g.MediumURL != "https://s/md.gif" {
		t.Errorf("hd item = %+v, want hd dims (strings tolerated) and md medium", g)
	}
	c, _, _ := stubClient(t, 200, body(1, false, `{"slug":"b","type":"gif","file":{"sm":{"gif":{"url":"https://s/sm.gif"}},"md":{"gif":{"url":"https://s/md.gif","width":200,"height":150,"size":"bad"}}}}`))
	p, err := c.Search(context.Background(), "cat", "")
	if err != nil || len(p.Items) != 1 {
		t.Fatal(p, err)
	}
	if g := p.Items[0]; g.DownloadURL != "https://s/md.gif" || g.Width != 200 || g.Height != 150 || g.Size != 0 || g.PageURL != "" {
		t.Errorf("md item = %+v, want md dims, unknown size, no page URL", g)
	}
}
