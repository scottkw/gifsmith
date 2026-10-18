package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gifsmith/internal/provider"
)

const host = "media.test"

// allow stands in for the staging allowlist: https on media.test only.
func allow(_ provider.ID, u *url.URL) bool { return u.Scheme == "https" && u.Host == host }

func gifN(p provider.ID, id string) provider.Gif {
	return provider.Gif{Provider: p, ID: id, Title: "Moment " + id,
		PreviewURL: "https://" + host + "/" + id + "-200w.gif", DownloadURL: "https://" + host + "/" + id + ".gif"}
}

// newStore keeps every file under t.TempDir(): tests never touch Application Support.
func newStore(t *testing.T) *Store {
	t.Helper()
	var ms int64 = 1_000
	return &Store{Dir: t.TempDir(), Allow: allow, Now: func() time.Time { ms++; return time.UnixMilli(ms) }}
}

func ids(es []Entry) string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Gif.ID
	}
	return strings.Join(out, ",")
}

func TestTouchMovesToTop(t *testing.T) {
	s := newStore(t)
	for _, id := range []string{"a", "b"} {
		if err := s.Touch(gifN(provider.GIPHY, id)); err != nil {
			t.Fatal(err)
		}
	}
	if got := ids(s.Snapshot().Recents); got != "b,a" {
		t.Fatalf("recents = %s, want b,a", got)
	}
	if err := s.Touch(gifN(provider.GIPHY, "a")); err != nil {
		t.Fatal(err)
	}
	snap := s.Snapshot()
	if got := ids(snap.Recents); got != "a,b" {
		t.Fatalf("recents = %s, want a,b (moved, not duplicated)", got)
	}
	if snap.Recents[0].At <= snap.Recents[1].At {
		t.Errorf("At %d not newer than %d", snap.Recents[0].At, snap.Recents[1].At)
	}
	// Same id under another provider is another item.
	if err := s.Touch(gifN(provider.Klipy, "a")); err != nil {
		t.Fatal(err)
	}
	if n := len(s.Snapshot().Recents); n != 3 {
		t.Errorf("%d recents, want 3 (keys are provider:id)", n)
	}
}

func TestTouchCap50(t *testing.T) {
	s := newStore(t)
	for i := 0; i < 51; i++ {
		if err := s.Touch(gifN(provider.GIPHY, fmt.Sprint(i))); err != nil {
			t.Fatal(err)
		}
	}
	r := s.Snapshot().Recents
	if len(r) != RecentsMax {
		t.Fatalf("%d recents, want %d", len(r), RecentsMax)
	}
	if r[0].Gif.ID != "50" || r[len(r)-1].Gif.ID != "1" {
		t.Errorf("%d recents from %s to %s, want 50 from 50 to 1 (oldest dropped)", len(r), r[0].Gif.ID, r[len(r)-1].Gif.ID)
	}
}

func TestRecentsPersistAcrossStores(t *testing.T) {
	s := newStore(t)
	g := gifN(provider.FinerGifs, "office-1")
	g.Tags = []string{"office"}
	if err := s.Touch(g); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.Dir, "recents.json")
	fi, err := os.Stat(path)
	if err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("recents.json: %v, mode %v; want 0600", err, fi)
	}
	var f struct {
		Version int     `json:"version"`
		Items   []Entry `json:"items"`
	}
	b, _ := os.ReadFile(path)
	if err := json.Unmarshal(b, &f); err != nil || f.Version != 1 || len(f.Items) != 1 {
		t.Fatalf("recents.json = %s, %v", b, err)
	}
	got := (&Store{Dir: s.Dir, Allow: allow}).Snapshot().Recents
	if len(got) != 1 || got[0].Gif.Title != g.Title || got[0].Gif.Tags[0] != "office" {
		t.Errorf("new store reads %+v, want the touched gif", got)
	}
}

func TestSnapshotListsNeverNil(t *testing.T) {
	snap := newStore(t).Snapshot()
	b, _ := json.Marshal(snap)
	if snap.Favorites == nil || snap.Recents == nil || snap.Searches == nil || snap.Unreadable == nil {
		t.Errorf("snapshot %s has a null list", b)
	}
}

func TestValidate(t *testing.T) {
	ok := gifN(provider.GIPHY, "a")
	ok.PageURL, ok.SourceURL = "https://"+host+"/gifs/a", "https://example.org/x"
	if err := Validate(ok, allow); err != nil {
		t.Fatalf("valid gif: %v", err)
	}
	long := strings.Repeat("é", 501)
	bad := map[string]func(g *provider.Gif){
		"unknown provider": func(g *provider.Gif) { g.Provider = "tenor" },
		"empty id":         func(g *provider.Gif) { g.ID = "" },
		"long id":          func(g *provider.Gif) { g.ID = strings.Repeat("a", 201) },
		"long title":       func(g *provider.Gif) { g.Title = long },
		"no preview":       func(g *provider.Gif) { g.PreviewURL = "" },
		"no download":      func(g *provider.Gif) { g.DownloadURL = "" },
		"download host":    func(g *provider.Gif) { g.DownloadURL = "https://evil.test/a.gif" },
		"preview scheme":   func(g *provider.Gif) { g.PreviewURL = "javascript:alert(1)" },
		"medium host":      func(g *provider.Gif) { g.MediumURL = "https://evil.test/m.gif" },
		"gif host":         func(g *provider.Gif) { g.GifURL = "https://evil.test/g.gif" },
		"page host":        func(g *provider.Gif) { g.PageURL = "https://evil.test/p" },
		"page on finer":    func(g *provider.Gif) { g.Provider = provider.FinerGifs },
		"source http":      func(g *provider.Gif) { g.SourceURL = "http://example.org/x" },
		"too many tags":    func(g *provider.Gif) { g.Tags = make([]string, 51) },
		"long tag":         func(g *provider.Gif) { g.Tags = []string{strings.Repeat("t", 101)} },
	}
	for name, mut := range bad {
		g := ok
		mut(&g)
		if err := Validate(g, allow); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}

func TestTouchInvalidNoWrite(t *testing.T) {
	s := newStore(t)
	g := gifN(provider.GIPHY, "a")
	g.DownloadURL = "https://evil.test/a.gif"
	if err := s.Touch(g); !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "recents.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("recents.json written for an invalid gif (%v)", err)
	}
}
