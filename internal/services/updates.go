package services

import (
	"cmp"
	"context"
	"errors"
	"log"
	"net/http"
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
		u.rel = rel
		return u.set(UpdateStatus{State: "available", Latest: rel.Version}), nil
	default:
		log.Printf("update: up to date")
	}
	u.rel = update.Release{}
	return u.set(UpdateStatus{State: "current"}), nil
}

func (u *UpdateService) Download(ctx context.Context) (UpdateStatus, error) {
	return UpdateStatus{}, nil
}
