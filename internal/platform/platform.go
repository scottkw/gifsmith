// Package platform holds the only cgo in the app (darwin && !server). The stub
// keeps -tags server builds CGO-free; Fake records calls for unit tests.
package platform

import (
	"errors"
	"unsafe"
)

var ErrUnsupported = errors.New("not supported in this build")

type Platform interface {
	// CopyFile puts a file URL for path on the general pasteboard.
	CopyFile(path string) error
	// FocusWebView makes the window's WKWebView first responder so typing reaches the page.
	FocusWebView(nsWindow unsafe.Pointer)
}
