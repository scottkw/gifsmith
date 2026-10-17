// Package services holds the methods bound to the webview. They take no paths or
// URLs from JS: the only file they can touch is SamplePath, fixed in Go.
package services

import (
	"context"
	"log"
	"time"

	"gifsmith/internal/platform"
)

// PasteDelay is the wait between hiding the panel (or reactivating the previous app) and ⌘V.
const PasteDelay = 80 * time.Millisecond

type Result struct {
	Copied             bool   `json:"copied"`
	Pasted             bool   `json:"pasted"`
	NeedsAccessibility bool   `json:"needsAccessibility"`
	Message            string `json:"message"`
}

type ActionService struct {
	Platform   platform.Platform
	SamplePath string
	Hide       func()
	Previous   func() platform.App
	Sleep      func(time.Duration) // nil means time.Sleep
}

// Copy puts the staged sample GIF on the clipboard as a file and hides the panel.
// On failure the panel stays open so the frontend can show the error.
func (a *ActionService) Copy(ctx context.Context) (Result, error) {
	if err := a.Platform.CopyFile(a.SamplePath); err != nil {
		return Result{}, err
	}
	log.Printf("copy: copied, hiding panel")
	a.Hide()
	return Result{Copied: true, Message: "Copied GIF — paste with ⌘V"}, nil
}

// Paste copies the GIF, hides the panel and posts one ⌘V to the previous app.
// Without Accessibility it only copies and keeps the panel open with instructions.
func (a *ActionService) Paste(ctx context.Context) (Result, error) {
	if err := a.Platform.CopyFile(a.SamplePath); err != nil {
		return Result{}, err
	}
	if !a.Platform.AccessibilityTrusted(false) {
		log.Printf("paste: accessibility not trusted, copied only (UI shows Enable automatic paste…)")
		return Result{Copied: true, NeedsAccessibility: true, Message: "Copied — press ⌘V to paste. To paste automatically, turn on Gifsmith in System Settings › Privacy & Security › Accessibility."}, nil
	}
	sleep := a.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	a.Hide()
	sleep(PasteDelay)
	if prev := a.Previous(); prev.PID != 0 {
		if front, err := a.Platform.FrontmostApp(); err == nil && front.PID != prev.PID {
			if err := a.Platform.ActivateApp(prev.PID); err != nil {
				log.Printf("reactivate %s: %v", prev.Name, err) // best effort; post anyway
			}
			sleep(PasteDelay)
		}
	}
	if err := a.Platform.PasteShortcut(); err != nil {
		return Result{Copied: true}, err
	}
	log.Printf("paste: posted ⌘V (previous app %q)", a.Previous().Name)
	return Result{Copied: true, Pasted: true, Message: "Pasted GIF"}, nil
}

// RequestAccessibility raises the macOS prompt and, if still untrusted, opens the settings pane.
// Only an explicit button click calls this.
func (a *ActionService) RequestAccessibility(ctx context.Context) (bool, error) {
	if a.Platform.AccessibilityTrusted(true) {
		log.Printf("request accessibility: already trusted")
		return true, nil
	}
	log.Printf("request accessibility: prompted, opening the Accessibility pane")
	return false, a.Platform.OpenAccessibilitySettings()
}
