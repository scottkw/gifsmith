//go:build !darwin || server

package platform

import "unsafe"

type stub struct{}

func New() Platform { return stub{} }

func (stub) CopyFile(string) error            { return ErrUnsupported }
func (stub) CopyText(string) error            { return ErrUnsupported }
func (stub) OpenURL(string) error             { return ErrUnsupported }
func (stub) FocusWebView(unsafe.Pointer)      {}
func (stub) CursorScreenWorkArea() Rect       { return Rect{} }
func (stub) FrontmostApp() (App, error)       { return App{}, ErrUnsupported }
func (stub) ActivateApp(int) error            { return ErrUnsupported }
func (stub) AccessibilityTrusted(bool) bool   { return false }
func (stub) OpenAccessibilitySettings() error { return ErrUnsupported }
func (stub) PasteShortcut() error             { return ErrUnsupported }
func (stub) RevealFile(string) error          { return ErrUnsupported }
