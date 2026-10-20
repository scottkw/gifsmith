package services

import (
	"context"
	"sync"

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

// UpdateService checks GitHub Releases for a newer Gifsmith (RELS-05).
type UpdateService struct {
	Checker *update.Checker
	Current string
	Emit    func(name string, data ...any)

	mu  sync.Mutex
	st  UpdateStatus
	rel update.Release
}

func (u *UpdateService) Status(ctx context.Context) UpdateStatus { return UpdateStatus{} }

func (u *UpdateService) Check(ctx context.Context) (UpdateStatus, error) { return UpdateStatus{}, nil }
