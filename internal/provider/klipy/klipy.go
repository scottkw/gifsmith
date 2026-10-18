// Package klipy is the Klipy native v1 client (api.klipy.com/api/v1/{key}/gifs).
// Klipy's integration rules forbid reordering, removing or deduplicating results,
// so the mapper keeps the returned order exactly.
package klipy

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/httpx"
)

type Client struct {
	BaseURL string // https://api.klipy.com
	Key     func() (string, bool)
	HTTP    *http.Client
	// Params is read on every request (page size, country); nil = provider.DefaultParams().
	Params func() provider.Params
}

// Dimensions and size are display-only, so decoded loosely.
type size struct {
	GIF struct {
		URL    string         `json:"url"`
		Width  httpx.LooseInt `json:"width"`
		Height httpx.LooseInt `json:"height"`
		Size   httpx.LooseInt `json:"size"`
	} `json:"gif"`
}

const (
	maxTags     = 50
	maxTagRunes = 100
)

// tags keeps the non-empty string elements of Klipy's undocumented tags value; anything else is ignored.
func tags(raw json.RawMessage) []string {
	var elems []json.RawMessage
	if json.Unmarshal(raw, &elems) != nil {
		return nil
	}
	var out []string
	for _, e := range elems {
		var s string
		if json.Unmarshal(e, &s) != nil || s == "" {
			continue
		}
		if r := []rune(s); len(r) > maxTagRunes {
			s = string(r[:maxTagRunes])
		}
		if out = append(out, s); len(out) == maxTags {
			break
		}
	}
	return out
}

// The 16-digit numeric id is deliberately not decoded (float64 precision); slug is the id.
type response struct {
	Result bool `json:"result"`
	Data   *struct {
		Data []struct {
			Slug  string          `json:"slug"`
			Title string          `json:"title"`
			Type  string          `json:"type"`
			Tags  json.RawMessage `json:"tags"`
			File  struct {
				HD size `json:"hd"`
				MD size `json:"md"`
				SM size `json:"sm"`
			} `json:"file"`
		} `json:"data"`
		CurrentPage int  `json:"current_page"`
		HasNext     bool `json:"has_next"`
	} `json:"data"`
}

func bad(status int) error {
	return &provider.Error{Kind: provider.HTTP, Provider: provider.Klipy, Status: status}
}

// badKey reports a key that cannot safely be a URL path segment.
func badKey(k string) bool {
	return strings.ContainsAny(k, "/?#") || strings.ContainsFunc(k, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) })
}

func (c *Client) Search(ctx context.Context, q, cursor string) (provider.Page, error) {
	return c.get(ctx, "search", url.Values{"q": {q}}, cursor)
}

func (c *Client) Trending(ctx context.Context, cursor string) (provider.Page, error) {
	return c.get(ctx, "trending", url.Values{}, cursor)
}

// get pages with an opaque cursor that is the 1-based page number.
func (c *Client) get(ctx context.Context, kind string, v url.Values, cursor string) (provider.Page, error) {
	page := 1
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil || n < 1 {
			return provider.Page{}, bad(0) // never silently restart at the first page
		}
		page = n
	}
	key, ok := c.Key()
	if !ok || key == "" {
		return provider.Page{}, &provider.Error{Kind: provider.KeyMissing, Provider: provider.Klipy}
	}
	if badKey(key) {
		return provider.Page{}, &provider.Error{Kind: provider.KeyRejected, Provider: provider.Klipy}
	}
	v.Set("page", strconv.Itoa(page))
	prm := provider.ParamsOf(c.Params)
	v.Set("per_page", strconv.Itoa(min(max(prm.PageSize, 8), 50))) // Klipy accepts 8..50
	v.Set("content_filter", "medium")

	var r response
	u := c.BaseURL + "/api/v1/" + url.PathEscape(key) + "/gifs/" + kind + "?" + v.Encode()
	status, err := httpx.GetJSON(ctx, c.HTTP, provider.Klipy, u, &r)
	if err != nil {
		return provider.Page{}, httpx.Remap(err, 401, 403, 404) // Klipy reports an invalid key as 404
	}
	if !r.Result || r.Data == nil || r.Data.Data == nil {
		return provider.Page{}, bad(status)
	}
	items := make([]provider.Gif, 0, len(r.Data.Data))
	skipped := 0
	for _, d := range r.Data.Data {
		file := d.File.HD.GIF
		if file.URL == "" {
			file = d.File.MD.GIF
		}
		dl := file.URL
		// Klipy forbids filtering results, so any item with GIF renditions (ads included) is
		// kept in place. An item without them (e.g. an HTML ad) has nothing a GIF grid can show.
		if d.Type != "gif" && (d.File.SM.GIF.URL == "" || dl == "") {
			skipped++ // surfaced in the log rather than silently dropped
			continue
		}
		if d.File.SM.GIF.URL == "" || dl == "" {
			return provider.Page{}, bad(status)
		}
		title := d.Title
		if title == "" {
			title = d.Slug
		}
		// No PageURL: Klipy returns none, and one is never constructed.
		items = append(items, provider.Gif{Provider: provider.Klipy, ID: provider.ClampID(d.Slug), Title: provider.ClampTitle(title), PreviewURL: d.File.SM.GIF.URL, DownloadURL: dl,
			Width: int(file.Width), Height: int(file.Height), Size: int64(file.Size), Tags: tags(d.Tags), MediumURL: d.File.MD.GIF.URL})
	}
	if skipped > 0 {
		log.Printf("klipy: skipped %d non-gif items", skipped)
	}
	p := provider.Page{Items: items}
	if r.Data.HasNext && len(items) > 0 { // an empty page with has_next would make the grid fetch again at once
		cur := r.Data.CurrentPage
		if cur < 1 {
			cur = page
		}
		p.Next = strconv.Itoa(cur + 1)
	}
	return p, nil
}
