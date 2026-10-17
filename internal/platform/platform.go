// Package platform holds the only cgo in the app (darwin && !server). The stub
// keeps -tags server builds CGO-free; Fake records calls for unit tests.
package platform

import (
	"errors"
	"unsafe"
)

var ErrUnsupported = errors.New("not supported in this build")

// Rect is in Wails window coordinates: points, Y down from the top of the primary screen.
type Rect struct{ X, Y, Width, Height int }

type Platform interface {
	// CopyFile puts a file URL for path on the general pasteboard.
	CopyFile(path string) error
	// FocusWebView makes the window's WKWebView first responder so typing reaches the page.
	FocusWebView(nsWindow unsafe.Pointer)
}
