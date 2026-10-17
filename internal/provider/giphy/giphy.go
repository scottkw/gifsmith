// Package giphy is the GIPHY GIFs client (RED skeleton).
package giphy

import (
	"context"
	"net/http"

	"gifsmith/internal/provider"
)

type Client struct {
	BaseURL string
	Key     func() (string, bool)
	HTTP    *http.Client
}

type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error { return nil }

func (c *Client) Search(ctx context.Context, q, cursor string) (provider.Page, error) {
	return provider.Page{}, nil
}

func (c *Client) Trending(ctx context.Context, cursor string) (provider.Page, error) {
	return provider.Page{}, nil
}
