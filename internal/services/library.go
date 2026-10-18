package services

import (
	"context"

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

// Skeletons (RED).
func (l *LibraryService) SetFavorite(ctx context.Context, g provider.Gif, on bool) (library.Snapshot, error) {
	return library.Snapshot{}, nil
}
func (l *LibraryService) RemoveRecent(ctx context.Context, p, id string) (library.Snapshot, error) {
	return library.Snapshot{}, nil
}
func (l *LibraryService) ClearRecents(ctx context.Context) (library.Snapshot, error) {
	return library.Snapshot{}, nil
}
func (l *LibraryService) Undo(ctx context.Context) (library.Snapshot, error) {
	return library.Snapshot{}, nil
}
func (l *LibraryService) SaveSearch(ctx context.Context, p, q string) error { return nil }
func (l *LibraryService) RemoveSearch(ctx context.Context, p, q string) (library.Snapshot, error) {
	return library.Snapshot{}, nil
}
