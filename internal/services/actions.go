// Package services holds the methods bound to the webview. They take no paths from
// JS, and the only URL they fetch is a Gif's DownloadURL that passes the staging allowlist.
package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image/gif"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gifsmith/internal/imaging"
	"gifsmith/internal/platform"
	"gifsmith/internal/provider"
	"gifsmith/internal/staging"
)

// PasteDelay is the wait between hiding the panel (or reactivating the previous app) and ⌘V.
const PasteDelay = 80 * time.Millisecond

// KlipyPasteGrace is how long a pasted Klipy file stays for the target app to read it (D-19).
const KlipyPasteGrace = 30 * time.Second

type Result struct {
	Copied             bool   `json:"copied"`
	Pasted             bool   `json:"pasted"`
	NeedsAccessibility bool   `json:"needsAccessibility"`
	Message            string `json:"message"`
}

type ActionService struct {
	Platform platform.Platform
	Staging  *staging.Store
	Hide     func()
	Previous func() platform.App
	Sleep    func(time.Duration)         // nil means time.Sleep
	After    func(time.Duration, func()) // nil means time.AfterFunc
	// DownloadsDir is the download folder, read at each Download (SettingsService.DownloadDir).
	DownloadsDir func() string
	// HideFilenames names downloads gifsmith-{id} instead of the title slug; nil means off.
	HideFilenames func() bool
	// Recents records every successful copy or paste (D-07); nil records nothing.
	Recents      interface{ Touch(provider.Gif) error }
	busy         sync.Mutex    // one action at a time: no racing pasteboard writes, no double ⌘V
	gen          atomic.Uint64 // bumped on every clipboard write; a Klipy timer deletes only if unchanged
	lastDownload string        // the file Show in Finder reveals; guarded by busy
}

// stage downloads g, puts its file URL on the pasteboard and returns the staged path. Older
// staged files are swept only after that succeeds: until then the clipboard may point at one.
func (a *ActionService) stage(ctx context.Context, g provider.Gif) (string, error) {
	path, err := a.Staging.Fetch(ctx, g)
	if err != nil {
		log.Printf("stage %s: %v", g.Provider, err)
		return "", stageErr(err)
	}
	if err := a.Platform.CopyFile(path); err != nil {
		log.Printf("stage %s: pasteboard: %v", g.Provider, err)
		return "", actionErr("copy_failed")
	}
	a.gen.Add(1)
	if err := a.Staging.Sweep(path); err != nil {
		log.Printf("staging sweep: %v", err)
	}
	return path, nil
}

// stageErr maps the staging sentinels to action kinds (they never carry a URL either).
func stageErr(err error) error {
	switch {
	case errors.Is(err, staging.ErrNotAllowed):
		return actionErr("not_allowed")
	case errors.Is(err, staging.ErrNotGIF), errors.Is(err, staging.ErrNotMP4):
		return actionErr("not_media")
	case errors.Is(err, staging.ErrTooLarge):
		return actionErr("too_large")
	case errors.Is(err, staging.ErrFolder):
		return actionErr("folder_unwritable")
	}
	return actionErr("download_failed")
}

// Copy puts the selected GIF (a Clip's MP4) on the clipboard as a file. The panel stays open:
// the frontend shows the confirmation, then hides it.
func (a *ActionService) Copy(ctx context.Context, g provider.Gif) (Result, error) {
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
	if _, err := a.stage(ctx, g); err != nil {
		return Result{}, err
	}
	log.Printf("copy: staged and copied")
	a.recent(g, "copy")
	if g.IsClip {
		return Result{Copied: true, Message: "Copied MP4 — press ⌘V to paste"}, nil
	}
	return Result{Copied: true, Message: "Copied — press ⌘V to paste"}, nil
}

// Paste copies the GIF, hides the panel and posts one ⌘V to the previous app.
// Without Accessibility it only copies and keeps the panel open with instructions.
func (a *ActionService) Paste(ctx context.Context, g provider.Gif) (Result, error) {
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
	msg := "Pasted GIF"
	if g.IsClip {
		msg = "Pasted MP4"
	}
	var path string
	r, err := a.paste(func() (err error) { path, err = a.stage(ctx, g); return err }, msg)
	a.recentIf(r, g, "paste")
	if r.Pasted && g.Provider == provider.Klipy {
		a.expire(path)
	}
	return r, err
}

// expire deletes a pasted Klipy file KlipyPasteGrace after ⌘V (D-19: Klipy media is not retained),
// unless the clipboard was written since: a newer copy of the same GIF reuses the same directory.
// The caller holds busy, so gen is this paste's generation.
func (a *ActionService) expire(path string) {
	dir, gen := filepath.Dir(path), a.gen.Load()
	after := a.After
	if after == nil {
		after = func(d time.Duration, f func()) { time.AfterFunc(d, f) }
	}
	after(KlipyPasteGrace, func() {
		// busy keeps a copy from re-staging dir between the check and the removal.
		// shortcut: a keypress landing in the microseconds this holds busy is refused as busy; fine at this rate.
		a.busy.Lock()
		defer a.busy.Unlock()
		if a.gen.Load() != gen {
			return
		}
		if err := a.Staging.Remove(dir); err != nil {
			log.Printf("klipy grace delete: %v", err)
		}
	})
}

// paste runs put (which fills the clipboard), then hides the panel and posts one ⌘V to the
// previous app; okMsg is the result message when it did. The caller holds busy.
func (a *ActionService) paste(put func() error, okMsg string) (Result, error) {
	if err := put(); err != nil {
		return Result{}, err
	}
	if !a.Platform.AccessibilityTrusted(false) {
		log.Printf("paste: accessibility not trusted, copied only (UI shows Enable Automatic Paste…)")
		return Result{Copied: true, NeedsAccessibility: true, Message: "Copied — press ⌘V to paste"}, nil // T5a: the button's title explains
	}
	sleep := a.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	a.Hide()
	sleep(PasteDelay)
	// Never post ⌘V into an app the user did not pick: an unknown target or a failed lookup copies only.
	prev := a.Previous()
	if prev.PID == 0 {
		log.Printf("paste: no remembered target, copied only")
		return Result{Copied: true, Message: "Copied — press ⌘V to paste"}, nil
	}
	front, err := a.Platform.FrontmostApp()
	if err != nil {
		log.Printf("paste: frontmost app: %v, copied only", err)
		return Result{Copied: true, Message: "Copied — press ⌘V to paste"}, nil
	}
	if front.PID != prev.PID {
		if err := a.Platform.ActivateApp(prev.PID); err != nil {
			log.Printf("reactivate %s: %v", prev.Name, err)
		}
		sleep(PasteDelay)
		// macOS 14+ may decline the activation.
		if now, err := a.Platform.FrontmostApp(); err != nil || now.PID != prev.PID {
			log.Printf("paste: could not return to %q, copied only", prev.Name)
			return Result{Copied: true, Message: "Copied — press ⌘V to paste (could not return to " + prev.Name + ")"}, nil
		}
	}
	if err := a.Platform.PasteShortcut(); err != nil {
		// The panel is already hidden, so an error here would never be seen; report it as a result.
		log.Printf("paste: post ⌘V: %v", err)
		return Result{Copied: true, Message: "Copied — automatic paste failed, press ⌘V"}, nil
	}
	log.Printf("paste: posted ⌘V (previous app %q)", prev.Name)
	return Result{Copied: true, Pasted: true, Message: okMsg}, nil
}

var (
	copiedText = map[string]string{"link": "Copied link", "markdown": "Copied Markdown", "page": "Copied page link"}
	pastedText = map[string]string{"link": "Pasted link", "markdown": "Pasted Markdown"}
)

// CopyText puts text Go builds from g on the clipboard as a plain string. kind: link | markdown | page.
// The webview never composes the text: it only names the kind.
func (a *ActionService) CopyText(ctx context.Context, g provider.Gif, kind string) (Result, error) {
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
	if err := a.putText(g, kind); err != nil {
		return Result{}, err
	}
	if kind != "page" { // a page link is not the GIF itself (D-07)
		a.recent(g, kind)
	}
	return Result{Copied: true, Message: copiedText[kind]}, nil
}

// PasteText copies text like CopyText, then pastes it like Paste. kind: link | markdown.
func (a *ActionService) PasteText(ctx context.Context, g provider.Gif, kind string) (Result, error) {
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
	msg, ok := pastedText[kind]
	if !ok {
		return Result{}, actionErr("not_allowed")
	}
	r, err := a.paste(func() error { return a.putText(g, kind) }, msg)
	a.recentIf(r, g, "paste "+kind)
	return r, err
}

// recent records g in Recents after a successful copy or paste (D-07). A failure is logged,
// never reported: the clipboard already holds what the user asked for.
func (a *ActionService) recent(g provider.Gif, kind string) {
	if a.Recents == nil {
		return
	}
	if err := a.Recents.Touch(g); err != nil {
		log.Printf("recents: %s: %v", kind, err)
	}
}

// recentIf records g when r says the clipboard was filled (a paste that fell back to copying counts).
func (a *ActionService) recentIf(r Result, g provider.Gif, kind string) {
	if r.Copied {
		a.recent(g, kind)
	}
}

func (a *ActionService) putText(g provider.Gif, kind string) error {
	text, err := a.textFor(g, kind)
	if err != nil {
		return err
	}
	if err := a.Platform.CopyText(text); err != nil {
		log.Printf("copy text %s %s: %v", g.Provider, kind, err)
		return actionErr("copy_failed")
	}
	a.gen.Add(1)
	// The clipboard no longer points at a staged file, so none is kept (D-17).
	if err := a.Staging.Clear(); err != nil {
		log.Printf("staging clear: %v", err)
	}
	log.Printf("copy text: %s %s", g.Provider, kind)
	return nil
}

func (a *ActionService) textFor(g provider.Gif, kind string) (string, error) {
	switch kind {
	case "link":
		return a.linkFor(g)
	case "markdown":
		link, err := a.linkFor(g)
		if err != nil {
			return "", err
		}
		return markdownFor(g.Title, link), nil
	case "page":
		return pageFor(g)
	}
	return "", actionErr("not_allowed")
}

// linkFor is the GIF's media URL (a Clip's GIF rendition, never its MP4: D-18), checked against
// the same allowlist as downloads. GIPHY links lose their tracking query; Klipy forbids altering
// URLs, so other providers' links pass byte for byte.
func (a *ActionService) linkFor(g provider.Gif) (string, error) {
	raw := g.GifURL
	if raw == "" {
		if g.IsClip || g.Provider == provider.Clips {
			return "", actionErr("no_link")
		}
		raw = g.DownloadURL
	}
	u, err := url.Parse(raw)
	if err != nil || !a.Staging.Allows(g.Provider, u) {
		return "", actionErr("not_allowed")
	}
	if g.Provider != provider.GIPHY && g.Provider != provider.Clips {
		return raw, nil
	}
	u.RawQuery, u.ForceQuery, u.Fragment, u.RawFragment = "", false, "", ""
	return u.String(), nil
}

// markdownFor is ![alt](link): whitespace runs collapse, an empty title becomes "GIF", and
// backslash and brackets are escaped so the title cannot end the alt text early.
func markdownFor(title, link string) string {
	alt := strings.Join(strings.Fields(title), " ")
	if alt == "" {
		alt = "GIF"
	}
	alt = strings.NewReplacer(`\`, `\\`, `[`, `\[`, `]`, `\]`).Replace(alt)
	return "![" + alt + "](" + link + ")"
}

// pageFor is the item's vendor page: GIPHY and Clips only (staging.PageAllow), never constructed.
func pageFor(g provider.Gif) (string, error) {
	u, err := url.Parse(g.PageURL)
	if err != nil || !staging.PageAllow(g.Provider, u) {
		return "", actionErr("no_page")
	}
	return u.String(), nil
}

// OpenPage opens the item's GIPHY page in the default browser (platform.Fake in the UAT build).
func (a *ActionService) OpenPage(ctx context.Context, g provider.Gif) error {
	if !a.busy.TryLock() {
		return ErrBusy
	}
	defer a.busy.Unlock()
	page, err := pageFor(g)
	if err != nil {
		return err
	}
	return a.open(g, "page", page)
}

// OpenSource opens the item's third-party source page; https only, it can point anywhere.
func (a *ActionService) OpenSource(ctx context.Context, g provider.Gif) error {
	if !a.busy.TryLock() {
		return ErrBusy
	}
	defer a.busy.Unlock()
	u, err := url.Parse(g.SourceURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return actionErr("no_page")
	}
	return a.open(g, "source", u.String())
}

func (a *ActionService) open(g provider.Gif, kind, u string) error {
	if err := a.Platform.OpenURL(u); err != nil {
		log.Printf("open %s %s: %v", g.Provider, kind, err)
		return actionErr("open_failed")
	}
	log.Printf("open: %s %s", g.Provider, kind)
	return nil
}

// RequestAccessibility raises the macOS prompt and, if still untrusted, opens the settings pane.
// Only an explicit button click calls this.
func (a *ActionService) RequestAccessibility(ctx context.Context) (bool, error) {
	if !a.busy.TryLock() {
		return false, ErrBusy
	}
	defer a.busy.Unlock()
	if a.Platform.AccessibilityTrusted(true) {
		log.Printf("request accessibility: already trusted")
		return true, nil
	}
	log.Printf("request accessibility: prompted, opening the Accessibility pane")
	return false, a.Platform.OpenAccessibilitySettings()
}

// ServiceShutdown is the Wails service hook run on Quit, ⌘Q and SIGINT/SIGTERM: staging holds only
// the clipboard's file, and it does not outlive the app (D-17). It runs on the main thread inside
// Wails cleanup, so no busy lock and no onMain: an action still running must not block quitting.
func (a *ActionService) ServiceShutdown() error {
	return a.Staging.Clear()
}

// DownloadResult names the saved file and its folder for the toast; never a path.
type DownloadResult struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

// Download saves the selected GIF (a Clip's MP4) into the download folder; the panel stays open.
// It never goes through staging, the pasteboard or Recents (D-17, D-07).
func (a *ActionService) Download(ctx context.Context, g provider.Gif) (DownloadResult, error) {
	if !a.busy.TryLock() {
		return DownloadResult{}, ErrBusy
	}
	defer a.busy.Unlock()
	dir := ""
	if a.DownloadsDir != nil {
		dir = a.DownloadsDir()
	}
	if dir == "" {
		return DownloadResult{}, actionErr("folder_unwritable")
	}
	path, err := a.Staging.Save(ctx, g, dir, a.HideFilenames != nil && a.HideFilenames())
	if err != nil {
		log.Printf("download %s: %v", g.Provider, err)
		return DownloadResult{}, stageErr(err)
	}
	a.lastDownload = path
	log.Printf("download %s: saved", g.Provider)
	return DownloadResult{Name: filepath.Base(path), Folder: filepath.Base(dir)}, nil
}

// ShowDownload reveals the last downloaded file in Finder (platform.Fake in the UAT build).
// It takes no path from JS: only a file Download saved itself can be shown.
func (a *ActionService) ShowDownload(ctx context.Context) error {
	if !a.busy.TryLock() {
		return ErrBusy
	}
	defer a.busy.Unlock()
	if a.lastDownload == "" {
		return actionErr("nothing_to_show")
	}
	if err := a.Platform.RevealFile(a.lastDownload); err != nil {
		log.Printf("show download: %v", err)
		return actionErr("open_failed")
	}
	log.Printf("show download: revealed")
	return nil
}

// Square crop limits (D-16, T-03-18): DecodeAll holds every frame in memory. The byte and side
// limits bound the input; SquareMaxPixels bounds the decoded frames (CR-01: a 15 MB file can
// declare a thousand 4096x4096 frames).
const (
	SquareMaxBytes  = 15 << 20
	SquareMaxSide   = 4096
	SquareMaxPixels = 256 << 20  // summed w*h + frameOverhead over all frames, about 256 MB of *image.Paletted
	frameOverhead   = 4096 + 256 // Go's fixed per-frame cost (palette clone, image header), so empty frames count
)

// CopySquare puts a center-cropped square copy of the GIF on the clipboard as a file (D-16).
// The panel stays open: the frontend shows the confirmation, then hides it.
func (a *ActionService) CopySquare(ctx context.Context, g provider.Gif) (Result, error) {
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
	if _, err := a.stageSquare(ctx, g); err != nil {
		return Result{}, err
	}
	log.Printf("copy square: staged and copied")
	a.recent(g, "square")
	return Result{Copied: true, Message: "Copied square GIF — press ⌘V to paste"}, nil
}

// stageSquare is stage for the square crop: download (capped), crop in memory, stage the square
// file, put it on the pasteboard, then sweep everything else, including the downloaded original.
func (a *ActionService) stageSquare(ctx context.Context, g provider.Gif) (string, error) {
	if g.IsClip { // a Clip is an MP4 (D-15)
		return "", actionErr("not_allowed")
	}
	if g.Size > SquareMaxBytes { // refused before any request; FetchMax caps the real size anyway
		return "", actionErr("crop_too_large")
	}
	orig, err := a.Staging.FetchMax(ctx, g, SquareMaxBytes)
	if err != nil {
		log.Printf("square %s: %v", g.Provider, err)
		if errors.Is(err, staging.ErrTooLarge) {
			return "", actionErr("crop_too_large")
		}
		return "", stageErr(err)
	}
	b, err := cropSquare(orig)
	if err != nil {
		log.Printf("square %s: crop: %v", g.Provider, err)
		return "", actionErr("crop_failed")
	}
	path, err := a.Staging.WriteDerived(g, "square", b)
	if err != nil {
		log.Printf("square %s: %v", g.Provider, err)
		return "", stageErr(err)
	}
	if err := a.Platform.CopyFile(path); err != nil {
		log.Printf("square %s: pasteboard: %v", g.Provider, err)
		return "", actionErr("copy_failed")
	}
	a.gen.Add(1)
	if err := a.Staging.Sweep(path); err != nil {
		log.Printf("staging sweep: %v", err)
	}
	return path, nil
}

// cropSquare decodes the staged GIF at path and returns its square crop, encoded. The canvas size
// is checked before DecodeAll, which holds every frame in memory (Pitfall 2).
func cropSquare(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }() // read-only
	cfg, err := gif.DecodeConfig(f)
	if err != nil {
		return nil, err
	}
	if cfg.Width > SquareMaxSide || cfg.Height > SquareMaxSide {
		return nil, fmt.Errorf("canvas %dx%d over %d", cfg.Width, cfg.Height, SquareMaxSide)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	px, err := framePixels(f)
	if err != nil {
		return nil, err
	}
	if px > SquareMaxPixels {
		return nil, fmt.Errorf("frames total %d pixels, over %d", px, SquareMaxPixels)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, imaging.Square(g)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// framePixels sums the declared w*h of every frame by walking the GIF block structure, skipping
// color tables and data sub-blocks without decoding any LZW data (CR-01).
func framePixels(r io.Reader) (int64, error) {
	br := bufio.NewReader(r)
	var hdr [13]byte // signature, logical screen descriptor
	if _, err := io.ReadFull(br, hdr[:]); err != nil {
		return 0, err
	}
	if hdr[10]&0x80 != 0 { // global color table
		if _, err := br.Discard(3 << ((hdr[10] & 7) + 1)); err != nil {
			return 0, err
		}
	}
	var px int64
	for {
		b, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		switch b {
		case 0x21: // extension: walked exactly as image/gif's readExtension, so the two agree on its end
			label, err := br.ReadByte()
			if err != nil {
				return 0, err
			}
			switch label {
			case 0x01: // Plain Text: 13 raw bytes, then sub-blocks
				if _, err := br.Discard(13); err != nil {
					return 0, err
				}
			case 0xF9: // Graphic Control: exactly 6 bytes, no sub-blocks
				var gc [6]byte
				if _, err := io.ReadFull(br, gc[:]); err != nil {
					return 0, err
				}
				if gc[0] != 4 || gc[5] != 0 {
					return 0, fmt.Errorf("gif: bad graphic control extension")
				}
				continue
			case 0xFE: // Comment: sub-blocks only
			case 0xFF: // Application: a size byte and that many bytes (even 0), then sub-blocks
				n, err := br.ReadByte()
				if err != nil {
					return 0, err
				}
				if _, err := br.Discard(int(n)); err != nil {
					return 0, err
				}
			default:
				return 0, fmt.Errorf("gif: unknown extension 0x%02x", label)
			}
		case 0x2C: // image descriptor: x, y, w, h (uint16 LE), flags; then a local table, LZW code size, sub-blocks
			var d [9]byte
			if _, err := io.ReadFull(br, d[:]); err != nil {
				return 0, err
			}
			px += int64(binary.LittleEndian.Uint16(d[4:]))*int64(binary.LittleEndian.Uint16(d[6:])) + frameOverhead
			skip := 1
			if d[8]&0x80 != 0 {
				skip += 3 << ((d[8] & 7) + 1)
			}
			if _, err := br.Discard(skip); err != nil {
				return 0, err
			}
		case 0x3B: // trailer
			return px, nil
		default:
			return 0, fmt.Errorf("gif: unknown block type 0x%02x", b)
		}
		for { // data sub-blocks, ended by a zero length
			n, err := br.ReadByte()
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break
			}
			if _, err := br.Discard(int(n)); err != nil {
				return 0, err
			}
		}
	}
}

// PasteSquare copies the square GIF like CopySquare, then pastes it like Paste. A pasted Klipy
// square file gets the same grace delete as a pasted Klipy GIF (D-19).
func (a *ActionService) PasteSquare(ctx context.Context, g provider.Gif) (Result, error) {
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
	var path string
	r, err := a.paste(func() (err error) { path, err = a.stageSquare(ctx, g); return err }, "Pasted square GIF")
	a.recentIf(r, g, "paste square")
	if r.Pasted && g.Provider == provider.Klipy {
		a.expire(path)
	}
	return r, err
}
