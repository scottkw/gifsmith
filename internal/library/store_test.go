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

func touchAll(t *testing.T, s *Store, idList ...string) {
	t.Helper()
	for _, id := range idList {
		if err := s.Touch(gifN(provider.GIPHY, id)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSetFavorite(t *testing.T) {
	s := newStore(t)
	for _, id := range []string{"a", "b", "a"} { // the repeat is a no-op
		if err := s.SetFavorite(gifN(provider.GIPHY, id), true); err != nil {
			t.Fatal(err)
		}
	}
	if got := ids(s.Snapshot().Favorites); got != "b,a" {
		t.Fatalf("favorites = %s, want b,a (newest first, no duplicate)", got)
	}
	if fi, err := os.Stat(filepath.Join(s.Dir, "favorites.json")); err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("favorites.json: %v %v, want mode 0600", err, fi)
	}
	if err := s.SetFavorite(gifN(provider.GIPHY, "b"), false); err != nil {
		t.Fatal(err)
	}
	if got := ids(s.Snapshot().Favorites); got != "a" {
		t.Errorf("favorites = %s, want a", got)
	}
	if err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := ids(s.Snapshot().Favorites); got != "b,a" {
		t.Errorf("after Undo favorites = %s, want b,a", got)
	}
	if err := s.SetFavorite(gifN(provider.GIPHY, "zzz"), false); err != nil {
		t.Errorf("unfavorite of a missing item: %v, want nil", err)
	}
}

func TestSetFavoriteCap(t *testing.T) {
	s := newStore(t)
	f := listFile{Version: 1}
	for i := 0; i < FavoritesMax; i++ {
		f.Items = append(f.Items, Entry{Gif: gifN(provider.GIPHY, fmt.Sprint(i)), At: 1})
	}
	b, _ := json.Marshal(f)
	if err := os.WriteFile(filepath.Join(s.Dir, "favorites.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.SetFavorite(gifN(provider.GIPHY, "new"), true); err != nil {
		t.Fatal(err)
	}
	fav := s.Snapshot().Favorites
	if len(fav) != FavoritesMax || fav[0].Gif.ID != "new" {
		t.Errorf("%d favorites starting %s, want %d starting new", len(fav), fav[0].Gif.ID, FavoritesMax)
	}
}

func TestSetFavoriteInvalid(t *testing.T) {
	s := newStore(t)
	g := gifN(provider.GIPHY, "a")
	g.PreviewURL = "https://evil.test/p.gif"
	if err := s.SetFavorite(g, true); !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestUndoRestoresPosition(t *testing.T) {
	s := newStore(t)
	touchAll(t, s, "e", "d", "c", "b", "a") // a,b,c,d,e
	before := s.Snapshot().Recents
	if err := s.RemoveRecent(provider.GIPHY, "c"); err != nil {
		t.Fatal(err)
	}
	if got := ids(s.Snapshot().Recents); got != "a,b,d,e" {
		t.Fatalf("after remove = %s", got)
	}
	if err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	after := s.Snapshot().Recents
	if ids(after) != "a,b,c,d,e" || after[2].At != before[2].At {
		t.Fatalf("after Undo = %s (At %d), want a,b,c,d,e with At %d", ids(after), after[2].At, before[2].At)
	}
	if err := s.Undo(); err != nil || ids(s.Snapshot().Recents) != "a,b,c,d,e" {
		t.Errorf("second Undo changed the list: %v %s", err, ids(s.Snapshot().Recents))
	}
	// Clear Recents undoes as a whole, in order.
	if err := s.ClearRecents(); err != nil || len(s.Snapshot().Recents) != 0 {
		t.Fatalf("ClearRecents: %v, %d left", err, len(s.Snapshot().Recents))
	}
	if err := s.Undo(); err != nil || ids(s.Snapshot().Recents) != "a,b,c,d,e" {
		t.Fatalf("Undo of Clear: %v %s", err, ids(s.Snapshot().Recents))
	}
	// Only the latest removal is undoable.
	_ = s.RemoveRecent(provider.GIPHY, "a")
	_ = s.RemoveRecent(provider.GIPHY, "e")
	if err := s.Undo(); err != nil || ids(s.Snapshot().Recents) != "b,c,d,e" {
		t.Errorf("Undo after two removals: %v %s, want b,c,d,e", err, ids(s.Snapshot().Recents))
	}
}

func TestUndoSkipsKeysPresentAgain(t *testing.T) {
	s := newStore(t)
	touchAll(t, s, "c", "b", "a")
	if err := s.ClearRecents(); err != nil {
		t.Fatal(err)
	}
	touchAll(t, s, "b") // re-copied meanwhile
	if err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := ids(s.Snapshot().Recents); got != "a,b,c" {
		t.Errorf("recents = %s, want a,b,c (b once, at its restored index)", got)
	}
	// Caps re-apply: a full list plus an undone item stays at RecentsMax.
	s = newStore(t)
	for i := 0; i < RecentsMax; i++ {
		touchAll(t, s, fmt.Sprint(i))
	}
	_ = s.RemoveRecent(provider.GIPHY, "0") // the oldest
	touchAll(t, s, "new")
	if err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if n := len(s.Snapshot().Recents); n != RecentsMax {
		t.Errorf("%d recents after Undo, want %d", n, RecentsMax)
	}
}

func TestUnreadableFileNeverOverwritten(t *testing.T) {
	s := newStore(t)
	path := filepath.Join(s.Dir, "recents.json")
	bad := []byte("{not json")
	if err := os.WriteFile(path, bad, 0o600); err != nil {
		t.Fatal(err)
	}
	snap := s.Snapshot()
	if fmt.Sprint(snap.Unreadable) != "[recents]" || len(snap.Recents) != 0 {
		t.Fatalf("snapshot unreadable %v, %d recents; want [recents], 0", snap.Unreadable, len(snap.Recents))
	}
	for name, err := range map[string]error{
		"Touch":        s.Touch(gifN(provider.GIPHY, "a")),
		"RemoveRecent": s.RemoveRecent(provider.GIPHY, "a"),
		"ClearRecents": s.ClearRecents(),
	} {
		if !errors.Is(err, ErrUnreadable) {
			t.Errorf("%s: err = %v, want ErrUnreadable", name, err)
		}
	}
	if b, _ := os.ReadFile(path); string(b) != string(bad) {
		t.Fatalf("recents.json = %q, want it untouched", b)
	}
	// Favorites still work while Recents is unreadable.
	if err := s.SetFavorite(gifN(provider.GIPHY, "f"), true); err != nil {
		t.Errorf("favorite while recents unreadable: %v", err)
	}
	// Fixed on disk: the next Snapshot reads it and writes work again.
	if err := os.WriteFile(path, []byte(`{"version":1,"items":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if snap := s.Snapshot(); len(snap.Unreadable) != 0 {
		t.Fatalf("still unreadable after the fix: %v", snap.Unreadable)
	}
	if err := s.Touch(gifN(provider.GIPHY, "a")); err != nil {
		t.Errorf("Touch after the fix: %v", err)
	}
}

func TestUnknownVersionNeverOverwritten(t *testing.T) {
	s := newStore(t)
	path := filepath.Join(s.Dir, "favorites.json")
	v2 := []byte(`{"version":2,"items":[]}`)
	if err := os.WriteFile(path, v2, 0o600); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(s.Snapshot().Unreadable) != "[favorites]" {
		t.Errorf("unreadable = %v, want [favorites]", s.Snapshot().Unreadable)
	}
	if err := s.SetFavorite(gifN(provider.GIPHY, "a"), true); !errors.Is(err, ErrUnreadable) {
		t.Errorf("SetFavorite: %v, want ErrUnreadable", err)
	}
	if b, _ := os.ReadFile(path); string(b) != string(v2) {
		t.Errorf("favorites.json = %q, want it untouched", b)
	}
}

// An entry that fails validation makes the whole file unreadable; it is never dropped silently.
func TestInvalidEntryOnDiskIsUnreadable(t *testing.T) {
	s := newStore(t)
	g := gifN(provider.GIPHY, "a")
	g.DownloadURL = "https://evil.test/a.gif"
	b, _ := json.Marshal(listFile{Version: 1, Items: []Entry{{Gif: gifN(provider.GIPHY, "ok"), At: 1}, {Gif: g, At: 2}}})
	if err := os.WriteFile(filepath.Join(s.Dir, "recents.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if snap := s.Snapshot(); fmt.Sprint(snap.Unreadable) != "[recents]" || len(snap.Recents) != 0 {
		t.Errorf("snapshot %v / %d recents, want [recents] / 0", snap.Unreadable, len(snap.Recents))
	}
}

func searches(s *Store, p provider.ID) string {
	return strings.Join(s.Snapshot().Searches[string(p)], "|")
}

func TestSaveSearchTrim(t *testing.T) {
	s := newStore(t)
	if err := s.SaveSearch(provider.GIPHY, "  Cats "); err != nil {
		t.Fatal(err)
	}
	if got := searches(s, provider.GIPHY); got != "Cats" {
		t.Errorf("searches = %q, want Cats", got)
	}
	if fi, err := os.Stat(filepath.Join(s.Dir, "searches.json")); err != nil || fi.Mode().Perm() != 0o600 {
		t.Errorf("searches.json: %v %v", err, fi)
	}
}

func TestSaveSearchDedupeLatestCasing(t *testing.T) {
	s := newStore(t)
	for _, q := range []string{"Cats", "dogs", "cats"} {
		if err := s.SaveSearch(provider.GIPHY, q); err != nil {
			t.Fatal(err)
		}
	}
	if got := searches(s, provider.GIPHY); got != "cats|dogs" {
		t.Errorf("searches = %q, want cats|dogs", got)
	}
}

func TestSaveSearchCap20(t *testing.T) {
	s := newStore(t)
	for i := 0; i < 21; i++ {
		if err := s.SaveSearch(provider.Klipy, fmt.Sprintf("q%02d", i)); err != nil {
			t.Fatal(err)
		}
	}
	got := s.Snapshot().Searches["klipy"]
	if len(got) != SearchesMax || got[0] != "q20" || got[19] != "q01" {
		t.Errorf("%d searches %v, want 20 from q20 to q01", len(got), got)
	}
}

func TestSaveSearchMinMaxLength(t *testing.T) {
	s := newStore(t)
	for _, q := range []string{"", " a ", "é", strings.Repeat("x", 101)} {
		if err := s.SaveSearch(provider.GIPHY, q); err != nil {
			t.Errorf("%q: %v, want nil (ignored)", q, err)
		}
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "searches.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("searches.json written for ignored queries (%v)", err)
	}
	for _, q := range []string{"ab", "日本", strings.Repeat("x", 100)} {
		if err := s.SaveSearch(provider.GIPHY, q); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(s.Snapshot().Searches["giphy"]); n != 3 {
		t.Errorf("%d searches, want 3", n)
	}
}

func TestSearchPerProvider(t *testing.T) {
	s := newStore(t)
	_ = s.SaveSearch(provider.GIPHY, "cats")
	_ = s.SaveSearch(provider.FinerGifs, "dwight")
	if searches(s, provider.GIPHY) != "cats" || searches(s, provider.FinerGifs) != "dwight" || searches(s, provider.Klipy) != "" {
		t.Errorf("searches = %v", s.Snapshot().Searches)
	}
	if err := s.SaveSearch("tenor", "cats"); !errors.Is(err, ErrInvalid) {
		t.Errorf("unknown provider: %v, want ErrInvalid", err)
	}
	if err := s.RemoveSearch(provider.GIPHY, "CATS"); err != nil {
		t.Fatal(err)
	}
	if searches(s, provider.GIPHY) != "" || searches(s, provider.FinerGifs) != "dwight" {
		t.Errorf("after RemoveSearch: %v", s.Snapshot().Searches)
	}
	// Persisted across stores.
	if got := (&Store{Dir: s.Dir, Allow: allow}).Snapshot().Searches["finergifs"]; len(got) != 1 || got[0] != "dwight" {
		t.Errorf("new store reads %v", got)
	}
}

func TestSearchesFileUnreadable(t *testing.T) {
	s := newStore(t)
	path := filepath.Join(s.Dir, "searches.json")
	bad := []byte(`{"version":1,"searches":{"tenor":["x"]}}`)
	if err := os.WriteFile(path, bad, 0o600); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(s.Snapshot().Unreadable) != "[searches]" {
		t.Errorf("unreadable = %v, want [searches]", s.Snapshot().Unreadable)
	}
	if err := s.SaveSearch(provider.GIPHY, "cats"); !errors.Is(err, ErrUnreadable) {
		t.Errorf("SaveSearch: %v, want ErrUnreadable", err)
	}
	if b, _ := os.ReadFile(path); string(b) != string(bad) {
		t.Errorf("searches.json changed")
	}
}
