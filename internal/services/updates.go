package services

import (
	"cmp"
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gifsmith/internal/update"
)

// Version is this build's version. 04-05's version_test keeps it equal to Info.plist
// CFBundleShortVersionString.
const Version = "1.0.0"

// UpdateStatus is what the Updates row renders. State: idle | checking | current | available |
// downloading | verifying | opened.
type UpdateStatus struct {
	State   string `json:"state"`
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Done    int64  `json:"done"`
	Total   int64  `json:"total"`
}

// UpdateService checks GitHub Releases for a newer Gifsmith (RELS-05). Every state change
// emits update:status. Exported methods are bound to JS; the automatic loop is a package func.
type UpdateService struct {
	Checker *update.Checker
	Current string
	Emit    func(name string, data ...any)
	Dir     string       // the app cache's updates folder
	Client  *http.Client // update.NewDownloadClient
	Run     update.Runner
	Open    func(string) error // platform OpenURL
	Now     func() time.Time   // nil: time.Now
	// OnAvailable gets the version whenever a check finds a newer release (tray item);
	// RevealFunc opens Settings at Updates. Both nil in tests.
	OnAvailable func(version string)
	RevealFunc  func()

	mu  sync.Mutex
	st  UpdateStatus
	rel update.Release // the available release; zero unless State is available or later
}

func (u *UpdateService) status() UpdateStatus {
	st := u.st
	st.State = cmp.Or(st.State, "idle")
	st.Current = u.Current
	return st
}

// set replaces the status and emits it; callers hold u.mu so emits keep their order.
func (u *UpdateService) set(st UpdateStatus) UpdateStatus {
	u.st = st
	st = u.status()
	if u.Emit != nil {
		u.Emit("update:status", st)
	}
	return st
}

func (u *UpdateService) Status(ctx context.Context) UpdateStatus {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.status()
}

func busy(state string) bool {
	return state == "checking" || state == "downloading" || state == "verifying"
}

// Check asks GitHub for the latest release. One check or download at a time: a call while one
// runs returns the current status. A failure restores the previous state.
func (u *UpdateService) Check(ctx context.Context) (UpdateStatus, error) {
	u.mu.Lock()
	if busy(u.st.State) {
		defer u.mu.Unlock()
		return u.status(), nil
	}
	prev := u.st
	u.set(UpdateStatus{State: "checking"})
	u.mu.Unlock()

	rel, found, err := u.Checker.Latest(ctx)
	newer := 0
	if found {
		newer, err = update.Compare(rel.Version, u.Current)
	}

	var avail string
	defer func() { // after the unlock below: the tray setter waits on the main thread
		if avail != "" && u.OnAvailable != nil {
			u.OnAvailable(avail)
		}
	}()
	u.mu.Lock()
	defer u.mu.Unlock()
	switch {
	case errors.Is(err, update.ErrNotInstallable):
		log.Printf("update: release v%s not installable: %v", rel.Version, err)
	case err != nil:
		log.Printf("update: check failed: %v", err)
		return u.set(prev), &SettingsError{Kind: "update_check_failed"}
	case !found:
		log.Printf("update: no published release")
	case newer > 0:
		log.Printf("update: %s available", rel.Version)
		u.rel, avail = rel, rel.Version
		return u.set(UpdateStatus{State: "available", Latest: rel.Version}), nil
	default:
		log.Printf("update: up to date")
	}
	u.rel = update.Release{}
	return u.set(UpdateStatus{State: "current"}), nil
}

// Download fetches the available release's DMG into Dir, verifies it (digest, signature, Team ID,
// notarization, quarantine) and only then opens it; a failed check deletes the file. It runs on
// its own 10-minute deadline, so a Settings reload does not abort it. Any error restores available.
func (u *UpdateService) Download(context.Context) (UpdateStatus, error) {
	u.mu.Lock()
	if busy(u.st.State) {
		defer u.mu.Unlock()
		return u.status(), nil
	}
	if u.st.State != "available" && u.st.State != "opened" {
		defer u.mu.Unlock()
		return u.status(), errInvalidValue
	}
	rel := u.rel
	u.set(UpdateStatus{State: "downloading", Latest: rel.Version, Total: rel.DMG.Size})
	u.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	fail := func(kind string) (UpdateStatus, error) {
		u.mu.Lock()
		defer u.mu.Unlock()
		return u.set(UpdateStatus{State: "available", Latest: rel.Version}), &SettingsError{Kind: kind}
	}
	if err := update.Clean(u.Dir); err != nil {
		log.Printf("update: clean: %v", err)
	}
	path, err := update.Download(ctx, u.Client, rel.DMG, u.Dir, func(done, total int64) {
		u.mu.Lock()
		defer u.mu.Unlock()
		u.set(UpdateStatus{State: "downloading", Latest: rel.Version, Done: done, Total: total})
	})
	if errors.Is(err, update.ErrDigest) {
		log.Printf("update: verification failed: digest")
		return fail("update_verify_failed")
	}
	if err != nil {
		log.Printf("update: download failed")
		return fail("update_download_failed")
	}

	u.mu.Lock()
	u.set(UpdateStatus{State: "verifying", Latest: rel.Version})
	u.mu.Unlock()
	now := time.Now
	if u.Now != nil {
		now = u.Now
	}
	if err := update.Verify(ctx, u.Run, path, update.TeamID, now()); err != nil {
		log.Printf("update: verification failed: %v", err)
		_ = os.Remove(path)
		return fail("update_verify_failed")
	}
	if err := u.Open((&url.URL{Scheme: "file", Path: path}).String()); err != nil {
		log.Printf("update: open failed") // the verified file stays for a retry
		return fail("update_open_failed")
	}
	log.Printf("update: opened %s", filepath.Base(path))
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.set(UpdateStatus{State: "opened", Latest: rel.Version}), nil
}

// Reveal opens Settings at General → Updates (the panel toast's View action).
func (u *UpdateService) Reveal() {
	if u.RevealFunc != nil {
		u.RevealFunc()
	}
}

// RunAutoUpdates checks after delay and then every every while enabled() holds; failures only
// log (Check does). A package func, not a method, so Wails never binds it to JS.
func RunAutoUpdates(ctx context.Context, u *UpdateService, enabled func() bool, delay, every time.Duration) {
	runAutoUpdates(ctx, u, enabled, delay, every, time.After)
}

// runAutoUpdates takes the timer as after, so tests drive it with a channel instead of sleeping.
func runAutoUpdates(ctx context.Context, u *UpdateService, enabled func() bool, delay, every time.Duration, after func(time.Duration) <-chan time.Time) {
	wait := delay
	for {
		select {
		case <-ctx.Done():
			return
		case <-after(wait):
		}
		if enabled() {
			_, _ = u.Check(ctx)
		}
		wait = every
	}
}
