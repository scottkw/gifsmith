package update

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

func NewDownloadClient(allow func(*url.URL) bool) *http.Client { return &http.Client{} }

func Download(ctx context.Context, client *http.Client, a Asset, dir string, progress func(done, total int64)) (string, error) {
	return "", nil
}

type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

func ExecRunner(ctx context.Context, name string, args ...string) ([]byte, error) { return nil, nil }

func Verify(ctx context.Context, run Runner, path, teamID string, now time.Time) error { return nil }

func Clean(dir string) error { return nil }
