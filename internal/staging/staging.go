// Package staging downloads the selected GIF (or a Clip's MP4) to the one file the pasteboard points at.
// There is no media cache: Sweep removes every older staged file after a successful copy.
// Save (user downloads) writes into a destination folder instead and never touches Dir.
package staging

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	ErrFolder     = errors.New("the download folder is not writable") // missing, read-only or denied
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
	return s.FetchMax(ctx, g, s.limit())
}

// FetchMax is Fetch with a byte cap of maxBytes: a larger download is ErrTooLarge and leaves nothing.
func (s *Store) FetchMax(ctx context.Context, g provider.Gif, maxBytes int64) (string, error) {
	body, ext, done, err := s.media(ctx, g, maxBytes)
	if err != nil {
		return "", err
	}
	defer done()
	return s.put(string(g.Provider)+":"+g.ID, slug(g.Title)+ext, func(f *os.File) error { return spool(f, body, maxBytes) })
}

// WriteDerived stages b, a file made from g, as <Dir>/<sha1(provider:id:variant)[:16]>/<slug>-<variant>.gif.
// Its own directory means Sweep(path) removes the original it was made from.
func (s *Store) WriteDerived(g provider.Gif, variant string, b []byte) (string, error) {
	return s.put(string(g.Provider)+":"+g.ID+":"+variant, slug(g.Title)+"-"+variant+".gif", func(f *os.File) error {
		if _, err := f.Write(b); err != nil {
			return ErrDownload
		}
		return nil
	})
}

// put writes name inside the hash directory of key through a .part file and a rename, so the
// pasteboard never points at a partial file.
func (s *Store) put(key, name string, fill func(*os.File) error) (string, error) {
	sum := sha1.Sum([]byte(key))
	dir := filepath.Join(s.Dir, hex.EncodeToString(sum[:])[:16])
	final := filepath.Join(dir, name)
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
	err = fill(f)
	if cerr := f.Close(); err == nil && cerr != nil {
		err = ErrDownload
	}
	if err == nil {
		if os.Rename(part, final) == nil {
			return final, nil
		}
		err = ErrDownload
	}
	_ = os.Remove(part)
	_ = os.Remove(dir) // only succeeds when empty: an earlier copy of the same GIF stays
	return "", err
}

// Save downloads g into dir as <FileStem><ext>, then "<stem> (2)<ext>", " (3)"… and returns the
// path. It never overwrites, never creates dir, and never leaves a partial file under a final name:
// the body goes to a hidden temp file that os.Link claims under the first free name.
func (s *Store) Save(ctx context.Context, g provider.Gif, dir string, hide bool) (string, error) {
	body, ext, done, err := s.media(ctx, g, s.limit()) // validated before anything exists in dir
	if err != nil {
		return "", err
	}
	defer done()
	f, err := os.CreateTemp(dir, ".gifsmith-*.part")
	if err != nil {
		return "", ErrFolder
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	err = spool(f, body, s.limit())
	if err == nil && (f.Sync() != nil || f.Chmod(0o644) != nil) { // CreateTemp makes 0600; a link shares it
		err = ErrFolder
	}
	if cerr := f.Close(); err == nil && cerr != nil {
		err = ErrFolder
	}
	if err != nil {
		return "", err
	}
	stem, clean := FileStem(g, hide), filepath.Clean(dir)
	for n := 1; n < 1000; n++ {
		name := stem + ext
		if n > 1 {
			name = fmt.Sprintf("%s (%d)%s", stem, n, ext)
		}
		candidate := filepath.Join(dir, name)
		if filepath.Dir(candidate) != clean {
			return "", ErrNotAllowed
		}
		err := os.Link(tmp, candidate)
		if err == nil {
			return candidate, nil
		}
		if !errors.Is(err, fs.ErrExist) { // a LinkError holds both paths: never wrapped
			return "", ErrFolder
		}
	}
	return "", ErrFolder
}

// FileStem is the download name without its extension: the title slug plus the id's
// [A-Za-z0-9_-] characters (at most 64 bytes); hide replaces the slug with "gifsmith".
func FileStem(g provider.Gif, hide bool) string {
	head := slug(g.Title)
	if hide {
		head = "gifsmith"
	}
	var id strings.Builder
	for _, r := range g.ID {
		if id.Len() == 64 {
			break
		}
		if r < utf8.RuneSelf && (r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)) {
			id.WriteRune(r)
		}
	}
	if id.Len() == 0 {
		return head
	}
	return head + "-" + id.String()
}

// media requests g.DownloadURL through the allowlist (the URL and every redirect hop), needs a 200
// and the GIF magic (a Clip: an MP4 ftyp box) before returning. body yields the whole file, capped
// one byte past maxBytes; done releases the response and the 45 s deadline.
func (s *Store) media(ctx context.Context, g provider.Gif, maxBytes int64) (body io.Reader, ext string, done func(), err error) {
	allow := s.Allows
	u, err := url.Parse(g.DownloadURL)
	if err != nil || !allow(g.Provider, u) {
		return nil, "", nil, ErrNotAllowed
	}
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		cancel()
		return nil, "", nil, ErrNotAllowed
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
	if err != nil {
		cancel()
		if errors.Is(err, ErrNotAllowed) {
			return nil, "", nil, ErrNotAllowed
		}
		return nil, "", nil, ErrDownload
	}
	done = func() { _ = resp.Body.Close(); cancel() }
	if resp.StatusCode != http.StatusOK {
		done()
		return nil, "", nil, ErrDownload
	}
	head, ext, notMedia := make([]byte, 6), ".gif", ErrNotGIF
	if g.IsClip {
		head, ext, notMedia = make([]byte, 8), ".mp4", ErrNotMP4
	}
	if _, err := io.ReadFull(resp.Body, head); err != nil || !magic(g.IsClip, head) {
		done()
		return nil, "", nil, notMedia
	}
	return io.MultiReader(bytes.NewReader(head), io.LimitReader(resp.Body, maxBytes+1-int64(len(head)))), ext, done, nil
}

// spool copies body into f: ErrDownload on a read or write failure, ErrTooLarge past maxBytes.
func spool(f *os.File, body io.Reader, maxBytes int64) error {
	n, err := io.Copy(f, body)
	switch {
	case err != nil:
		return ErrDownload
	case n > maxBytes:
		return ErrTooLarge
	}
	return nil
}

func (s *Store) limit() int64 {
	if s.maxBytes == 0 {
		return defaultMax
	}
	return s.maxBytes
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
	return s.removeExcept(strings.SplitN(filepath.ToSlash(rel), "/", 2)[0])
}

// Clear removes every entry in Dir (at launch, at quit, after a text copy); a missing Dir is fine.
func (s *Store) Clear() error {
	if err := s.removeExcept(""); !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// Remove deletes one staged entry of Dir; Dir itself and anything outside it are refused.
func (s *Store) Remove(dir string) error {
	rel, err := filepath.Rel(s.Dir, dir)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return ErrNotAllowed
	}
	return os.RemoveAll(filepath.Join(s.Dir, rel))
}

func (s *Store) removeExcept(keepDir string) error {
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
