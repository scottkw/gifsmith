package services

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"gifsmith/internal/platform"
)

const sample = "/tmp/staging/sample.gif"

func newService(f *platform.Fake) *ActionService {
	return &ActionService{Platform: f, SamplePath: sample, Hide: func() { f.Note("hide") }}
}

func TestCopyCopiesThenHides(t *testing.T) {
	f := &platform.Fake{}
	r, err := newService(f).Copy(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Message == "" {
		t.Errorf("result = %+v, want Copied with a message", r)
	}
	if want := []string{"CopyFile:" + sample, "hide"}; !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestCopyErrorKeepsPanelOpen(t *testing.T) {
	f := &platform.Fake{CopyErr: errors.New("pasteboard write failed")}
	r, err := newService(f).Copy(context.Background())
	if !errors.Is(err, f.CopyErr) {
		t.Errorf("err = %v, want %v", err, f.CopyErr)
	}
	if r.Copied {
		t.Error("Copied = true on error")
	}
	if slices.Contains(f.Calls, "hide") {
		t.Errorf("panel hidden on error: calls = %q", f.Calls)
	}
}

func TestCopyTwiceUsesSamePath(t *testing.T) {
	f := &platform.Fake{}
	s := newService(f)
	for range 2 {
		if _, err := s.Copy(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"CopyFile:" + sample, "hide", "CopyFile:" + sample, "hide"}
	if !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

var prevApp = platform.App{BundleID: "com.apple.TextEdit", Name: "TextEdit", PID: 42}

func newPasteService(f *platform.Fake) *ActionService {
	s := newService(f)
	s.Previous = func() platform.App { return prevApp }
	s.Sleep = func(d time.Duration) { f.Note("sleep:" + d.String()) }
	return s
}

func TestPasteWithoutAccessibility(t *testing.T) {
	f := &platform.Fake{}
	r, err := newPasteService(f).Paste(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || !r.NeedsAccessibility || r.Pasted || r.Message == "" {
		t.Errorf("result = %+v, want Copied+NeedsAccessibility with a message", r)
	}
	// Point at the button, not a Settings path (macOS 27 has no Privacy & Security › Accessibility).
	if !strings.Contains(r.Message, "Enable automatic paste…") || strings.Contains(r.Message, "System Settings") {
		t.Errorf("message = %q, want it to name the Enable automatic paste… button and no Settings path", r.Message)
	}
	// App.svelte shows "Enable automatic paste…" off this exact JSON key (models.ts Result).
	if b, _ := json.Marshal(r); !strings.Contains(string(b), `"needsAccessibility":true`) {
		t.Errorf("JSON = %s, want \"needsAccessibility\":true", b)
	}
	if want := []string{"CopyFile:" + sample, "AccessibilityTrusted:false"}; !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestPasteTrustedOrder(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp}
	r, err := newPasteService(f).Paste(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || !r.Pasted {
		t.Errorf("result = %+v, want Copied+Pasted", r)
	}
	want := []string{"CopyFile:" + sample, "AccessibilityTrusted:false", "hide", "sleep:80ms", "FrontmostApp", "PasteShortcut"}
	if !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestPasteReactivatesPrevious(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: platform.App{BundleID: "com.other", PID: 7}, ActivateSetsFrontmost: true}
	r, err := newPasteService(f).Paste(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.Pasted {
		t.Errorf("result = %+v, want Pasted", r)
	}
	want := []string{"CopyFile:" + sample, "AccessibilityTrusted:false", "hide", "sleep:80ms", "FrontmostApp", "ActivateApp:42", "sleep:80ms", "FrontmostApp", "PasteShortcut"}
	if !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

// macOS 14+ often declines activation from a background app; ⌘V must not land in the wrong app.
func TestPasteDeclinedReactivationCopiesOnly(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: platform.App{BundleID: "com.other", PID: 7}}
	r, err := newPasteService(f).Paste(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.Copied || r.Pasted || !strings.Contains(r.Message, "TextEdit") {
		t.Errorf("result = %+v, want Copied, not Pasted, message naming TextEdit", r)
	}
	if slices.Contains(f.Calls, "PasteShortcut") {
		t.Errorf("⌘V posted after declined activation: calls = %q", f.Calls)
	}
}

func TestPasteCopyError(t *testing.T) {
	f := &platform.Fake{Trusted: true, CopyErr: errors.New("pasteboard write failed")}
	r, err := newPasteService(f).Paste(context.Background())
	if !errors.Is(err, f.CopyErr) {
		t.Errorf("err = %v, want %v", err, f.CopyErr)
	}
	if r.Copied || r.Pasted {
		t.Errorf("result = %+v on copy error", r)
	}
	if want := []string{"CopyFile:" + sample}; !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}

func TestPastePostError(t *testing.T) {
	f := &platform.Fake{Trusted: true, Frontmost: prevApp, PasteErr: errors.New("could not post")}
	r, err := newPasteService(f).Paste(context.Background())
	if !errors.Is(err, f.PasteErr) {
		t.Errorf("err = %v, want %v", err, f.PasteErr)
	}
	if !r.Copied || r.Pasted {
		t.Errorf("result = %+v, want Copied, not Pasted", r)
	}
}

func TestRequestAccessibilityOpensSettings(t *testing.T) {
	f := &platform.Fake{}
	ok, err := newPasteService(f).RequestAccessibility(context.Background())
	if err != nil || ok {
		t.Errorf("got (%v, %v), want (false, nil)", ok, err)
	}
	if want := []string{"AccessibilityTrusted:true", "OpenAccessibilitySettings"}; !slices.Equal(f.Calls, want) {
		t.Errorf("calls = %q, want %q", f.Calls, want)
	}
}
