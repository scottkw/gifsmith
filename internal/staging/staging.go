// Package staging downloads the selected GIF (or a Clip's MP4) to the one file the pasteboard points at.
// There is no media cache: Sweep removes every older staged file after a successful copy.
package staging

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/httpx"
)

// Fixed, URL-free error text: these reach the webview through Wails.
var (
	ErrNotAllowed = errors.New("this GIF's address is not allowed")
	ErrDownload   = errors.New("the GIF download failed")
	ErrNotGIF     = errors.New("the download is not a GIF")
	ErrNotMP4     = errors.New("the download is not an MP4")
	ErrTooLarge   = errors.New("the GIF is too large to copy")
)

const (
	defaultMax   = 50 << 20
	fetchTimeout = 45 * time.Second
	slugMax      = 80
)

type Store struct {
	Dir      string
	HTTP     *http.Client
	Allow    func(p provider.ID, u *url.URL) bool // nil = VendorAllow
	maxBytes int64                                // 0 = 50 MB (tests lower it)
}

// Fetch downloads g.DownloadURL to <Dir>/<sha1(provider:id)[:16]>/<slug>.gif (.mp4 for a Clip) and
// returns that path. The first bytes must be the GIF magic, or for a Clip an MP4 ftyp box.
// Transport errors are never wrapped: their text holds the URL.
func (s *Store) Fetch(ctx context.Context, g provider.Gif) (string, error) {
	allow := s.Allows
	u, err := url.Parse(g.DownloadURL)
	if err != nil || !allow(g.Provider, u) {
		return "", ErrNotAllowed
	}
	limit := s.maxBytes
	if limit == 0 {
		limit = defaultMax
	}
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", ErrNotAllowed
	}
	req.Header.Set("User-Agent", httpx.UserAgent)
	hc := *s.HTTP // every redirect hop must pass the same allowlist
	hc.CheckRedirect = func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 || !allow(g.Provider, r.URL) {
			return ErrNotAllowed
		}
		return nil
	}
	resp, err := hc.Do(req)
	if errors.Is(err, ErrNotAllowed) {
		return "", ErrNotAllowed
	}
	if err != nil {
		return "", ErrDownload
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", ErrDownload
	}
	head, ext, notMedia := make([]byte, 6), ".gif", ErrNotGIF
	if g.IsClip {
		head, ext, notMedia = make([]byte, 8), ".mp4", ErrNotMP4
	}
	if _, err := io.ReadFull(resp.Body, head); err != nil || !magic(g.IsClip, head) {
		return "", notMedia
	}

	sum := sha1.Sum([]byte(string(g.Provider) + ":" + g.ID))
	dir := filepath.Join(s.Dir, hex.EncodeToString(sum[:])[:16])
	final := filepath.Join(dir, slug(g.Title)+ext)
	if rel, err := filepath.Rel(s.Dir, final); err != nil || strings.HasPrefix(rel, "..") {
		return "", ErrNotAllowed
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", ErrDownload
	}
	part := final + ".part"
	f, err := os.Create(part)
	if err != nil {
		return "", ErrDownload
	}
	n, err := io.Copy(f, io.MultiReader(strings.NewReader(string(head)), io.LimitReader(resp.Body, limit+1-int64(len(head)))))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	switch {
	case err != nil:
		err = ErrDownload
	case n > limit:
		err = ErrTooLarge
	default:
		if os.Rename(part, final) == nil {
			return final, nil
		}
		err = ErrDownload
	}
	_ = os.Remove(part)
	_ = os.Remove(dir) // only succeeds when empty: an earlier copy of the same GIF stays
	return "", err
}

// magic reports whether head starts a GIF or, for a Clip, an MP4 (an ftyp box at offset 4).
func magic(clip bool, head []byte) bool {
	if clip {
		return string(head[4:8]) == "ftyp"
	}
	return string(head) == "GIF87a" || string(head) == "GIF89a"
}

// Allows applies the store's allowlist (nil Allow = VendorAllow).
func (s *Store) Allows(p provider.ID, u *url.URL) bool {
	if s.Allow == nil {
		return VendorAllow(p, u)
	}
	return s.Allow(p, u)
}

// Sweep removes every entry in Dir except the directory holding keep.
func (s *Store) Sweep(keep string) error {
	rel, err := filepath.Rel(s.Dir, keep)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ErrNotAllowed
	}
	keepDir := strings.SplitN(filepath.ToSlash(rel), "/", 2)[0]
	ents, err := os.ReadDir(s.Dir)
	if err != nil {
		return err
	}
	var errs []error
	for _, e := range ents {
		if e.Name() != keepDir {
			errs = append(errs, os.RemoveAll(filepath.Join(s.Dir, e.Name())))
		}
	}
	return errors.Join(errs...)
}

// Clear removes every entry in Dir. Skeleton (03-03 RED).
func (s *Store) Clear() error { return nil }

// Remove deletes one staged directory. Skeleton (03-03 RED).
func (s *Store) Remove(dir string) error { return nil }

// VendorAllow admits only https URLs on the provider's own media hosts (default port).
func VendorAllow(p provider.ID, u *url.URL) bool {
	if u.Scheme != "https" || u.Port() != "" {
		return false
	}
	h := strings.ToLower(u.Hostname())
	under := func(d string) bool { return h == d || strings.HasSuffix(h, "."+d) }
	switch p {
	case provider.GIPHY, provider.Clips:
		return under("giphy.com")
	case provider.Klipy:
		return under("klipy.com")
	case provider.FinerGifs:
		return h == "media.thefinergifs.club"
	}
	return false
}

// PageAllow admits only GIPHY and Clips page URLs: https, default port, giphy.com or a subdomain.
// No other provider has a page URL (a Klipy page is never constructed).
func PageAllow(p provider.ID, u *url.URL) bool {
	return (p == provider.GIPHY || p == provider.Clips) && VendorAllow(p, u) // same rule as GIPHY media hosts
}

// slug is the lowercase letters and digits of title with "-" for each run of anything else,
// capped at 80 bytes on a rune boundary; "gif" when nothing is left.
func slug(title string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			dash = false
			if b.Len()+utf8.RuneLen(r) > slugMax {
				break
			}
			b.WriteRune(r)
		} else {
			dash = true
		}
	}
	s := strings.TrimRight(b.String(), "-")
	if s == "" {
		return "gif"
	}
	return s
}
