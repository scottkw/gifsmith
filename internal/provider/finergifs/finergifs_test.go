package finergifs

import (
	"context"
	"net/http"
	"net/http/httptest"
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
