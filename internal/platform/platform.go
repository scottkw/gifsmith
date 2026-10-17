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

// App identifies a running application.
type App struct {
	BundleID, Name string
	PID            int
}

type Platform interface {
	// CopyFile puts a file URL for path on the general pasteboard.
	CopyFile(path string) error
	// FocusWebView makes the window's WKWebView first responder so typing reaches the page.
	FocusWebView(nsWindow unsafe.Pointer)
	// CursorScreenWorkArea is the visible frame of the screen under the mouse cursor.
	CursorScreenWorkArea() Rect
	// FrontmostApp is the app that currently owns the menu bar.
	FrontmostApp() (App, error)
	// ActivateApp asks the app with pid to come forward (best effort on macOS 14+).
	ActivateApp(pid int) error
	// AccessibilityTrusted reports the AX grant; prompt=true raises the system prompt.
	AccessibilityTrusted(prompt bool) bool
	// OpenAccessibilitySettings opens the Accessibility app-permission list
	// (x-apple.systempreferences:…?Privacy_Accessibility; macOS 27 shows no such Settings path).
	OpenAccessibilitySettings() error
	// PasteShortcut posts one ⌘V (key down + key up) to the frontmost app.
	PasteShortcut() error
}
