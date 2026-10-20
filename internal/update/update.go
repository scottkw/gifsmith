// Package update is the Wails-free GitHub Releases update check, download and verification; RELS-05/06.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	Repo        = "scottkw/gifsmith"
	GitHubAPI   = "https://api.github.com"
	TeamID      = "S2K7P43927"
	MaxDMGBytes = 200 << 20
	maxJSON     = 1 << 20
)

var (
	ErrNotInstallable = errors.New("release is not installable")
	ErrDigest         = errors.New("download digest mismatch")
	ErrTooLarge       = errors.New("download too large")
	ErrVerify         = errors.New("download verification failed")
)

var (
	versionRE = regexp.MustCompile(`^v?(0|[1-9][0-9]{0,5})\.(0|[1-9][0-9]{0,5})\.(0|[1-9][0-9]{0,5})$`)
	sha256RE  = regexp.MustCompile(`^sha256:([0-9a-f]{64})$`)
)

// ParseVersion accepts only MAJOR.MINOR.PATCH with an optional leading v (no pre-release, no leading zeros).
func ParseVersion(s string) ([3]int, error) {
	m := versionRE.FindStringSubmatch(s)
	if m == nil {
		return [3]int{}, fmt.Errorf("not a version: %q", s)
	}
	var v [3]int
	for i := range v {
		v[i], _ = strconv.Atoi(m[i+1]) // at most 6 digits: cannot fail
	}
	return v, nil
}

// Compare returns -1, 0 or 1 as a is older than, equal to or newer than b.
func Compare(a, b string) (int, error) {
	va, err := ParseVersion(a)
	if err != nil {
		return 0, err
	}
	vb, err := ParseVersion(b)
	if err != nil {
		return 0, err
	}
	for i := range va {
		if va[i] != vb[i] {
			if va[i] < vb[i] {
				return -1, nil
			}
			return 1, nil
		}
	}
	return 0, nil
}

// DMGAssetName is the only release asset Gifsmith downloads (04-05's DMG_NAME).
func DMGAssetName(v string) string { return "Gifsmith-" + v + "-arm64.dmg" }

// GitHubAllow allows https to this repo's release downloads and the two asset CDN hosts
// GitHub redirects them to; hosts are compared exactly, never by suffix.
func GitHubAllow(u *url.URL) bool {
	if u.Scheme != "https" || u.User != nil || strings.Contains(u.Path, "..") {
		return false
	}
	switch u.Host {
	case "github.com":
		return strings.HasPrefix(u.Path, "/"+Repo+"/releases/download/")
	case "objects.githubusercontent.com", "release-assets.githubusercontent.com":
		return true
	}
	return false
}

// Asset is the validated DMG of a release; SHA256 is lowercase hex.
type Asset struct {
	Name, URL string
	Size      int64
	SHA256    string
}

type Release struct {
	Version string // without the leading v
	DMG     Asset
}

// Checker asks GitHub for the latest release. It sends only Accept, User-Agent and,
// after a 200, If-None-Match; a 304 reuses the cached body.
type Checker struct {
	API, Repo string
	HTTP      *http.Client
	Allow     func(*url.URL) bool // the download URL must pass it

	mu   sync.Mutex // one request at a time; guards the cache
	etag string
	body []byte
}

// Latest reports the newest published release. found is false for 404 (nothing published),
// drafts, prereleases and tags that are not v + MAJOR.MINOR.PATCH. A release without a valid
// DMG asset returns its version with ErrNotInstallable.
func (c *Checker) Latest(ctx context.Context) (Release, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.API+"/repos/"+c.Repo+"/releases/latest", nil)
	if err != nil {
		return Release{}, false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Gifsmith")
	if c.etag != "" {
		req.Header.Set("If-None-Match", c.etag)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Release{}, false, err
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxJSON+1))
		if err != nil {
			return Release{}, false, err
		}
		if len(body) > maxJSON {
			return Release{}, false, errors.New("release JSON over 1 MiB")
		}
		c.etag, c.body = resp.Header.Get("ETag"), body
		return c.parse(body)
	case http.StatusNotModified:
		if c.body == nil {
			return Release{}, false, errors.New("304 without a cached release")
		}
		return c.parse(c.body)
	case http.StatusNotFound:
		return Release{}, false, nil
	}
	return Release{}, false, fmt.Errorf("status %d", resp.StatusCode)
}

func (c *Checker) parse(body []byte) (Release, bool, error) {
	var r struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name   string `json:"name"`
			Size   int64  `json:"size"`
			State  string `json:"state"`
			URL    string `json:"browser_download_url"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return Release{}, false, err
	}
	if r.Draft || r.Prerelease || !versionRE.MatchString(r.Tag) {
		return Release{}, false, nil
	}
	rel := Release{Version: strings.TrimPrefix(r.Tag, "v")}
	name := DMGAssetName(rel.Version) // built from the validated version, never taken from JSON
	for _, a := range r.Assets {
		if a.Name != name {
			continue
		}
		m := sha256RE.FindStringSubmatch(a.Digest)
		u, err := url.Parse(a.URL)
		switch {
		case a.State != "uploaded":
			return rel, false, fmt.Errorf("%w: asset state %q", ErrNotInstallable, a.State)
		case a.Size < 1 || a.Size > MaxDMGBytes:
			return rel, false, fmt.Errorf("%w: asset size %d", ErrNotInstallable, a.Size)
		case m == nil:
			return rel, false, fmt.Errorf("%w: no sha256 digest", ErrNotInstallable)
		case err != nil || !c.Allow(u):
			return rel, false, fmt.Errorf("%w: download URL not allowed", ErrNotInstallable)
		}
		rel.DMG = Asset{Name: name, URL: a.URL, Size: a.Size, SHA256: m[1]}
		return rel, true, nil
	}
	return rel, false, fmt.Errorf("%w: no %s asset", ErrNotInstallable, name)
}
