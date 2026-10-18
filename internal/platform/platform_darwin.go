//go:build darwin && !server

package platform

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=13.0
#cgo LDFLAGS: -framework Cocoa -framework Carbon -framework ApplicationServices
#include <stdlib.h>
#include "platform_darwin.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

type darwin struct{}

func New() Platform { return darwin{} }

func (darwin) CopyFile(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}
	return copyFileToPasteboard("", path)
}

func (darwin) CopyText(text string) error { return copyTextToPasteboard("", text) }

func (darwin) OpenURL(u string) error { return errors.New("not implemented") }

// FocusWebView takes the pointer from WebviewWindow.NativeWindow() (nil before the app runs).
func (darwin) FocusWebView(nsWindow unsafe.Pointer) {
	if nsWindow != nil {
		C.gsFocusWebView(nsWindow)
	}
}

func (darwin) CursorScreenWorkArea() Rect {
	var x, y, w, h C.int
	C.gsCursorScreenVisibleFrame(&x, &y, &w, &h)
	return Rect{X: int(x), Y: int(y), Width: int(w), Height: int(h)}
}

func (darwin) FrontmostApp() (App, error) {
	var pid C.int
	var bid, name *C.char
	if C.gsFrontmostApp(&pid, &bid, &name) == 0 {
		return App{}, errors.New("no frontmost app")
	}
	defer C.free(unsafe.Pointer(bid))
	defer C.free(unsafe.Pointer(name))
	return App{BundleID: C.GoString(bid), Name: C.GoString(name), PID: int(pid)}, nil
}

func (darwin) ActivateApp(pid int) error {
	if C.gsActivate(C.int(pid)) == 0 {
		return fmt.Errorf("could not activate pid %d", pid)
	}
	return nil
}

func (darwin) AccessibilityTrusted(prompt bool) bool {
	p := C.int(0)
	if prompt {
		p = 1
	}
	return C.gsAXTrusted(p) != 0
}

func (darwin) OpenAccessibilitySettings() error {
	if C.gsOpenAXSettings() == 0 {
		return errors.New("could not open Accessibility settings")
	}
	return nil
}

func (darwin) PasteShortcut() error {
	if C.gsPostCmdV() == 0 {
		return errors.New("could not post ⌘V")
	}
	return nil
}

// The helpers below take a pasteboard name ("" = general) so tests can use a private one;
// they live here because _test.go files cannot use cgo.

func copyFileToPasteboard(name, path string) error {
	cn, cp := C.CString(name), C.CString(path)
	defer C.free(unsafe.Pointer(cn))
	defer C.free(unsafe.Pointer(cp))
	if C.gsCopyFile(cn, cp) == 0 {
		return errors.New("pasteboard write failed")
	}
	return nil
}

// pasteboardFileURLs returns the file paths on the pasteboard and whether public.file-url is among its types.
func pasteboardFileURLs(name string) ([]string, bool, error) {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	out := C.gsPasteboardFileURLs(cn)
	if out == nil {
		return nil, false, errors.New("pasteboard read failed")
	}
	defer C.free(unsafe.Pointer(out))
	lines := strings.Split(C.GoString(out), "\n")
	types := strings.Fields(strings.TrimPrefix(lines[0], "types:"))
	hasFileURL := false
	for _, t := range types {
		hasFileURL = hasFileURL || t == "public.file-url"
	}
	return lines[1:], hasFileURL, nil
}

func copyTextToPasteboard(name, text string) error { return errors.New("not implemented") }

func pasteboardString(name string) (string, error) { return "", errors.New("not implemented") }

func releasePasteboard(name string) {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	C.gsReleasePasteboard(cn)
}
