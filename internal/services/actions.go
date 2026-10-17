// Package services holds the methods bound to the webview. They take no paths or
// URLs from JS: the only file they can touch is SamplePath, fixed in Go.
package services

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"gifsmith/internal/platform"
)

// ErrBusy is returned when another action is still running (each bound call has its own goroutine).
var ErrBusy = errors.New("another action is still running")

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
	busy       sync.Mutex          // one action at a time: no racing pasteboard writes, no double ⌘V
}

// Copy puts the staged sample GIF on the clipboard as a file and hides the panel.
// On failure the panel stays open so the frontend can show the error.
func (a *ActionService) Copy(ctx context.Context) (Result, error) {
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
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
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
	if err := a.Platform.CopyFile(a.SamplePath); err != nil {
		return Result{}, err
	}
	if !a.Platform.AccessibilityTrusted(false) {
		log.Printf("paste: accessibility not trusted, copied only (UI shows Enable automatic paste…)")
		return Result{Copied: true, NeedsAccessibility: true, Message: "Copied — press ⌘V to paste. To paste automatically, click Enable automatic paste… and turn on Gifsmith."}, nil
	}
	sleep := a.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	a.Hide()
	sleep(PasteDelay)
	// Never post ⌘V into an app the user did not pick: an unknown target or a failed lookup copies only.
	prev := a.Previous()
	if prev.PID == 0 {
		log.Printf("paste: no remembered target, copied only")
		return Result{Copied: true, Message: "Copied — press ⌘V to paste"}, nil
	}
	front, err := a.Platform.FrontmostApp()
	if err != nil {
		log.Printf("paste: frontmost app: %v, copied only", err)
		return Result{Copied: true, Message: "Copied — press ⌘V to paste"}, nil
	}
	if front.PID != prev.PID {
		if err := a.Platform.ActivateApp(prev.PID); err != nil {
			log.Printf("reactivate %s: %v", prev.Name, err)
		}
		sleep(PasteDelay)
		// macOS 14+ may decline the activation.
		if now, err := a.Platform.FrontmostApp(); err != nil || now.PID != prev.PID {
			log.Printf("paste: could not return to %q, copied only", prev.Name)
			return Result{Copied: true, Message: "Copied — press ⌘V to paste (could not return to " + prev.Name + ")"}, nil
		}
	}
	if err := a.Platform.PasteShortcut(); err != nil {
		// The panel is already hidden, so an error here would never be seen; report it as a result.
		log.Printf("paste: post ⌘V: %v", err)
		return Result{Copied: true, Message: "Copied — automatic paste failed, press ⌘V"}, nil
	}
	log.Printf("paste: posted ⌘V (previous app %q)", prev.Name)
	return Result{Copied: true, Pasted: true, Message: "Pasted GIF"}, nil
}

// RequestAccessibility raises the macOS prompt and, if still untrusted, opens the settings pane.
// Only an explicit button click calls this.
func (a *ActionService) RequestAccessibility(ctx context.Context) (bool, error) {
	if !a.busy.TryLock() {
		return false, ErrBusy
	}
	defer a.busy.Unlock()
	if a.Platform.AccessibilityTrusted(true) {
		log.Printf("request accessibility: already trusted")
		return true, nil
	}
	log.Printf("request accessibility: prompted, opening the Accessibility pane")
	return false, a.Platform.OpenAccessibilitySettings()
}
