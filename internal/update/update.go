// Package update is the Wails-free GitHub Releases update check, download and verification; RELS-05/06.
package update

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
)

const (
	Repo        = "scottkw/gifsmith"
	GitHubAPI   = "https://api.github.com"
	TeamID      = "S2K7P43927"
	MaxDMGBytes = 200 << 20
)

var (
	ErrNotInstallable = errors.New("release is not installable")
	ErrDigest         = errors.New("download digest mismatch")
	ErrTooLarge       = errors.New("download too large")
	ErrVerify         = errors.New("download verification failed")
)

func ParseVersion(s string) ([3]int, error) { return [3]int{}, nil }

func Compare(a, b string) (int, error) { return 0, nil }

func DMGAssetName(v string) string { return "" }

func GitHubAllow(u *url.URL) bool { return false }

type Asset struct {
	Name, URL string
	Size      int64
	SHA256    string
}

type Release struct {
	Version string
	DMG     Asset
}

type Checker struct {
	API, Repo string
	HTTP      *http.Client
	Allow     func(*url.URL) bool

	mu   sync.Mutex
	etag string
	body []byte
}

func (c *Checker) Latest(ctx context.Context) (Release, bool, error) { return Release{}, false, nil }
