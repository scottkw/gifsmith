package services

import (
	"context"
	"errors"
	"log"

	"gifsmith/internal/library"
	"gifsmith/internal/provider"
)

// LibraryService is Favorites, Recents and recent searches for the webview. It reads and writes
// local files only and never makes a provider request, so it works offline and while a provider
// is rate-limited (D-09). Every call returns the whole fresh Snapshot.
type LibraryService struct {
	Store *library.Store
}

// Load returns the saved lists; a file that failed to load is named in Snapshot.Unreadable.
func (l *LibraryService) Load(ctx context.Context) (library.Snapshot, error) {
	return l.Store.Snapshot(), nil
}

// SetFavorite adds g to Favorites (on) or removes it (undoable).
func (l *LibraryService) SetFavorite(ctx context.Context, g provider.Gif, on bool) (library.Snapshot, error) {
	return l.after(l.Store.SetFavorite(g, on))
}

func (l *LibraryService) RemoveRecent(ctx context.Context, p, id string) (library.Snapshot, error) {
	return l.after(l.Store.RemoveRecent(provider.ID(p), id))
}

func (l *LibraryService) ClearRecents(ctx context.Context) (library.Snapshot, error) {
	return l.after(l.Store.ClearRecents())
}

// Undo restores the latest removal (Remove from Favorites/Recents, Clear Recents).
func (l *LibraryService) Undo(ctx context.Context) (library.Snapshot, error) {
	return l.after(l.Store.Undo())
}

// SaveSearch records a search that returned results; too short or long queries are ignored.
func (l *LibraryService) SaveSearch(ctx context.Context, p, q string) error {
	return libraryErr(l.Store.SaveSearch(provider.ID(p), q))
}

func (l *LibraryService) RemoveSearch(ctx context.Context, p, q string) (library.Snapshot, error) {
	return l.after(l.Store.RemoveSearch(provider.ID(p), q))
}

// after returns the fresh snapshot of a mutation, or its error as an ActionError kind.
func (l *LibraryService) after(err error) (library.Snapshot, error) {
	if err != nil {
		return library.Snapshot{}, libraryErr(err)
	}
	return l.Store.Snapshot(), nil
}

// libraryErr maps store errors to fixed kinds; an I/O error's text (a path) stays in the log.
func libraryErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, library.ErrUnreadable):
		return actionErr("library_read")
	case errors.Is(err, library.ErrInvalid):
		log.Printf("library: %v", err)
		return actionErr("invalid_item")
	}
	log.Printf("library write: %v", err)
	return actionErr("library_write")
}
