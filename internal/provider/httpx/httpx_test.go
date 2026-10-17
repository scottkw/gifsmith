package httpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gifsmith/internal/provider"
)

func serve(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func asProvider(t *testing.T, err error) *provider.Error {
	t.Helper()
	var pe *provider.Error
	if !errors.As(err, &pe) {
		t.Fatalf("err = %v (%T), want *provider.Error", err, err)
	}
	return pe
}

func TestGetJSONClassification(t *testing.T) {
	date90 := time.Now().Add(90 * time.Second).UTC().Format(http.TimeFormat)
	cases := []struct {
		name       string
		status     int
		header     map[string]string
		body       string
		wantKind   provider.Kind // "" = success
		wantStatus int
		wantRetry  int
		retryTol   int
	}{
		{name: "200 json", status: 200, body: `{"ok":true}`},
		{name: "429 seconds", status: 429, header: map[string]string{"Retry-After": "120"}, body: "{}", wantKind: provider.RateLimited, wantRetry: 120},
		{name: "429 http-date", status: 429, header: map[string]string{"Retry-After": date90}, body: "{}", wantKind: provider.RateLimited, wantRetry: 90, retryTol: 1},
		{name: "429 zero", status: 429, header: map[string]string{"Retry-After": "0"}, wantKind: provider.RateLimited},
		{name: "429 negative", status: 429, header: map[string]string{"Retry-After": "-5"}, wantKind: provider.RateLimited},
		{name: "429 garbage", status: 429, header: map[string]string{"Retry-After": "soon"}, wantKind: provider.RateLimited},
		{name: "429 absent", status: 429, wantKind: provider.RateLimited},
		{name: "401 ignores retry-after", status: 401, header: map[string]string{"Retry-After": "30"}, body: `{"data":[]}`, wantKind: provider.HTTP, wantStatus: 401},
		{name: "500", status: 500, body: "oops", wantKind: provider.HTTP, wantStatus: 500},
		{name: "530 text", status: 530, header: map[string]string{"Content-Type": "text/plain"}, body: "error code: 1033", wantKind: provider.HTTP, wantStatus: 530},
		{name: "200 empty", status: 200, body: "", wantKind: provider.HTTP, wantStatus: 200},
		{name: "200 whitespace", status: 200, body: " \n ", wantKind: provider.HTTP, wantStatus: 200},
		{name: "200 not json", status: 200, body: "not json", wantKind: provider.HTTP, wantStatus: 200},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("User-Agent") != UserAgent {
					t.Errorf("User-Agent = %q, want %q", r.Header.Get("User-Agent"), UserAgent)
				}
				for k, v := range c.header {
					w.Header().Set(k, v)
				}
				w.WriteHeader(c.status)
				_, _ = fmt.Fprint(w, c.body)
			})
			var v map[string]any
			status, err := GetJSON(context.Background(), srv.Client(), provider.GIPHY, srv.URL+"/x", &v)
			if status != c.status {
				t.Errorf("status = %d, want %d", status, c.status)
			}
			if c.wantKind == "" {
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				if v["ok"] != true {
					t.Errorf("decoded %v, want ok:true", v)
				}
				return
			}
			pe := asProvider(t, err)
			if pe.Kind != c.wantKind || pe.Status != c.wantStatus || pe.Provider != provider.GIPHY {
				t.Errorf("err = %+v, want kind %s status %d provider giphy", pe, c.wantKind, c.wantStatus)
			}
			if d := pe.RetryAfterSec - c.wantRetry; d < -c.retryTol || d > c.retryTol {
				t.Errorf("RetryAfterSec = %d, want %d (±%d)", pe.RetryAfterSec, c.wantRetry, c.retryTol)
			}
		})
	}
}

func TestGetJSONTimeout(t *testing.T) {
	srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	})
	c := &http.Client{Timeout: 50 * time.Millisecond}
	_, err := GetJSON(context.Background(), c, provider.Klipy, srv.URL, new(any))
	if pe := asProvider(t, err); pe.Kind != provider.Timeout {
		t.Errorf("kind = %s, want timeout", pe.Kind)
	}
}

func TestGetJSONNetwork(t *testing.T) {
	_, err := GetJSON(context.Background(), &http.Client{Timeout: time.Second}, provider.FinerGifs, "http://127.0.0.1:1/search", new(any))
	if pe := asProvider(t, err); pe.Kind != provider.Network || pe.Status != 0 {
		t.Errorf("err = %+v, want network with no status", pe)
	}
}

func TestGetJSONCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	})
	go func() { <-started; cancel() }()
	_, err := GetJSON(ctx, srv.Client(), provider.GIPHY, srv.URL, new(any))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	var pe *provider.Error
	if errors.As(err, &pe) {
		t.Errorf("cancel became a provider error %+v; superseded calls are not provider failures", pe)
	}
}

func TestRedaction(t *testing.T) {
	const secret = "SECRETKEY123"
	slow := serve(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	})
	cases := []struct {
		name string
		c    *http.Client
		url  string
	}{
		{"refused", &http.Client{Timeout: time.Second}, "http://127.0.0.1:1/api/v1/" + secret + "/gifs/search?api_key=" + secret},
		{"timeout", &http.Client{Timeout: 50 * time.Millisecond}, slow.URL + "/api/v1/" + secret + "/gifs/search?api_key=" + secret},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := GetJSON(context.Background(), c.c, provider.Klipy, c.url, new(any))
			if err == nil {
				t.Fatal("err = nil")
			}
			for _, s := range []string{err.Error(), string(provider.MarshalError(err))} {
				for _, leak := range []string{secret, "127.0.0.1", "http"} {
					if strings.Contains(s, leak) {
						t.Errorf("%q leaks %q", s, leak)
					}
				}
			}
		})
	}
}

func TestRemap(t *testing.T) {
	reject := []int{401, 403, 404}
	got := Remap(&provider.Error{Kind: provider.HTTP, Provider: provider.Klipy, Status: 404}, reject...)
	if pe := asProvider(t, got); pe.Kind != provider.KeyRejected || pe.Provider != provider.Klipy {
		t.Errorf("404 remapped to %+v, want key_rejected for klipy", pe)
	}
	rl := &provider.Error{Kind: provider.RateLimited, Provider: provider.Klipy, RetryAfterSec: 60}
	if got := Remap(rl, reject...); got != error(rl) {
		t.Errorf("rate_limited changed to %v", got)
	}
	other := &provider.Error{Kind: provider.HTTP, Provider: provider.GIPHY, Status: 500}
	if got := Remap(other, reject...); got != error(other) {
		t.Errorf("HTTP 500 changed to %v", got)
	}
	plain := errors.New("plain")
	if got := Remap(plain, reject...); got != plain {
		t.Errorf("plain error changed to %v", got)
	}
	if got := Remap(nil, reject...); got != nil {
		t.Errorf("Remap(nil) = %v", got)
	}
}
