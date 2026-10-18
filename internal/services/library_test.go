package services

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"gifsmith/internal/library"
	"gifsmith/internal/provider"
)

// libService keeps its files in t.TempDir(), never Application Support.
func libService(t *testing.T) *LibraryService {
	t.Helper()
	allow := func(_ provider.ID, u *url.URL) bool { return u.Scheme == "https" && u.Host == "media.test" }
	return &LibraryService{Store: &library.Store{Dir: t.TempDir(), Allow: allow}}
}

func libGif(id string) provider.Gif {
	return provider.Gif{Provider: provider.GIPHY, ID: id, Title: "Moment " + id,
		PreviewURL: "https://media.test/" + id + "-200w.gif", DownloadURL: "https://media.test/" + id + ".gif"}
}

func TestLibraryServiceReturnsFreshSnapshots(t *testing.T) {
	ctx := context.Background()
	l := libService(t)
	snap, err := l.SetFavorite(ctx, libGif("a"), true)
	if err != nil || len(snap.Favorites) != 1 || snap.Favorites[0].Gif.ID != "a" {
		t.Fatalf("SetFavorite: %+v, %v", snap, err)
	}
	if err := l.Store.Touch(libGif("r")); err != nil {
		t.Fatal(err)
	}
	if snap, err = l.RemoveRecent(ctx, "giphy", "r"); err != nil || len(snap.Recents) != 0 || len(snap.Favorites) != 1 {
		t.Fatalf("RemoveRecent: %+v, %v", snap, err)
	}
	if snap, err = l.Undo(ctx); err != nil || len(snap.Recents) != 1 {
		t.Fatalf("Undo: %+v, %v", snap, err)
	}
	if snap, err = l.ClearRecents(ctx); err != nil || len(snap.Recents) != 0 {
		t.Fatalf("ClearRecents: %+v, %v", snap, err)
	}
	if err := l.SaveSearch(ctx, "giphy", "cats"); err != nil {
		t.Fatal(err)
	}
	if snap, err = l.RemoveSearch(ctx, "giphy", "Cats"); err != nil || len(snap.Searches["giphy"]) != 0 {
		t.Fatalf("RemoveSearch: %+v, %v", snap, err)
	}
	if snap, err = l.Load(ctx); err != nil || len(snap.Favorites) != 1 {
		t.Fatalf("Load: %+v, %v", snap, err)
	}
}

func TestLibraryServiceErrorKinds(t *testing.T) {
	ctx := context.Background()
	l := libService(t)
	bad := libGif("a")
	bad.DownloadURL = "https://evil.test/a.gif"
	if _, err := l.SetFavorite(ctx, bad, true); actionKind(err) != "invalid_item" {
		t.Errorf("invalid gif: %v, want invalid_item", err)
	}
	if _, err := l.RemoveRecent(ctx, "tenor", "a"); actionKind(err) != "invalid_item" {
		t.Errorf("unknown provider: %v, want invalid_item", err)
	}
	if err := l.SaveSearch(ctx, "tenor", "cats"); actionKind(err) != "invalid_item" {
		t.Errorf("SaveSearch unknown provider: %v, want invalid_item", err)
	}
	if err := os.WriteFile(filepath.Join(l.Store.Dir, "favorites.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetFavorite(ctx, libGif("a"), true); actionKind(err) != "library_read" {
		t.Errorf("unreadable favorites: %v, want library_read", err)
	}
	snap, err := l.Load(ctx)
	if err != nil || len(snap.Unreadable) != 1 || snap.Unreadable[0] != "favorites" {
		t.Errorf("Load: %+v, %v; want favorites unreadable, no error", snap, err)
	}
}

// A failed write reports library_write and leaves the in-memory list as it was.
func TestLibraryServiceWriteFailureKeepsList(t *testing.T) {
	ctx := context.Background()
	l := libService(t)
	if _, err := l.SetFavorite(ctx, libGif("a"), true); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(l.Store.Dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(l.Store.Dir, 0o700) })
	if _, err := l.SetFavorite(ctx, libGif("b"), true); actionKind(err) != "library_write" {
		t.Fatalf("read-only dir: %v, want library_write", err)
	}
	if _, err := l.SetFavorite(ctx, libGif("a"), false); actionKind(err) != "library_write" {
		t.Fatalf("read-only dir unfavorite: %v, want library_write", err)
	}
	snap, _ := l.Load(ctx)
	if len(snap.Favorites) != 1 || snap.Favorites[0].Gif.ID != "a" {
		t.Errorf("favorites after failed writes = %+v, want just a", snap.Favorites)
	}
	if _, err := l.Undo(ctx); err != nil {
		t.Errorf("Undo with nothing removed: %v, want nil", err) // the failed removal filled no Undo slot
	}
}
