//go:build darwin && !server

package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gifsmith/internal/platform"
)

// Env-gated checks on the real release artefacts; they skip in the normal suite. Plan 04-06
// runs both (verify only) against the notarized DMG and the draft release JSON.

// copyTemp copies src into a fresh temp dir, optionally flipping the middle byte.
func copyTemp(t *testing.T, src string, flip bool) string {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if flip {
		b[len(b)/2] ^= 0xFF
	}
	dst := filepath.Join(t.TempDir(), filepath.Base(src))
	if err := os.WriteFile(dst, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return dst
}

// UPDATE_REAL_DMG={path}: the real tools accept the DMG and reject a byte-flipped copy.
// UPDATE_REAL_OPEN=1 also opens the verified copy in Finder (only the user sets this).
func TestRealDMG(t *testing.T) {
	src := os.Getenv("UPDATE_REAL_DMG")
	if src == "" {
		t.Skip("UPDATE_REAL_DMG not set")
	}
	ctx := context.Background()
	good := copyTemp(t, src, false)
	if err := Verify(ctx, ExecRunner, good, TeamID, time.Now()); err != nil {
		t.Fatalf("Verify(real DMG) = %v", err)
	}
	if err := Verify(ctx, ExecRunner, copyTemp(t, src, true), TeamID, time.Now()); !errors.Is(err, ErrVerify) {
		t.Errorf("Verify(byte-flipped DMG) = %v, want ErrVerify", err)
	}
	if os.Getenv("UPDATE_REAL_OPEN") == "1" {
		if err := platform.New().OpenURL((&url.URL{Scheme: "file", Path: good}).String()); err != nil {
			t.Errorf("OpenURL: %v", err)
		}
	}
}

// UPDATE_RELEASE_JSON={file} (the REST JSON of one release) and UPDATE_DMG_PATH={dmg}: served as the
// latest release with draft forced false, the checker offers exactly that DMG and its digest.
func TestReleaseJSON(t *testing.T) {
	file, dmg := os.Getenv("UPDATE_RELEASE_JSON"), os.Getenv("UPDATE_DMG_PATH")
	if file == "" || dmg == "" {
		t.Skip("UPDATE_RELEASE_JSON or UPDATE_DMG_PATH not set")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var rel map[string]any
	if err := json.Unmarshal(raw, &rel); err != nil {
		t.Fatal(err)
	}
	rel["draft"] = false // 04-06 feeds the draft; a published release differs only in this flag
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(rel)
	}))
	defer srv.Close()
	got, found, err := (&Checker{API: srv.URL, Repo: Repo, HTTP: srv.Client(), Allow: GitHubAllow}).Latest(context.Background())
	if err != nil || !found {
		t.Fatalf("Latest = %v, %v", found, err)
	}
	tag, _ := rel["tag_name"].(string)
	if got.Version != strings.TrimPrefix(tag, "v") || got.DMG.Name != DMGAssetName(got.Version) {
		t.Errorf("release %+v for tag %q", got, tag)
	}
	b, err := os.ReadFile(dmg)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if got.DMG.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("asset digest %s, local DMG %x", got.DMG.SHA256, sum)
	}
}
