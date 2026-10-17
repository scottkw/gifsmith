// Package giphy is the GIPHY GIFs client (api.giphy.com/v1/gifs).
package giphy

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/httpx"
)

const (
	searchOffsetCap   = 4999 // GIPHY rejects larger offsets; reaching a cap ends paging
	trendingOffsetCap = 499
	maxQueryRunes     = 50 // longer q gets HTTP 414
)

type Client struct {
	BaseURL string // https://api.giphy.com
	Key     func() (string, bool)
	HTTP    *http.Client
}

// flexInt accepts "200", 200, "" and null: GIPHY sends some numerics as strings.
type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = flexInt(n)
	return nil
}

// rendition decodes only the URL; width/height/size (strings on the wire) wait for Phase 3 details.
type rendition struct {
	URL string `json:"url"`
}

type response struct {
	Data []struct {
		ID     string `json:"id"`
		Slug   string `json:"slug"`
		Title  string `json:"title"`
		Images struct {
			FixedWidth rendition `json:"fixed_width"`
			Original   rendition `json:"original"`
		} `json:"images"`
	} `json:"data"`
	Pagination struct {
		TotalCount flexInt `json:"total_count"` // not returned on every endpoint
	} `json:"pagination"`
}

func bad(status int) error {
	return &provider.Error{Kind: provider.HTTP, Provider: provider.GIPHY, Status: status}
}

func (c *Client) Search(ctx context.Context, q, cursor string) (provider.Page, error) {
	if r := []rune(q); len(r) > maxQueryRunes {
		q = string(r[:maxQueryRunes])
	}
	return c.get(ctx, "/v1/gifs/search", url.Values{"q": {q}}, cursor, searchOffsetCap)
}

func (c *Client) Trending(ctx context.Context, cursor string) (provider.Page, error) {
	return c.get(ctx, "/v1/gifs/trending", url.Values{}, cursor, trendingOffsetCap)
}

// get pages with an opaque cursor that is the decimal offset.
func (c *Client) get(ctx context.Context, path string, v url.Values, cursor string, offsetCap int) (provider.Page, error) {
	offset := 0
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil || n < 0 {
			return provider.Page{}, bad(0) // never silently restart at the first page
		}
		offset = n
	}
	key, ok := c.Key()
	if !ok || key == "" {
		return provider.Page{}, &provider.Error{Kind: provider.KeyMissing, Provider: provider.GIPHY}
	}
	v.Set("api_key", key)
	v.Set("limit", strconv.Itoa(provider.PageSize))
	v.Set("offset", strconv.Itoa(offset))
	v.Set("rating", "pg-13")
	v.Set("bundle", "messaging_non_clips")

	var r response
	status, err := httpx.GetJSON(ctx, c.HTTP, provider.GIPHY, c.BaseURL+path+"?"+v.Encode(), &r)
	if err != nil {
		return provider.Page{}, httpx.Remap(err, 401, 403)
	}
	if r.Data == nil {
		return provider.Page{}, bad(status)
	}
	items := make([]provider.Gif, 0, len(r.Data))
	for _, d := range r.Data {
		if d.Images.FixedWidth.URL == "" || d.Images.Original.URL == "" {
			return provider.Page{}, bad(status) // a silently shorter page would hide the problem
		}
		title := d.Title
		if title == "" {
			title = d.Slug
		}
		items = append(items, provider.Gif{Provider: provider.GIPHY, ID: d.ID, Title: title, PreviewURL: d.Images.FixedWidth.URL, DownloadURL: d.Images.Original.URL})
	}
	next := offset + len(items)
	more := len(items) == provider.PageSize
	if total := int(r.Pagination.TotalCount); total > 0 {
		more = next < total
	}
	p := provider.Page{Items: items}
	if more && len(items) > 0 && next <= offsetCap {
		p.Next = strconv.Itoa(next)
	}
	return p, nil
}
