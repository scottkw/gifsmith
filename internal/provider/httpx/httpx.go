// Package httpx is the single HTTP choke point for provider APIs. It classifies
// every failure into a fresh *provider.Error and never wraps or formats a
// transport error: *url.Error text holds the full URL, including API keys.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"gifsmith/internal/provider"
)

var UserAgent = "Gifsmith/0.2"

const maxBody = 8 << 20

// GetJSON GETs rawURL and decodes a 2xx JSON body into v. A cancelled or expired
// caller context comes back as ctx.Err() (not a provider failure, not logged).
func GetJSON(ctx context.Context, c *http.Client, p provider.ID, rawURL string, v any) (status int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, &provider.Error{Kind: provider.HTTP, Provider: p} // malformed URL is a bug; never echo it
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return 0, &provider.Error{Kind: provider.Timeout, Provider: p}
		}
		return 0, &provider.Error{Kind: provider.Network, Provider: p}
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return resp.StatusCode, &provider.Error{Kind: provider.Timeout, Provider: p}
		}
		return resp.StatusCode, &provider.Error{Kind: provider.Network, Provider: p}
	}
	s := resp.StatusCode
	switch {
	case s == http.StatusTooManyRequests:
		return s, &provider.Error{Kind: provider.RateLimited, Provider: p, RetryAfterSec: retryAfter(resp.Header.Get("Retry-After"), time.Now())}
	case s < 200 || s > 299:
		return s, &provider.Error{Kind: provider.HTTP, Provider: p, Status: s}
	case len(bytes.TrimSpace(body)) == 0: // GIPHY documents an empty 200 as an error
		return s, &provider.Error{Kind: provider.HTTP, Provider: p, Status: s}
	}
	if err := json.Unmarshal(body, v); err != nil {
		return s, &provider.Error{Kind: provider.HTTP, Provider: p, Status: s}
	}
	return s, nil
}

// retryAfter reads Retry-After as seconds or an HTTP-date; <=0, absent or garbage is 0 (unknown).
func retryAfter(h string, now time.Time) int {
	h = strings.TrimSpace(h)
	if n, err := strconv.Atoi(h); err == nil {
		return max(n, 0)
	}
	if t, err := http.ParseTime(h); err == nil {
		return max(int(t.Sub(now).Seconds()+0.999), 0)
	}
	return 0
}

// Remap turns an HTTP-kind error whose status is in reject into KeyRejected
// (per-vendor bad-key statuses: GIPHY 401/403, Klipy 401/403/404).
func Remap(err error, reject ...int) error {
	var pe *provider.Error
	if errors.As(err, &pe) && pe.Kind == provider.HTTP && slices.Contains(reject, pe.Status) {
		return &provider.Error{Kind: provider.KeyRejected, Provider: pe.Provider, Status: pe.Status}
	}
	return err
}
