package services

import (
	"bytes"
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
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gifsmith/internal/platform"
	"gifsmith/internal/update"
)

// updateFixture serves the latest-release route; tag "" answers 404, status overrides 200.
type updateFixture struct {
	svc    *UpdateService
	hits   *atomic.Int32
	mu     *sync.Mutex
	events *[]UpdateStatus
}

func (f updateFixture) states() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var s []string
	for _, e := range *f.events {
		s = append(s, e.State)
	}
	return s
}

func releaseJSON(srvURL, tag string, assetOK bool) map[string]any {
	v := strings.TrimPrefix(tag, "v")
	name := update.DMGAssetName(v)
	if !assetOK {
		name += ".zip"
	}
	return map[string]any{"tag_name": tag, "assets": []map[string]any{{
		"name": name, "size": 4, "state": "uploaded",
		"digest":               "sha256:" + strings.Repeat("ab", 32),
		"browser_download_url": srvURL + "/download/" + tag + "/" + name,
	}}}
}

func newUpdates(t *testing.T, h func(w http.ResponseWriter, r *http.Request, base string)) updateFixture {
	t.Helper()
	var hits atomic.Int32
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/scottkw/gifsmith/releases/latest" {
			hits.Add(1)
		}
		h(w, r, srv.URL)
	}))
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	allow := func(u *url.URL) bool { return u.Scheme == "http" && u.Host == host }
	var mu sync.Mutex
	var events []UpdateStatus
	svc := &UpdateService{
		Checker: &update.Checker{API: srv.URL, Repo: update.Repo, HTTP: srv.Client(), Allow: allow},
		Current: "1.0.0",
		Emit: func(name string, data ...any) {
			if name != "update:status" {
				t.Errorf("event %s", name)
				return
			}
			mu.Lock()
			events = append(events, data[0].(UpdateStatus))
			mu.Unlock()
		},
	}
	return updateFixture{svc: svc, hits: &hits, mu: &mu, events: &events}
}

func serveRelease(tag string, assetOK bool) func(http.ResponseWriter, *http.Request, string) {
	return func(w http.ResponseWriter, r *http.Request, base string) {
		if tag == "" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(releaseJSON(base, tag, assetOK))
	}
}

func TestUpdateCheckAvailable(t *testing.T) {
	f := newUpdates(t, serveRelease("v1.1.0", true))
	st, err := f.svc.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := (UpdateStatus{State: "available", Current: "1.0.0", Latest: "1.1.0"}); st != want {
		t.Errorf("Check = %+v, want %+v", st, want)
	}
	if got := strings.Join(f.states(), ","); got != "checking,available" {
		t.Errorf("emitted %s", got)
	}
	if got := f.svc.Status(context.Background()); got.State != "available" || got.Latest != "1.1.0" {
		t.Errorf("Status = %+v", got)
	}
}

func TestUpdateCheckCurrent(t *testing.T) {
	for name, h := range map[string]func(http.ResponseWriter, *http.Request, string){
		"same version":    serveRelease("v1.0.0", true),
		"older":           serveRelease("v0.9.0", true),
		"404":             serveRelease("", true),
		"not installable": serveRelease("v1.1.0", false),
	} {
		t.Run(name, func(t *testing.T) {
			buf := captureLog(t)
			f := newUpdates(t, h)
			st, err := f.svc.Check(context.Background())
			if err != nil || st.State != "current" || st.Current != "1.0.0" || st.Latest != "" {
				t.Errorf("Check = %+v, %v; want current", st, err)
			}
			if name == "not installable" && !strings.Contains(buf.String(), "update: release v1.1.0 not installable") {
				t.Errorf("log %q", buf.String())
			}
			if name == "404" && !strings.Contains(buf.String(), "update: no published release") {
				t.Errorf("log %q", buf.String())
			}
		})
	}
}

func TestUpdateCheckFailed(t *testing.T) {
	buf := captureLog(t)
	f := newUpdates(t, func(w http.ResponseWriter, r *http.Request, _ string) { w.WriteHeader(503) })
	st, err := f.svc.Check(context.Background())
	wantSettingsErr(t, err, "update_check_failed")
	if st.State != "idle" {
		t.Errorf("state after failure = %q, want idle", st.State)
	}
	if got := strings.Join(f.states(), ","); got != "checking,idle" {
		t.Errorf("emitted %s", got)
	}
	if !strings.Contains(buf.String(), "update: check failed") {
		t.Errorf("log %q", buf.String())
	}
}

func TestUpdateCheckSingleFlight(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{}, 1)
	f := newUpdates(t, func(w http.ResponseWriter, r *http.Request, base string) {
		entered <- struct{}{}
		<-release
		_ = json.NewEncoder(w).Encode(releaseJSON(base, "v1.1.0", true))
	})
	done := make(chan UpdateStatus)
	go func() {
		st, _ := f.svc.Check(context.Background())
		done <- st
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("the first Check never reached the server")
	}
	st, err := f.svc.Check(context.Background())
	if err != nil || st.State != "checking" {
		t.Errorf("second Check = %+v, %v; want checking, nil", st, err)
	}
	close(release)
	if st := <-done; st.State != "available" {
		t.Errorf("first Check = %+v", st)
	}
	if n := f.hits.Load(); n != 1 {
		t.Errorf("%d requests, want 1", n)
	}
}

func TestMarshalErrorUpdate(t *testing.T) {
	if got := string(MarshalError(&SettingsError{Kind: "update_check_failed"})); got != `{"kind":"update_check_failed"}` {
		t.Errorf("MarshalError = %s", got)
	}
	for _, k := range []string{"update_check_failed", "update_download_failed", "update_verify_failed", "update_open_failed"} {
		if _, ok := settingsText[k]; !ok {
			t.Errorf("no settingsText for %s", k)
		}
	}
}

var dmgBody = bytes.Repeat([]byte("dmg!"), 40<<10) // 160 KiB

// serveDownloadable serves v1.1.0 whose asset digest is that of digestOf; short makes the
// download stop halfway with the full Content-Length.
func serveDownloadable(digestOf []byte, short bool, downloads *atomic.Int32) func(http.ResponseWriter, *http.Request, string) {
	sum := sha256.Sum256(digestOf)
	return func(w http.ResponseWriter, r *http.Request, base string) {
		name := update.DMGAssetName("1.1.0")
		if strings.HasPrefix(r.URL.Path, "/download/") {
			downloads.Add(1)
			w.Header().Set("Content-Length", strconv.Itoa(len(dmgBody)))
			if short {
				_, _ = w.Write(dmgBody[:len(dmgBody)/2])
				return
			}
			_, _ = w.Write(dmgBody)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "v1.1.0", "assets": []map[string]any{{
			"name": name, "size": len(dmgBody), "state": "uploaded",
			"digest":               "sha256:" + hex.EncodeToString(sum[:]),
			"browser_download_url": base + "/download/v1.1.0/" + name,
		}}})
	}
}

// okRun answers like codesign/spctl/xattr on a good notarized DMG, failing at the named tool.
func okRun(failAt string) update.Runner {
	return func(_ context.Context, name string, args ...string) ([]byte, error) {
		if filepath.Base(name) == failAt {
			return []byte("rejected"), errors.New("exit status 3")
		}
		switch filepath.Base(name) {
		case "codesign":
			return []byte("TeamIdentifier=" + update.TeamID + "\n"), nil
		case "spctl":
			return []byte("x.dmg: accepted\nsource=Notarized Developer ID\n"), nil
		}
		return nil, nil
	}
}

type downloadFixture struct {
	updateFixture
	plat      *platform.Fake
	dir       string
	downloads *atomic.Int32
}

func newDownload(t *testing.T, digestOf []byte, short bool, failAt string) downloadFixture {
	t.Helper()
	var downloads atomic.Int32
	f := newUpdates(t, serveDownloadable(digestOf, short, &downloads))
	plat := &platform.Fake{}
	dir := filepath.Join(t.TempDir(), "updates")
	f.svc.Dir, f.svc.Client, f.svc.Run, f.svc.Open = dir, update.NewDownloadClient(f.svc.Checker.Allow), okRun(failAt), plat.OpenURL
	if _, err := f.svc.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	*f.events = nil
	f.mu.Unlock()
	return downloadFixture{f, plat, dir, &downloads}
}

func (f downloadFixture) opens() []string {
	var o []string
	for _, c := range f.plat.Calls {
		if strings.HasPrefix(c, "OpenURL:") {
			o = append(o, c)
		}
	}
	return o
}

func TestUpdateDownloadOpens(t *testing.T) {
	f := newDownload(t, dmgBody, false, "")
	st, err := f.svc.Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.State != "opened" || st.Latest != "1.1.0" {
		t.Errorf("Download = %+v", st)
	}
	path := filepath.Join(f.dir, "Gifsmith-1.1.0-arm64.dmg")
	if got := f.opens(); !slices.Equal(got, []string{"OpenURL:file://" + path}) {
		t.Errorf("opens %v", got)
	}
	states := f.states()
	if len(states) < 3 {
		t.Fatalf("emitted %v", states)
	}
	if states[0] != "downloading" || states[len(states)-2] != "verifying" || states[len(states)-1] != "opened" {
		t.Errorf("emitted %v", states)
	}
	f.mu.Lock()
	last := (*f.events)[len(states)-3] // the final progress emit
	f.mu.Unlock()
	if last.State != "downloading" || last.Done != int64(len(dmgBody)) || last.Total != int64(len(dmgBody)) {
		t.Errorf("last progress %+v", last)
	}
}

func TestUpdateDownloadFailures(t *testing.T) {
	for _, c := range []struct {
		name     string
		digestOf []byte
		short    bool
		failAt   string
		openErr  error
		kind     string
		fileKept bool
	}{
		{name: "verify fails", digestOf: dmgBody, failAt: "spctl", kind: "update_verify_failed"},
		{name: "digest", digestOf: dmgBody[1:], kind: "update_verify_failed"},
		{name: "network", digestOf: dmgBody, short: true, kind: "update_download_failed"},
		{name: "open fails", digestOf: dmgBody, openErr: errors.New("no handler"), kind: "update_open_failed", fileKept: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newDownload(t, c.digestOf, c.short, c.failAt)
			f.plat.OpenErr = c.openErr
			st, err := f.svc.Download(context.Background())
			wantSettingsErr(t, err, c.kind)
			if st.State != "available" || f.svc.Status(context.Background()).State != "available" {
				t.Errorf("state %+v, want available", st)
			}
			if got := f.states(); len(got) == 0 || got[len(got)-1] != "available" {
				t.Errorf("last emit %v", got)
			}
			_, statErr := os.Stat(filepath.Join(f.dir, "Gifsmith-1.1.0-arm64.dmg"))
			if kept := statErr == nil; kept != c.fileKept {
				t.Errorf("file kept %v, want %v", kept, c.fileKept)
			}
			if wantOpens := map[bool]int{false: 0, true: 1}[c.openErr != nil]; len(f.opens()) != wantOpens {
				t.Errorf("opens %v", f.opens())
			}
		})
	}
}

func TestUpdateDownloadNotAvailable(t *testing.T) {
	var downloads atomic.Int32
	f := newUpdates(t, serveDownloadable(dmgBody, false, &downloads))
	f.svc.Dir, f.svc.Client, f.svc.Run, f.svc.Open = t.TempDir(), http.DefaultClient, okRun(""), (&platform.Fake{}).OpenURL
	_, err := f.svc.Download(context.Background())
	wantSettingsErr(t, err, "invalid_value")
	if downloads.Load() != 0 || f.hits.Load() != 0 {
		t.Error("a request was sent")
	}
}
