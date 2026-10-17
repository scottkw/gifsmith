// Package staging downloads the selected GIF to the one file the pasteboard points at.
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

// Fetch downloads g.DownloadURL to <Dir>/<sha1(provider:id)[:16]>/<slug>.gif and returns that path.
// Transport errors are never wrapped: their text holds the URL.
func (s *Store) Fetch(ctx context.Context, g provider.Gif) (string, error) {
	allow := s.Allow
	if allow == nil {
		allow = VendorAllow
	}
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
	head := make([]byte, 6)
	if _, err := io.ReadFull(resp.Body, head); err != nil || (string(head) != "GIF87a" && string(head) != "GIF89a") {
		return "", ErrNotGIF
	}

	sum := sha1.Sum([]byte(string(g.Provider) + ":" + g.ID))
	dir := filepath.Join(s.Dir, hex.EncodeToString(sum[:])[:16])
	final := filepath.Join(dir, slug(g.Title)+".gif")
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
