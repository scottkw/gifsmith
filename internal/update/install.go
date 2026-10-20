package update

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// NewDownloadClient follows at most 5 redirects, each of which allow must accept. No overall
// timeout: the caller's context bounds a download.
func NewDownloadClient(allow func(*url.URL) bool) *http.Client {
	return &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		if !allow(req.URL) {
			return errors.New("redirect not allowed")
		}
		return nil
	}}
}

// Download fetches a into dir/a.Name (0600, dir 0700) through a .part file. The byte count must
// equal a.Size and the SHA-256 a.SHA256 (else ErrDigest); on any error nothing is left behind.
// progress (may be nil) is called at most every 100 ms and once at the end.
func Download(ctx context.Context, client *http.Client, a Asset, dir string, progress func(done, total int64)) (path string, err error) {
	if a.Size > MaxDMGBytes {
		return "", ErrTooLarge
	}
	if a.Name == "" || filepath.Base(a.Name) != a.Name {
		return "", fmt.Errorf("%w: bad asset name", ErrNotInstallable)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	final := filepath.Join(dir, a.Name)
	part := final + ".part"
	_ = os.Remove(part)
	f, err := os.OpenFile(part, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(part)
			_ = os.Remove(final)
		}
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Gifsmith")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: status %d", resp.StatusCode)
	}
	h := sha256.New()
	pw := &progressWriter{total: a.Size, fn: progress}
	n, err := io.Copy(io.MultiWriter(f, h, pw), io.LimitReader(resp.Body, a.Size+1))
	if err != nil {
		return "", err
	}
	if n != a.Size {
		return "", fmt.Errorf("download: got %d bytes, want %d", n, a.Size)
	}
	if hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
		return "", ErrDigest
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = os.Rename(part, final); err != nil {
		return "", err
	}
	if progress != nil {
		progress(a.Size, a.Size)
	}
	return final, nil
}

type progressWriter struct {
	done, total int64
	last        time.Time
	fn          func(done, total int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	p.done += int64(len(b))
	if p.fn != nil && time.Since(p.last) >= 100*time.Millisecond {
		p.last = time.Now()
		p.fn(min(p.done, p.total), p.total)
	}
	return len(b), nil
}

// Runner runs a tool by absolute path and returns its combined output.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// ExecRunner runs the real tool (no shell).
func ExecRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func hasLine(out []byte, line string) bool {
	return slices.Contains(strings.Split(string(bytes.ReplaceAll(out, []byte("\r"), nil)), "\n"), line)
}

// Verify checks the downloaded DMG's signature, Team ID and notarization, then quarantines it so
// it behaves like a browser download. Every failure wraps ErrVerify with the step name.
func Verify(ctx context.Context, run Runner, path, teamID string, now time.Time) error {
	if _, err := run(ctx, "/usr/bin/codesign", "--verify", "--strict", path); err != nil {
		return fmt.Errorf("%w: codesign", ErrVerify)
	}
	out, err := run(ctx, "/usr/bin/codesign", "-dvv", path)
	if err != nil || !hasLine(out, "TeamIdentifier="+teamID) {
		return fmt.Errorf("%w: team", ErrVerify)
	}
	out, err = run(ctx, "/usr/sbin/spctl", "-a", "-t", "open", "--context", "context:primary-signature", "-vv", path)
	if err != nil || !bytes.Contains(out, []byte(": accepted")) || !hasLine(out, "source=Notarized Developer ID") {
		return fmt.Errorf("%w: spctl", ErrVerify)
	}
	flag := fmt.Sprintf("0081;%x;Gifsmith;", now.Unix())
	if _, err := run(ctx, "/usr/bin/xattr", "-w", "com.apple.quarantine", flag, path); err != nil {
		return fmt.Errorf("%w: quarantine", ErrVerify)
	}
	return nil
}

// Clean removes Gifsmith's own update files (Gifsmith-*-arm64.dmg and .part) from dir; other
// files stay. A missing dir is fine.
func Clean(dir string) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		n := e.Name()
		dmg, _ := filepath.Match("Gifsmith-*-arm64.dmg", n)
		part, _ := filepath.Match("Gifsmith-*-arm64.dmg.part", n)
		if (dmg || part) && e.Type().IsRegular() {
			if err := os.Remove(filepath.Join(dir, n)); err != nil {
				return err
			}
		}
	}
	return nil
}
