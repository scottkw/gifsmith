package finergifs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
)

func newClient(t *testing.T, o fakeapi.Options) (*Client, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	h := fakeapi.Handler(o)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL + "/finer", MediaBase: srv.URL + "/media", HTTP: srv.Client()}, &hits
}

func TestSearchPaging(t *testing.T) {
	c, _ := newClient(t, fakeapi.Options{})
	p, err := c.Search(context.Background(), "dwight", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 24 || p.Next != "24" {
		t.Fatalf("first page: %d items, Next %q; want 24, \"24\"", len(p.Items), p.Next)
	}
	g := p.Items[0]
	want := c.MediaBase + "/office-1.gif"
	if g.Provider != provider.FinerGifs || g.ID != "office-1" || g.Title != "Office moment 1" || g.PreviewURL != want || g.DownloadURL != want {
		t.Errorf("item 0 = %+v, want office-1 with both URLs %s", g, want)
	}
	p, err = c.Search(context.Background(), "dwight", "48")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 12 || p.Next != "" {
		t.Errorf("last page: %d items, Next %q; want 12, \"\"", len(p.Items), p.Next)
	}
	if !strings.HasSuffix(p.Items[0].DownloadURL, "/office-49.gif") {
		t.Errorf("last page starts at %s, want office-49", p.Items[0].DownloadURL)
	}
}

func TestSearchEmpty(t *testing.T) {
	c, _ := newClient(t, fakeapi.Options{})
	p, err := c.Search(context.Background(), "empty", "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Items == nil || len(p.Items) != 0 || p.Next != "" {
		t.Errorf("page = %+v, want non-nil empty items and no Next", p)
	}
}

// bodyClient serves body for every request and counts requests.
func bodyClient(t *testing.T, body string) (*Client, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL, MediaBase: "https://media.thefinergifs.club", HTTP: srv.Client()}, &hits
}

func wantHTTPErr(t *testing.T, p provider.Page, err error, status int) {
	t.Helper()
	var pe *provider.Error
	if !errors.As(err, &pe) || pe.Kind != provider.HTTP || pe.Provider != provider.FinerGifs || pe.Status != status {
		t.Fatalf("err = %v, want finergifs http error with status %d", err, status)
	}
	if len(p.Items) != 0 {
		t.Errorf("page has %d items alongside an error", len(p.Items))
	}
}

func TestMapperFixture(t *testing.T) {
	body, err := os.ReadFile("testdata/search.json")
	if err != nil {
		t.Fatal(err)
	}
	c, _ := bodyClient(t, string(body))
	p, err := c.Search(context.Background(), "dwight", "")
	if err != nil {
		t.Fatal(err)
	}
	const m = "https://media.thefinergifs.club/"
	want := []provider.Gif{
		{Provider: provider.FinerGifs, ID: "02x05-dwight-bears", Title: "Bears. Beets. Battlestar Galactica.", PreviewURL: m + "02x05-dwight-bears.gif", DownloadURL: m + "02x05-dwight-bears.gif"},
		{Provider: provider.FinerGifs, ID: "03x01-michael-no", Title: "03x01-michael-no", PreviewURL: m + "03x01-michael-no.gif", DownloadURL: m + "03x01-michael-no.gif"},
		// The mapper keeps duplicates in returned order; dedupe policy lives in the search service.
		{Provider: provider.FinerGifs, ID: "02x05-dwight-bears", Title: "Bears. Beets. Battlestar Galactica.", PreviewURL: m + "02x05-dwight-bears.gif", DownloadURL: m + "02x05-dwight-bears.gif"},
	}
	if !reflect.DeepEqual(p.Items, want) { // also: Finer items carry none of the optional metadata
		t.Errorf("items =\n%+v\nwant\n%+v", p.Items, want)
	}
	if p.Next != "3" {
		t.Errorf("Next = %q, want \"3\" (3 of 57 found)", p.Next)
	}
}

func TestMapperRejectsBadFileid(t *testing.T) {
	c, _ := bodyClient(t, `{"hits":{"found":2,"start":0},"results":[{"fields":{"name":"ok","text":"ok","fileid":"ok-1"}},{"fields":{"name":"x","text":"x","fileid":"../x"}}]}`)
	p, err := c.Search(context.Background(), "x", "")
	wantHTTPErr(t, p, err, 200)
}

func TestMapperMissingResults(t *testing.T) {
	c, _ := bodyClient(t, `{"hits":{"found":1,"start":0}}`)
	p, err := c.Search(context.Background(), "x", "")
	wantHTTPErr(t, p, err, 200)
}

func TestMapperBadCursor(t *testing.T) {
	for _, cur := range []string{"abc", "-24"} {
		c, hits := bodyClient(t, `{"results":[]}`)
		p, err := c.Search(context.Background(), "x", cur)
		wantHTTPErr(t, p, err, 0)
		if n := hits.Load(); n != 0 {
			t.Errorf("cursor %q made %d requests, want 0", cur, n)
		}
	}
}

// WR-02: a long subtitle is clamped to the library title limit.
func TestMapperLongTitleClamped(t *testing.T) {
	text := strings.Repeat("ü", provider.MaxTitleRunes+50)
	c, _ := bodyClient(t, `{"results":[{"fields":{"name":"n","text":"`+text+`","fileid":"f1"}}]}`)
	p, err := c.Search(context.Background(), "x", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 1 || p.Items[0].Title != string([]rune(text)[:provider.MaxTitleRunes]) {
		t.Errorf("items = %+v, want the title clamped", p.Items)
	}
}

func TestMapperNoHitsFullPage(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"results":[`)
	for i := range 24 {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"fields":{"name":"n%d","text":"t%d","fileid":"f%d"}}`, i, i, i)
	}
	b.WriteString(`]}`)
	c, _ := bodyClient(t, b.String())
	p, err := c.Search(context.Background(), "x", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 24 || p.Next != "24" {
		t.Errorf("%d items, Next %q; want 24, \"24\"", len(p.Items), p.Next)
	}
}

func TestFinerDown(t *testing.T) {
	c, _ := newClient(t, fakeapi.Options{FinerDown: true})
	p, err := c.Search(context.Background(), "dwight", "")
	wantHTTPErr(t, p, err, 530)
}

func TestTrendingMakesNoRequest(t *testing.T) {
	c, hits := newClient(t, fakeapi.Options{})
	p, err := c.Trending(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Items == nil || len(p.Items) != 0 || p.Next != "" {
		t.Errorf("page = %+v, want non-nil empty items and no Next", p)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("Trending made %d requests, want 0", n)
	}
}

func TestPageSizeMore(t *testing.T) {
	c, _ := newClient(t, fakeapi.Options{})
	c.Params = func() provider.Params { p := provider.DefaultParams(); p.PageSize = 36; return p }
	p, err := c.Search(context.Background(), "dwight", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 36 || p.Next != "36" {
		t.Errorf("size 36: %d items, Next %q; want 36, \"36\"", len(p.Items), p.Next)
	}
	var b strings.Builder
	b.WriteString(`{"results":[`)
	for i := range 36 {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"fields":{"name":"n%d","text":"t%d","fileid":"f%d"}}`, i, i, i)
	}
	b.WriteString(`]}`)
	c2, _ := bodyClient(t, b.String())
	c2.Params = c.Params
	if p, err := c2.Search(context.Background(), "x", ""); err != nil || p.Next != "36" {
		t.Errorf("36 items without hits at size 36: Next %q, %v; want \"36\"", p.Next, err)
	}
}
