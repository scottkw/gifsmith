// Package giphy is the GIPHY client: GIFs (api.giphy.com/v1/gifs) and, with Clips set, GIPHY Clips (/v1/clips).
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
	Clips   bool // GIPHY Clips endpoints: provider ID clips, no download URL until Phase 3 (MP4)
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

func (c *Client) id() provider.ID {
	if c.Clips {
		return provider.Clips
	}
	return provider.GIPHY
}

func (c *Client) bad(status int) error {
	return &provider.Error{Kind: provider.HTTP, Provider: c.id(), Status: status}
}

// path is /v1/gifs/<op> or /v1/clips/<op>.
func (c *Client) path(op string) string {
	if c.Clips {
		return "/v1/clips/" + op
	}
	return "/v1/gifs/" + op
}

func (c *Client) Search(ctx context.Context, q, cursor string) (provider.Page, error) {
	if r := []rune(q); len(r) > maxQueryRunes {
		q = string(r[:maxQueryRunes])
	}
	return c.get(ctx, c.path("search"), url.Values{"q": {q}}, cursor, searchOffsetCap)
}

func (c *Client) Trending(ctx context.Context, cursor string) (provider.Page, error) {
	return c.get(ctx, c.path("trending"), url.Values{}, cursor, trendingOffsetCap)
}

// ProbeClips asks whether the key is approved for Clips: nil on a 2xx JSON
// answer, key_rejected on 401/403. A 401 means nothing on its own (a bad key
// gets it too), so callers probe only after a GIFs call succeeded with the same key.
func (c *Client) ProbeClips(ctx context.Context) error {
	key, ok := c.Key()
	if !ok || key == "" {
		return &provider.Error{Kind: provider.KeyMissing, Provider: provider.Clips}
	}
	v := url.Values{"api_key": {key}, "limit": {"1"}, "rating": {"pg-13"}}
	var body any
	_, err := httpx.GetJSON(ctx, c.HTTP, provider.Clips, c.BaseURL+"/v1/clips/trending?"+v.Encode(), &body)
	return httpx.Remap(err, 401, 403)
}

// get pages with an opaque cursor that is the decimal offset.
func (c *Client) get(ctx context.Context, path string, v url.Values, cursor string, offsetCap int) (provider.Page, error) {
	offset := 0
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil || n < 0 {
			return provider.Page{}, c.bad(0) // never silently restart at the first page
		}
		offset = n
	}
	key, ok := c.Key()
	if !ok || key == "" {
		return provider.Page{}, &provider.Error{Kind: provider.KeyMissing, Provider: c.id()}
	}
	v.Set("api_key", key)
	v.Set("limit", strconv.Itoa(provider.PageSize))
	v.Set("offset", strconv.Itoa(offset))
	v.Set("rating", "pg-13") // always sent: Clips search defaults to R
	if c.Clips {
		v.Set("bundle", "clips_grid_picker")
	} else {
		v.Set("bundle", "messaging_non_clips")
	}

	var r response
	status, err := httpx.GetJSON(ctx, c.HTTP, c.id(), c.BaseURL+path+"?"+v.Encode(), &r)
	if err != nil {
		return provider.Page{}, httpx.Remap(err, 401, 403)
	}
	if r.Data == nil {
		return provider.Page{}, c.bad(status)
	}
	items := make([]provider.Gif, 0, len(r.Data))
	for _, d := range r.Data {
		// Clips carry no download URL until Phase 3 copies the MP4.
		if d.Images.FixedWidth.URL == "" || (!c.Clips && d.Images.Original.URL == "") {
			return provider.Page{}, c.bad(status) // a silently shorter page would hide the problem
		}
		title := d.Title
		if title == "" {
			title = d.Slug
		}
		g := provider.Gif{Provider: c.id(), ID: d.ID, Title: title, PreviewURL: d.Images.FixedWidth.URL, IsClip: c.Clips}
		if !c.Clips {
			g.DownloadURL = d.Images.Original.URL
		}
		items = append(items, g)
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
