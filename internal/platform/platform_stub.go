//go:build !darwin || server

package platform

import "unsafe"

type stub struct{}

func New() Platform { return stub{} }

func (stub) CopyFile(string) error       { return ErrUnsupported }
func (stub) FocusWebView(unsafe.Pointer) {}
func (stub) CursorScreenWorkArea() Rect  { return Rect{} }
