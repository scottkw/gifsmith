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

func releasePasteboard(name string) {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	C.gsReleasePasteboard(cn)
}
