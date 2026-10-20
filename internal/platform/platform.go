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
	// CopyText puts a plain string on the general pasteboard (public.utf8-plain-text).
	CopyText(text string) error
	// OpenURL opens a URL with its default handler: https links in the browser, the verified
	// update DMG's file URL in DiskImageMounter (callers validate).
	OpenURL(u string) error
	// FocusWebView makes the window's WKWebView first responder so typing reaches the page.
	FocusWebView(nsWindow unsafe.Pointer)
	// CursorScreenWorkArea is the visible frame of the screen under the mouse cursor (zero Rect if there is no screen).
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
	// RevealFile selects path in a Finder window (callers pass only files they saved).
	RevealFile(path string) error
	// SetAppearance sets the app-wide appearance: "" follows the system, "light" Aqua, "dark" DarkAqua.
	SetAppearance(mode string) error
	// SystemLocale is the first preferred language (BCP 47, e.g. "pt-BR") and the region code
	// (e.g. "DE"); either may be "" when macOS reports none.
	SystemLocale() (lang, region string)
}
