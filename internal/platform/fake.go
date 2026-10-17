package platform

import (
	"strconv"
	"sync"
	"unsafe"
)

// Fake records calls in order; tests read Calls after the code under test returns.
type Fake struct {
	mu        sync.Mutex
	Calls     []string
	CopyErr   error
	WorkArea  Rect
	Trusted   bool
	Frontmost App
	PasteErr  error
	// FrontmostErr makes FrontmostApp fail (nil frontmostApplication during app switches or at the lock screen).
	FrontmostErr error
	// ActivateSetsFrontmost makes ActivateApp succeed; otherwise it is declined (Frontmost unchanged).
	ActivateSetsFrontmost bool
}

func (f *Fake) Note(s string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, s)
}

func (f *Fake) CopyFile(path string) error {
	f.Note("CopyFile:" + path)
	return f.CopyErr
}

func (f *Fake) FocusWebView(unsafe.Pointer) { f.Note("FocusWebView") }

func (f *Fake) CursorScreenWorkArea() Rect {
	f.Note("CursorScreenWorkArea")
	return f.WorkArea
}

func (f *Fake) FrontmostApp() (App, error) {
	f.Note("FrontmostApp")
	if f.FrontmostErr != nil {
		return App{}, f.FrontmostErr
	}
	return f.Frontmost, nil
}

func (f *Fake) ActivateApp(pid int) error {
	f.Note("ActivateApp:" + strconv.Itoa(pid))
	if f.ActivateSetsFrontmost {
		f.Frontmost = App{PID: pid}
	}
	return nil
}

func (f *Fake) AccessibilityTrusted(prompt bool) bool {
	f.Note("AccessibilityTrusted:" + strconv.FormatBool(prompt))
	return f.Trusted
}

func (f *Fake) OpenAccessibilitySettings() error {
	f.Note("OpenAccessibilitySettings")
	return nil
}

func (f *Fake) PasteShortcut() error {
	f.Note("PasteShortcut")
	return f.PasteErr
}
