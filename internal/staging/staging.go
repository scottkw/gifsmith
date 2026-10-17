// Package staging downloads the selected GIF to the one file the pasteboard points at.
package staging

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"gifsmith/internal/provider"
)

// Fixed, URL-free error text: these reach the webview through Wails.
var (
	ErrNotAllowed = errors.New("this GIF's address is not allowed")
	ErrDownload   = errors.New("the GIF download failed")
	ErrNotGIF     = errors.New("the download is not a GIF")
	ErrTooLarge   = errors.New("the GIF is too large to copy")
)

type Store struct {
	Dir      string
	HTTP     *http.Client
	Allow    func(p provider.ID, u *url.URL) bool // nil = VendorAllow
	maxBytes int64                                // 0 = 50 MB
}

func (s *Store) Fetch(ctx context.Context, g provider.Gif) (string, error) {
	return "", ErrDownload
}

func (s *Store) Sweep(keep string) error { return nil }

func VendorAllow(p provider.ID, u *url.URL) bool { return false }

func slug(title string) string { return "" }
