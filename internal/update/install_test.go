package update

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var payload = bytes.Repeat([]byte("gifsmith"), 64<<10) // 512 KiB

func sum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// assetServer serves body at /x.dmg with a correct Content-Length unless lie is set.
func assetServer(t *testing.T, body []byte) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

func anyAllow(*url.URL) bool { return true }

func noFiles(t *testing.T, dir string) {
	t.Helper()
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("files left: %v", names)
	}
}

func TestDownload(t *testing.T) {
	srv, _ := assetServer(t, payload)
	dir := filepath.Join(t.TempDir(), "updates")
	var mu sync.Mutex
	var calls [][2]int64
	a := Asset{Name: "Gifsmith-1.1.0-arm64.dmg", URL: srv.URL + "/x.dmg", Size: int64(len(payload)), SHA256: sum(payload)}
	path, err := Download(context.Background(), NewDownloadClient(anyAllow), a, dir, func(done, total int64) {
		mu.Lock()
		calls = append(calls, [2]int64{done, total})
		mu.Unlock()
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, a.Name) {
		t.Errorf("path = %s", path)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("contents differ (%v)", err)
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o600 {
		t.Errorf("mode %v", fi.Mode().Perm())
	}
	if fi, _ := os.Stat(dir); fi.Mode().Perm() != 0o700 {
		t.Errorf("dir mode %v", fi.Mode().Perm())
	}
	if _, err := os.Stat(path + ".part"); !os.IsNotExist(err) {
		t.Error(".part left behind")
	}
	if len(calls) == 0 || calls[len(calls)-1] != [2]int64{a.Size, a.Size} {
		t.Fatalf("progress %v", calls)
	}
	if !slices.IsSortedFunc(calls, func(x, y [2]int64) int { return int(x[0] - y[0]) }) {
		t.Errorf("progress decreased: %v", calls)
	}
}

func TestDownloadDigestMismatch(t *testing.T) {
	srv, _ := assetServer(t, payload)
	dir := t.TempDir()
	a := Asset{Name: "Gifsmith-1.1.0-arm64.dmg", URL: srv.URL, Size: int64(len(payload)), SHA256: sum(payload[1:])}
	if _, err := Download(context.Background(), NewDownloadClient(anyAllow), a, dir, nil); !errors.Is(err, ErrDigest) {
		t.Fatalf("err = %v, want ErrDigest", err)
	}
	noFiles(t, dir)
}

func TestDownloadSize(t *testing.T) {
	srv, n := assetServer(t, payload)
	dir := t.TempDir()
	a := Asset{Name: "Gifsmith-1.1.0-arm64.dmg", URL: srv.URL, Size: MaxDMGBytes + 1, SHA256: sum(payload)}
	if _, err := Download(context.Background(), NewDownloadClient(anyAllow), a, dir, nil); !errors.Is(err, ErrTooLarge) {
		t.Errorf("err = %v, want ErrTooLarge", err)
	}
	if n.Load() != 0 {
		t.Error("a request was sent for an oversized asset")
	}
	for _, size := range []int64{int64(len(payload)) - 1, int64(len(payload)) + 1} {
		a := Asset{Name: "Gifsmith-1.1.0-arm64.dmg", URL: srv.URL, Size: size, SHA256: sum(payload)}
		if _, err := Download(context.Background(), NewDownloadClient(anyAllow), a, dir, nil); err == nil {
			t.Errorf("size %d of %d: want an error", size, len(payload))
		}
		noFiles(t, dir)
	}
}

func TestDownloadBadName(t *testing.T) {
	srv, n := assetServer(t, payload)
	dir := t.TempDir()
	a := Asset{Name: "../Gifsmith-1.1.0-arm64.dmg", URL: srv.URL, Size: int64(len(payload)), SHA256: sum(payload)}
	if _, err := Download(context.Background(), NewDownloadClient(anyAllow), a, dir, nil); err == nil || n.Load() != 0 {
		t.Errorf("err = %v, requests %d", err, n.Load())
	}
}

func TestDownloadShortBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload[:1000])
	}))
	defer srv.Close()
	dir := t.TempDir()
	a := Asset{Name: "Gifsmith-1.1.0-arm64.dmg", URL: srv.URL, Size: int64(len(payload)), SHA256: sum(payload)}
	if _, err := Download(context.Background(), NewDownloadClient(anyAllow), a, dir, nil); err == nil || errors.Is(err, ErrDigest) {
		t.Errorf("err = %v, want a transfer error", err)
	}
	noFiles(t, dir)
}

func TestDownloadRedirects(t *testing.T) {
	other, n := assetServer(t, payload)
	redirect := func(to string) *httptest.Server {
		srv := httptest.NewServer(http.RedirectHandler(to, http.StatusFound))
		t.Cleanup(srv.Close)
		return srv
	}
	a := Asset{Name: "Gifsmith-1.1.0-arm64.dmg", Size: int64(len(payload)), SHA256: sum(payload)}
	dir := t.TempDir()
	// The first hop is allowed; the redirect target's host is not.
	first := redirect(other.URL + "/x.dmg")
	firstHost := strings.TrimPrefix(first.URL, "http://")
	a.URL = first.URL
	allowFirst := func(u *url.URL) bool { return u.Host == firstHost }
	if _, err := Download(context.Background(), NewDownloadClient(allowFirst), a, dir, nil); err == nil {
		t.Error("redirect to a refused host: want an error")
	}
	// GitHubAllow refuses an https→http redirect.
	a.URL = redirect("http://objects.githubusercontent.com/x").URL
	if _, err := Download(context.Background(), NewDownloadClient(GitHubAllow), a, dir, nil); err == nil {
		t.Error("redirect to http: want an error")
	}
	if n.Load() != 0 {
		t.Errorf("the refused host saw %d requests", n.Load())
	}
	noFiles(t, dir)
}

// fakeRun answers like codesign/spctl/xattr on a good notarized DMG; fail names a step to break.
type fakeRun struct {
	mu    sync.Mutex
	calls []string
	fail  string // "codesign", "dvv", "spctl", "xattr"
	team  string // -dvv output
	spctl string // spctl output
}

func (f *fakeRun) run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
	step := filepath.Base(name)
	if slices.Contains(args, "-dvv") {
		step = "dvv"
	}
	if step == f.fail {
		return []byte("failed"), errors.New("exit status 1")
	}
	switch step {
	case "dvv":
		return []byte("Executable=/x\n" + f.team + "\nTimestamp=now\n"), nil
	case "spctl":
		return []byte(f.spctl), nil
	}
	return nil, nil
}

const goodSpctl = "/tmp/x.dmg: accepted\nsource=Notarized Developer ID\norigin=Developer ID Application: Ken Scott (S2K7P43927)\n"

func TestVerify(t *testing.T) {
	now := time.Unix(0x6720_0000, 0)
	f := &fakeRun{team: "TeamIdentifier=S2K7P43927", spctl: goodSpctl}
	if err := Verify(context.Background(), f.run, "/c/x.dmg", TeamID, now); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/usr/bin/codesign --verify --strict /c/x.dmg",
		"/usr/bin/codesign -dvv /c/x.dmg",
		"/usr/sbin/spctl -a -t open --context context:primary-signature -vv /c/x.dmg",
		fmt.Sprintf("/usr/bin/xattr -w com.apple.quarantine 0081;%x;Gifsmith; /c/x.dmg", now.Unix()),
	}
	if !slices.Equal(f.calls, want) {
		t.Errorf("calls\n%q\nwant\n%q", f.calls, want)
	}
}

func TestVerifyFailures(t *testing.T) {
	good := fakeRun{team: "TeamIdentifier=S2K7P43927", spctl: goodSpctl}
	cases := map[string]*fakeRun{
		"codesign fails":    {fail: "codesign", team: good.team, spctl: good.spctl},
		"no team line":      {team: "Authority=Developer ID Application", spctl: good.spctl},
		"other team":        {team: "TeamIdentifier=ABCDE12345", spctl: good.spctl},
		"team prefix":       {team: "TeamIdentifier=S2K7P439270", spctl: good.spctl},
		"spctl fails":       {fail: "spctl", team: good.team, spctl: good.spctl},
		"spctl rejected":    {team: good.team, spctl: "/tmp/x.dmg: rejected\nsource=Notarized Developer ID\n"},
		"not notarized":     {team: good.team, spctl: "/tmp/x.dmg: accepted\nsource=Developer ID\n"},
		"notarized prefix":  {team: good.team, spctl: "/tmp/x.dmg: accepted\nsource=Notarized Developer IDx\n"},
		"quarantine fails":  {fail: "xattr", team: good.team, spctl: good.spctl},
		"dvv command fails": {fail: "dvv", team: good.team, spctl: good.spctl},
	}
	for name, f := range cases {
		if err := Verify(context.Background(), f.run, "/c/x.dmg", TeamID, time.Now()); !errors.Is(err, ErrVerify) {
			t.Errorf("%s: err = %v, want ErrVerify", name, err)
		}
	}
}

func TestClean(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"Gifsmith-1.1.0-arm64.dmg", "Gifsmith-1.2.0-arm64.dmg.part", "keep.txt", "Gifsmith-1.1.0-x86_64.dmg"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "Gifsmith-9-arm64.dmg"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := Clean(dir); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	var left []string
	for _, e := range entries {
		left = append(left, e.Name())
	}
	if want := []string{"Gifsmith-1.1.0-x86_64.dmg", "Gifsmith-9-arm64.dmg", "keep.txt"}; !slices.Equal(left, want) {
		t.Errorf("left %v, want %v", left, want)
	}
	if err := Clean(filepath.Join(dir, "missing")); err != nil {
		t.Errorf("missing dir: %v", err)
	}
}
