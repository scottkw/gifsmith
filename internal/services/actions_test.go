package services

import (
	"context"
	"errors"
	"slices"
	"testing"

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
