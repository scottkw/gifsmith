// Package finergifs is The Finer Gifs Club client (keyless; The Office GIFs).
package finergifs

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/httpx"
)

// fileid becomes part of a URL and (via the id) a staging path; anything else is a malformed response.
var fileidRE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type Client struct {
	BaseURL   string // https://api.thefinergifs.club
	MediaBase string // https://media.thefinergifs.club
	HTTP      *http.Client
}

type response struct {
	Results []struct {
		Fields struct {
			Name   string `json:"name"`
			Text   string `json:"text"`
			Fileid string `json:"fileid"`
		} `json:"fields"`
	} `json:"results"`
	Hits *struct {
		Found int `json:"found"`
		Start int `json:"start"`
	} `json:"hits"`
}

func (c *Client) bad() error {
	return &provider.Error{Kind: provider.HTTP, Provider: provider.FinerGifs}
}

// Search pages with an opaque cursor that is the start offset.
func (c *Client) Search(ctx context.Context, q, cursor string) (provider.Page, error) {
	start := 0
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil || n < 0 {
			return provider.Page{}, c.bad() // never silently restart at the first page
		}
		start = n
	}
	v := url.Values{
		"q":        {q},
		"q.parser": {"simple"},
		"sort":     {"_score desc"},
		"size":     {strconv.Itoa(provider.PageSize)},
		"start":    {strconv.Itoa(start)},
	}
	var r response
	if _, err := httpx.GetJSON(ctx, c.HTTP, provider.FinerGifs, c.BaseURL+"/search?"+v.Encode(), &r); err != nil {
		return provider.Page{}, err
	}
	if r.Results == nil {
		return provider.Page{}, c.bad()
	}
	items := make([]provider.Gif, 0, len(r.Results))
	for _, res := range r.Results {
		f := res.Fields
		if !fileidRE.MatchString(f.Fileid) {
			return provider.Page{}, c.bad()
		}
		title := f.Text
		if title == "" {
			title = f.Name
		}
		u := c.MediaBase + "/" + f.Fileid + ".gif" // the only rendition: grid and copy share it
		items = append(items, provider.Gif{Provider: provider.FinerGifs, ID: f.Fileid, Title: title, PreviewURL: u, DownloadURL: u})
	}
	next := start + len(items)
	more := len(items) == provider.PageSize
	if r.Hits != nil {
		more = next < r.Hits.Found
	}
	p := provider.Page{Items: items}
	if more && len(items) > 0 {
		p.Next = strconv.Itoa(next)
	}
	return p, nil
}

// Trending: Finer Gifs has no trending list, so this is an empty final page and makes no request.
func (c *Client) Trending(ctx context.Context, cursor string) (provider.Page, error) {
	return provider.Page{Items: []provider.Gif{}}, nil
}
