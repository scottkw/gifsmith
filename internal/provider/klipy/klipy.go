// Package klipy is the Klipy native v1 client (RED skeleton).
package klipy

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

func (c *Client) Search(ctx context.Context, q, cursor string) (provider.Page, error) {
	return provider.Page{}, nil
}

func (c *Client) Trending(ctx context.Context, cursor string) (provider.Page, error) {
	return provider.Page{}, nil
}
