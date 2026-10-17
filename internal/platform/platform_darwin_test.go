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
	for i := 0; i < 2; i++ {
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

func TestCopyFileMissingPath(t *testing.T) {
	if err := New().CopyFile(filepath.Join(t.TempDir(), "missing.gif")); err == nil {
		t.Fatal("CopyFile on a missing path returned nil")
	}
}
