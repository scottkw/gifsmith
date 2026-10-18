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
	"gifsmith/internal/staging"
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
	Clips   bool // GIPHY Clips endpoints: provider ID clips, DownloadURL is the clip's MP4 (clipRendition)
	// Params is read on every request (page size, rating, language); nil = provider.DefaultParams().
	Params func() provider.Params
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

// rendition: width/height/size are strings on the wire and display-only, so decoded loosely.
type rendition struct {
	URL    string         `json:"url"`
	Width  httpx.LooseInt `json:"width"`
	Height httpx.LooseInt `json:"height"`
	Size   httpx.LooseInt `json:"size"`
}

type response struct {
	Data []struct {
		ID     string `json:"id"`
		Slug   string `json:"slug"`
		Title  string `json:"title"`
		URL    any    `json:"url"`    // page; display-only, so any type decodes
		Source any    `json:"source"` // third-party page; display-only
		Images struct {
			FixedWidth  rendition `json:"fixed_width"`
			FixedHeight rendition `json:"fixed_height"`
			Original    rendition `json:"original"`
		} `json:"images"`
		Video struct {
			Assets map[string]rendition `json:"assets"` // Clips: MP4 renditions keyed 360p, 720p, 4k, source, …
		} `json:"video"`
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
	// A one-item yes/no probe: the user's rating does not change the answer, so it stays fixed.
	v := url.Values{"api_key": {key}, "limit": {"1"}, "rating": {"pg-13"}}
	var body any
	_, err := httpx.GetJSON(ctx, c.HTTP, provider.Clips, c.BaseURL+"/v1/clips/trending?"+v.Encode(), &body)
	return httpx.Remap(err, 401, 403)
}

// pageURL keeps the vendor page only on giphy.com (http upgraded to https); anything else is dropped.
func pageURL(p provider.ID, v any) string {
	s, _ := v.(string)
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	if u.Scheme == "http" {
		u.Scheme = "https"
	}
	if !staging.PageAllow(p, u) {
		return ""
	}
	return u.String()
}

// sourceURL keeps a third-party source only when it is https with a host.
func sourceURL(v any) string {
	s, _ := v.(string)
	if u, err := url.Parse(s); err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	return s
}

// clipRendition is the MP4 a Clip copies (D-24): 720p, else the nearest of 480p, 1080p, 360p.
// Never 4k or source (too large to paste); none of them -> zero rendition, which fails the page.
func clipRendition(assets map[string]rendition) rendition {
	for _, k := range []string{"720p", "480p", "1080p", "360p"} {
		if a := assets[k]; a.URL != "" {
			return a
		}
	}
	return rendition{}
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
	prm := provider.ParamsOf(c.Params)
	limit := min(prm.PageSize, 50)
	v.Set("api_key", key)
	v.Set("limit", strconv.Itoa(limit))
	v.Set("offset", strconv.Itoa(offset))
	v.Set("rating", prm.Rating) // always sent: Clips search defaults to R
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
		dl := d.Images.Original
		if c.Clips {
			dl = clipRendition(d.Video.Assets)
		}
		if d.Images.FixedWidth.URL == "" || dl.URL == "" {
			return provider.Page{}, c.bad(status) // a silently shorter page would hide the problem
		}
		title := d.Title
		if title == "" {
			title = d.Slug
		}
		g := provider.Gif{Provider: c.id(), ID: provider.ClampID(d.ID), Title: provider.ClampTitle(title), PreviewURL: d.Images.FixedWidth.URL, IsClip: c.Clips,
			PageURL: pageURL(c.id(), d.URL), SourceURL: sourceURL(d.Source)}
		g.DownloadURL, g.Width, g.Height = dl.URL, int(dl.Width), int(dl.Height)
		if c.Clips {
			g.GifURL = d.Images.Original.URL // D-18: links and Markdown use the GIF rendition, not the MP4
		} else {
			g.Size = int64(dl.Size) // Clips assets carry no size
			g.MediumURL = d.Images.FixedHeight.URL
		}
		items = append(items, g)
	}
	next := offset + len(items)
	more := len(items) == limit // the size requested, not a constant
	if total := int(r.Pagination.TotalCount); total > 0 {
		more = next < total
	}
	p := provider.Page{Items: items}
	if more && len(items) > 0 && next <= offsetCap {
		p.Next = strconv.Itoa(next)
	}
	return p, nil
}
