package platform

import (
	"sync"
	"unsafe"
)

// Fake records calls in order; tests read Calls after the code under test returns.
type Fake struct {
	mu       sync.Mutex
	Calls    []string
	CopyErr  error
	WorkArea Rect
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
