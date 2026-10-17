// Package services holds the methods bound to the webview. They take no paths or
// URLs from JS: the only file they can touch is SamplePath, fixed in Go.
package services

import (
	"context"

	"gifsmith/internal/platform"
)

type Result struct {
	Copied  bool   `json:"copied"`
	Message string `json:"message"`
}

type ActionService struct {
	Platform   platform.Platform
	SamplePath string
	Hide       func()
}

// Copy puts the staged sample GIF on the clipboard as a file and hides the panel.
// On failure the panel stays open so the frontend can show the error.
func (a *ActionService) Copy(ctx context.Context) (Result, error) {
	if err := a.Platform.CopyFile(a.SamplePath); err != nil {
		return Result{}, err
	}
	a.Hide()
	return Result{Copied: true, Message: "Copied GIF — paste with ⌘V"}, nil
}
