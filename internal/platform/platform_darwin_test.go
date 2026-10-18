//go:build darwin && !server

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Uses a private named pasteboard; never touches the user's general pasteboard.
func TestCopyFileWritesFileURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.gif")
	if err := os.WriteFile(path, []byte("GIF89a"), 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := filepath.EvalSymlinks(path) // /var → /private/var, as NSURL reports it
	if err != nil {
		t.Fatal(err)
	}
	pb := fmt.Sprintf("com.kenscott.gifsmith.test.%d", time.Now().UnixNano())
	t.Cleanup(func() { releasePasteboard(pb) })

	// Twice: clearContents must leave exactly one URL.
	for range 2 {
		if err := copyFileToPasteboard(pb, path); err != nil {
			t.Fatal(err)
		}
	}
	paths, hasFileURL, err := pasteboardFileURLs(pb)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFileURL {
		t.Error("public.file-url missing from pasteboard types")
	}
	if len(paths) != 1 || paths[0] != path {
		t.Errorf("pasteboard paths = %q, want [%q]", paths, path)
	}
}

// Private named pasteboard only; OpenURL is never called from tests (it opens a real browser).
func TestCopyTextWritesString(t *testing.T) {
	pb := fmt.Sprintf("com.kenscott.gifsmith.test.%d", time.Now().UnixNano())
	t.Cleanup(func() { releasePasteboard(pb) })
	const want = "https://x/y.gif"
	if err := copyTextToPasteboard(pb, want); err != nil {
		t.Fatal(err)
	}
	if got, err := pasteboardString(pb); err != nil || got != want {
		t.Fatalf("pasteboard = %q, %v; want %q", got, err, want)
	}
	for _, bad := range []string{"bad\xff", "nul\x00byte"} {
		if err := copyTextToPasteboard(pb, bad); err == nil {
			t.Errorf("copyTextToPasteboard(%q) = nil, want an error", bad)
		}
	}
	if got, _ := pasteboardString(pb); got != want {
		t.Errorf("rejected text replaced the pasteboard: %q", got)
	}
}

func TestCopyFileMissingPath(t *testing.T) {
	if err := New().CopyFile(filepath.Join(t.TempDir(), "missing.gif")); err == nil {
		t.Fatal("CopyFile on a missing path returned nil")
	}
}

func TestFakeSetAppearanceRecords(t *testing.T) {
	f := &Fake{}
	if err := f.SetAppearance("dark"); err != nil {
		t.Fatal(err)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "SetAppearance:dark" {
		t.Errorf("Calls = %v, want [SetAppearance:dark]", f.Calls)
	}
}

// The mapping only; go test never calls the real SetAppearance (it would change NSApp for this process).
func TestAppearanceCode(t *testing.T) {
	for mode, want := range map[string]int{"": 0, "light": 1, "dark": 2} {
		if got, err := appearanceCode(mode); got != want || err != nil {
			t.Errorf("appearanceCode(%q) = %d, %v; want %d, nil", mode, got, err, want)
		}
	}
	if _, err := appearanceCode("neon"); err == nil || err.Error() != "unknown appearance" {
		t.Errorf("appearanceCode(neon) err = %v, want the fixed unknown appearance error", err)
	}
}
