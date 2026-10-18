// Package provider holds the contract every GIF source implements and the one
// error type that crosses the Wails boundary. Domain code only: no Wails imports.
package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

type ID string

const (
	GIPHY     ID = "giphy"
	Clips     ID = "clips"
	Klipy     ID = "klipy"
	FinerGifs ID = "finergifs"
)

// ParseID accepts exactly the four provider IDs.
func ParseID(s string) (ID, bool) {
	switch id := ID(s); id {
	case GIPHY, Clips, Klipy, FinerGifs:
		return id, true
	}
	return "", false
}

// Library limits (library.Validate): the mappers clamp vendor text to them, so a long
// user-supplied title or id never makes an item unsavable (WR-02).
const (
	MaxTitleRunes = 500
	MaxIDBytes    = 200
)

// ClampTitle shortens s to MaxTitleRunes runes.
func ClampTitle(s string) string {
	if r := []rune(s); len(r) > MaxTitleRunes {
		return string(r[:MaxTitleRunes])
	}
	return s
}

// ClampID shortens s to at most MaxIDBytes bytes, never splitting a rune.
func ClampID(s string) string {
	if len(s) <= MaxIDBytes {
		return s
	}
	i := MaxIDBytes
	for i > 0 && !utf8.RuneStart(s[i]) {
		i--
	}
	return s[:i]
}

// DefaultPageSize is the page size when the user has not chosen one.
const DefaultPageSize = 24

// Params are the user's per-request provider settings (Settings → Search). Clients read them
// on every request, so a change applies to the next call.
type Params struct {
	PageSize int    // one of PageSizes; clients clamp to vendor limits
	Rating   string // GIPHY and Clips only; one of Ratings
	Lang     string // GIPHY language code, search only
	Country  string // Klipy locale, lowercase ISO 3166-1 alpha-2
}

// PageSizes and Ratings are the allowed choices, in display order.
var (
	PageSizes = []int{24, 36, 48}
	Ratings   = []string{"g", "pg", "pg-13", "r"}
)

// DefaultParams are the values a fresh install sends.
func DefaultParams() Params {
	return Params{PageSize: DefaultPageSize, Rating: "pg-13", Lang: "en", Country: "us"}
}

// ParamsOf is f's current Params, or DefaultParams when a client has no Params func (tests).
func ParamsOf(f func() Params) Params {
	if f == nil {
		return DefaultParams()
	}
	return f()
}

type Gif struct {
	Provider    ID     `json:"provider"`
	ID          string `json:"id"` // GIPHY id, Klipy slug, Finer fileid
	Title       string `json:"title"`
	PreviewURL  string `json:"previewUrl"`  // grid only
	DownloadURL string `json:"downloadUrl"` // copy only
	IsClip      bool   `json:"isClip"`
	// Optional metadata, filled only from fields the vendor response carries (absent = unknown).
	// Width/Height/Size describe the file Copy puts on the clipboard.
	Width  int      `json:"width,omitempty"`
	Height int      `json:"height,omitempty"`
	Size   int64    `json:"size,omitempty"` // bytes
	Tags   []string `json:"tags,omitempty"`
	// PageURL is the vendor's https page for the item; "" = no page actions (Klipy, Finer Gifs).
	PageURL   string `json:"pageUrl,omitempty"`
	SourceURL string `json:"sourceUrl,omitempty"` // https only; can point anywhere
	MediumURL string `json:"mediumUrl,omitempty"` // Details preview; "" = use PreviewURL
	// GifURL is the GIF rendition used for link and Markdown when DownloadURL is not a GIF (Clips).
	GifURL string `json:"gifUrl,omitempty"`
}

type Page struct {
	Items []Gif  `json:"items"` // never nil: JS gets [] rather than null
	Next  string `json:"next"`  // opaque cursor; "" = end of results
}

type Provider interface {
	Search(ctx context.Context, q, cursor string) (Page, error)
	Trending(ctx context.Context, cursor string) (Page, error)
}

type Kind string

const (
	KeyMissing  Kind = "key_missing"
	KeyRejected Kind = "key_rejected"
	RateLimited Kind = "rate_limited"
	Network     Kind = "network"
	Timeout     Kind = "timeout"
	HTTP        Kind = "http" // non-2xx other than the above, or an empty/invalid body (Status = HTTP status)
	// ClipsDisabled: Go refuses Clips calls until the user's GIPHY key passed the Clips probe (no request made).
	ClipsDisabled Kind = "clips_disabled"
)

type Error struct {
	Kind          Kind `json:"kind"`
	Provider      ID   `json:"provider"`
	Status        int  `json:"status,omitempty"`
	RetryAfterSec int  `json:"retryAfterSec,omitempty"` // 0 = unknown
}

// Error never includes a URL, query, key or response body: Wails sends this string to JS verbatim.
func (e *Error) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("%s: %s (HTTP %d)", e.Provider, e.Kind, e.Status)
	}
	return fmt.Sprintf("%s: %s", e.Provider, e.Kind)
}

// MarshalError is Wails' Options.MarshalError: a *Error becomes err.cause in JS;
// nil falls back to Wails' default for every other error.
func MarshalError(err error) []byte {
	var pe *Error
	if !errors.As(err, &pe) {
		return nil
	}
	b, _ := json.Marshal(pe) // cannot fail for this struct
	return b
}
