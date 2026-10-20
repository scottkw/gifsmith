package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
