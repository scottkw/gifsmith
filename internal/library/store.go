// Package library keeps Favorites, Recents and recent searches as JSON files in Dir. It stores
// metadata only (URLs, title, dimensions, tags), never media bytes (D-09), and never makes a
// provider request: no Wails import, no provider HTTP client.
package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gifsmith/internal/provider"
	"gifsmith/internal/settings"
)

const (
	RecentsMax   = 50
	FavoritesMax = 1000 // sanity cap
	SearchesMax  = 20   // per provider
)

var (
	// ErrUnreadable: the file exists but could not be read, parsed or validated, or has an
	// unknown version. Writes to it are refused until it reads cleanly again (Pitfall 11).
	ErrUnreadable = errors.New("library file unreadable")
	ErrInvalid    = errors.New("invalid library item")
)

type Entry struct {
	Gif provider.Gif `json:"gif"`
	At  int64        `json:"at"` // unix ms: added (Favorites) or last used (Recents)
}

// Snapshot is everything the library tabs render. Lists are never nil; Unreadable names the
// files ("favorites" | "recents" | "searches") that failed to load.
type Snapshot struct {
	Favorites  []Entry             `json:"favorites"`
	Recents    []Entry             `json:"recents"`
	Searches   map[string][]string `json:"searches"`
	Unreadable []string            `json:"unreadable"`
}

// listFile is favorites.json and recents.json, newest first.
type listFile struct {
	Version int     `json:"version"`
	Items   []Entry `json:"items"`
}

// list is one list file in memory. ok is false until the file read cleanly; until then every
// access re-reads it and every write is refused.
type list struct {
	name  string
	max   int
	items []Entry
	ok    bool
}

// Store is safe for concurrent use. Every mutation writes the new list first and swaps it into
// memory only after the write succeeded.
type Store struct {
	Dir   string
	Allow func(provider.ID, *url.URL) bool // the staging media allowlist
	Now   func() time.Time                 // nil means time.Now

	mu       sync.Mutex
	fav, rec list
	srch     map[string][]string // nil until searches.json read cleanly
	undo     *undo
}

func key(g provider.Gif) string { return string(g.Provider) + ":" + g.ID }

// Validate checks a Gif before it is persisted (it may come from JS or a tampered file), so a
// saved item can never later feed a disallowed URL to copy or open (T-03-21).
func Validate(g provider.Gif, allow func(provider.ID, *url.URL) bool) error {
	bad := func(what string) error { return fmt.Errorf("%w: %s", ErrInvalid, what) }
	if _, ok := provider.ParseID(string(g.Provider)); !ok {
		return bad("provider")
	}
	if g.ID == "" || len(g.ID) > 200 {
		return bad("id")
	}
	if utf8.RuneCountInString(g.Title) > 500 {
		return bad("title")
	}
	allowed := func(raw string) bool {
		u, err := url.Parse(raw)
		return err == nil && allow != nil && allow(g.Provider, u)
	}
	optional := func(raw string) bool { return raw == "" || allowed(raw) }
	if !allowed(g.PreviewURL) || !allowed(g.DownloadURL) || !optional(g.MediumURL) || !optional(g.GifURL) {
		return bad("media url")
	}
	// staging.PageAllow's rule (that package pulls in provider HTTP): GIPHY and Clips only, on an allowed host.
	if g.PageURL != "" && (g.Provider != provider.GIPHY && g.Provider != provider.Clips || !allowed(g.PageURL)) {
		return bad("page url")
	}
	if g.SourceURL != "" {
		if u, err := url.Parse(g.SourceURL); err != nil || u.Scheme != "https" || u.Host == "" {
			return bad("source url")
		}
	}
	if len(g.Tags) > 50 || slices.ContainsFunc(g.Tags, func(t string) bool { return utf8.RuneCountInString(t) > 100 }) {
		return bad("tags")
	}
	if g.Width < 0 || g.Height < 0 || g.Size < 0 {
		return bad("dimensions")
	}
	return nil
}

func (s *Store) now() int64 {
	if s.Now == nil {
		return time.Now().UnixMilli()
	}
	return s.Now().UnixMilli()
}

// read decodes Dir/name.json into v. A missing file is (false, nil); any other failure is ErrUnreadable.
func (s *Store) read(name string, v any) (bool, error) {
	b, err := os.ReadFile(filepath.Join(s.Dir, name+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err == nil {
		err = json.Unmarshal(b, v)
	}
	if err != nil {
		log.Printf("library: read %s: %v", name, err)
		return true, ErrUnreadable
	}
	return true, nil
}

// load reads l's file unless it already read cleanly. Content that does not validate makes the
// whole file unreadable: entries are never silently dropped.
func (s *Store) load(l *list, name string, max int) (*list, error) {
	if l.ok {
		return l, nil
	}
	l.name, l.max = name, max
	var f listFile
	found, err := s.read(name, &f)
	if err == nil && found {
		if err = s.check(f); err != nil {
			log.Printf("library: %s: %v", name, err)
		}
	}
	if err != nil {
		return l, ErrUnreadable
	}
	l.items, l.ok = f.Items, true
	return l, nil
}

func (s *Store) check(f listFile) error {
	if f.Version != 1 {
		return fmt.Errorf("version %d", f.Version)
	}
	seen := map[string]bool{}
	for _, e := range f.Items {
		if err := Validate(e.Gif, s.Allow); err != nil {
			return err
		}
		if seen[key(e.Gif)] {
			return errors.New("duplicate item")
		}
		seen[key(e.Gif)] = true
	}
	return nil
}

func (s *Store) recents() (*list, error)   { return s.load(&s.rec, "recents", RecentsMax) }
func (s *Store) favorites() (*list, error) { return s.load(&s.fav, "favorites", FavoritesMax) }

// write caps items, writes l's file and only then swaps items into memory.
func (s *Store) write(l *list, items []Entry) error {
	if len(items) > l.max {
		items = items[:l.max]
	}
	if err := settings.WriteJSONAtomic(s.Dir, l.name+".json", listFile{Version: 1, Items: items}); err != nil {
		return err
	}
	l.items = items
	return nil
}

// without returns a copy of items with k removed.
func without(items []Entry, k string) []Entry {
	return slices.DeleteFunc(slices.Clone(items), func(e Entry) bool { return key(e.Gif) == k })
}

// Snapshot loads whatever has not loaded yet (re-reading unreadable files) and returns copies.
func (s *Store) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := Snapshot{Favorites: []Entry{}, Recents: []Entry{}, Searches: map[string][]string{}, Unreadable: []string{}}
	if l, err := s.favorites(); err != nil {
		snap.Unreadable = append(snap.Unreadable, "favorites")
	} else {
		snap.Favorites = append(snap.Favorites, l.items...)
	}
	if l, err := s.recents(); err != nil {
		snap.Unreadable = append(snap.Unreadable, "recents")
	} else {
		snap.Recents = append(snap.Recents, l.items...)
	}
	// Lists are replaced, never changed in place, so a shallow copy is safe to hand out.
	if m, err := s.loadSearches(); err != nil {
		snap.Unreadable = append(snap.Unreadable, "searches")
	} else {
		snap.Searches = maps.Clone(m)
	}
	return snap
}

// Touch records a successful copy or paste: g moves to the top of Recents (no duplicates).
func (s *Store) Touch(g provider.Gif) error {
	if err := Validate(g, s.Allow); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, err := s.recents()
	if err != nil {
		return err
	}
	return s.write(l, append([]Entry{{Gif: g, At: s.now()}}, without(l.items, key(g))...))
}

// SetFavorite adds g at the top of Favorites (a repeat is a no-op) or removes it (undoable).
func (s *Store) SetFavorite(g provider.Gif, on bool) error {
	if on {
		if err := Validate(g, s.Allow); err != nil {
			return err
		}
	} else if _, ok := provider.ParseID(string(g.Provider)); !ok {
		return fmt.Errorf("%w: provider", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, err := s.favorites()
	if err != nil {
		return err
	}
	at := l.index(key(g))
	if !on {
		return s.remove(l, at)
	}
	if at != nil {
		return nil
	}
	return s.write(l, append([]Entry{{Gif: g, At: s.now()}}, l.items...))
}

// RemoveRecent removes one Recents item (undoable); a missing item is a no-op.
func (s *Store) RemoveRecent(p provider.ID, id string) error {
	if _, ok := provider.ParseID(string(p)); !ok {
		return fmt.Errorf("%w: provider", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, err := s.recents()
	if err != nil {
		return err
	}
	return s.remove(l, l.index(string(p)+":"+id))
}

// ClearRecents empties Recents; Undo restores the whole list.
func (s *Store) ClearRecents() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, err := s.recents()
	if err != nil {
		return err
	}
	all := make([]int, len(l.items))
	for i := range all {
		all[i] = i
	}
	return s.remove(l, all)
}

// Undo restores the latest removal at its original positions, skipping items that are back in
// the list meanwhile, then re-applies the cap. With nothing to undo it is a no-op.
func (s *Store) Undo() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.undo
	if u == nil {
		return nil
	}
	items := slices.Clone(u.l.items)
	for i, e := range u.entries {
		if slices.ContainsFunc(items, func(x Entry) bool { return key(x.Gif) == key(e.Gif) }) {
			continue
		}
		items = slices.Insert(items, min(u.indices[i], len(items)), e)
	}
	if err := s.write(u.l, items); err != nil {
		return err
	}
	s.undo = nil
	return nil
}

// undo is the one Undo slot: the latest removal from one list, entries with their indices (ascending).
type undo struct {
	l       *list
	entries []Entry
	indices []int
}

func (l *list) index(k string) []int {
	if i := slices.IndexFunc(l.items, func(e Entry) bool { return key(e.Gif) == k }); i >= 0 {
		return []int{i}
	}
	return nil
}

// remove deletes the entries at idx (ascending) from l and makes them the Undo slot, replacing
// any earlier removal. The slot changes only after the write succeeded.
func (s *Store) remove(l *list, idx []int) error {
	if len(idx) == 0 {
		return nil
	}
	items := slices.Clone(l.items)
	gone := make([]Entry, len(idx))
	for i := len(idx) - 1; i >= 0; i-- {
		gone[i] = items[idx[i]]
		items = slices.Delete(items, idx[i], idx[i]+1)
	}
	if err := s.write(l, items); err != nil {
		return err
	}
	s.undo = &undo{l: l, entries: gone, indices: idx}
	return nil
}

// searchFile is searches.json: per provider id, newest first.
type searchFile struct {
	Version  int                 `json:"version"`
	Searches map[string][]string `json:"searches"`
}

// cleanQuery trims q; ok is false when it is shorter than 2 code points or longer than 100.
func cleanQuery(q string) (string, bool) {
	q = strings.TrimSpace(q)
	n := utf8.RuneCountInString(q)
	return q, n >= 2 && n <= 100
}

func checkSearches(f searchFile) error {
	if f.Version != 1 {
		return fmt.Errorf("version %d", f.Version)
	}
	for p, qs := range f.Searches {
		if _, ok := provider.ParseID(p); !ok || len(qs) > SearchesMax {
			return fmt.Errorf("provider %q", p)
		}
		for i, q := range qs {
			if c, ok := cleanQuery(q); !ok || c != q || slices.ContainsFunc(qs[:i], func(x string) bool { return strings.EqualFold(x, q) }) {
				return errors.New("bad query")
			}
		}
	}
	return nil
}

// loadSearches reads searches.json unless it already read cleanly (same rules as the lists).
func (s *Store) loadSearches() (map[string][]string, error) {
	if s.srch != nil {
		return s.srch, nil
	}
	var f searchFile
	found, err := s.read("searches", &f)
	if err == nil && found {
		if err = checkSearches(f); err != nil {
			log.Printf("library: searches: %v", err)
		}
	}
	if err != nil {
		return nil, ErrUnreadable
	}
	if f.Searches == nil {
		f.Searches = map[string][]string{}
	}
	s.srch = f.Searches
	return s.srch, nil
}

// setSearches writes p's list (empty deletes it), then swaps the new map into memory.
func (s *Store) setSearches(m map[string][]string, p provider.ID, qs []string) error {
	next := maps.Clone(m)
	if len(qs) == 0 {
		delete(next, string(p))
	} else {
		next[string(p)] = qs
	}
	if err := settings.WriteJSONAtomic(s.Dir, "searches.json", searchFile{Version: 1, Searches: next}); err != nil {
		return err
	}
	s.srch = next
	return nil
}

func withoutQuery(qs []string, q string) []string {
	return slices.DeleteFunc(slices.Clone(qs), func(x string) bool { return strings.EqualFold(x, q) })
}

// SaveSearch puts q at the front of p's recent searches (case-insensitive dedupe keeps the latest
// casing, cap SearchesMax). Queries under 2 or over 100 code points are ignored.
func (s *Store) SaveSearch(p provider.ID, q string) error {
	if _, ok := provider.ParseID(string(p)); !ok {
		return fmt.Errorf("%w: provider", ErrInvalid)
	}
	q, ok := cleanQuery(q)
	if !ok {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.loadSearches()
	if err != nil {
		return err
	}
	qs := append([]string{q}, withoutQuery(m[string(p)], q)...)
	return s.setSearches(m, p, qs[:min(len(qs), SearchesMax)])
}

// RemoveSearch drops q (case-insensitive) from p's recent searches; a missing query is a no-op.
func (s *Store) RemoveSearch(p provider.ID, q string) error {
	if _, ok := provider.ParseID(string(p)); !ok {
		return fmt.Errorf("%w: provider", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.loadSearches()
	if err != nil {
		return err
	}
	qs := withoutQuery(m[string(p)], strings.TrimSpace(q))
	if len(qs) == len(m[string(p)]) {
		return nil
	}
	return s.setSearches(m, p, qs)
}
