package services

import (
	"context"

	"gifsmith/internal/library"
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
