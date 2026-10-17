package provider

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorString(t *testing.T) {
	cases := []struct {
		e    *Error
		want string
	}{
		{&Error{Kind: HTTP, Provider: FinerGifs, Status: 530}, "finergifs: http (HTTP 530)"},
		{&Error{Kind: Network, Provider: GIPHY}, "giphy: network"},
		{&Error{Kind: RateLimited, Provider: Klipy, RetryAfterSec: 120}, "klipy: rate_limited"},
	}
	for _, c := range cases {
		if got := c.e.Error(); got != c.want {
			t.Errorf("Error() = %q, want %q", got, c.want)
		}
	}

	wrapped := fmt.Errorf("x: %w", &Error{Kind: Timeout, Provider: Klipy})
	if got, want := string(MarshalError(wrapped)), `{"kind":"timeout","provider":"klipy"}`; got != want {
		t.Errorf("MarshalError(wrapped) = %s, want %s", got, want)
	}
	if got, want := string(MarshalError(&Error{Kind: RateLimited, Provider: GIPHY, Status: 429, RetryAfterSec: 90})),
		`{"kind":"rate_limited","provider":"giphy","status":429,"retryAfterSec":90}`; got != want {
		t.Errorf("MarshalError = %s, want %s", got, want)
	}
	if b := MarshalError(errors.New("plain")); b != nil {
		t.Errorf("MarshalError(plain) = %s, want nil (Wails default)", b)
	}
}

func TestParseID(t *testing.T) {
	for _, s := range []string{"giphy", "clips", "klipy", "finergifs"} {
		if id, ok := ParseID(s); !ok || string(id) != s {
			t.Errorf("ParseID(%q) = %q, %v", s, id, ok)
		}
	}
	for _, s := range []string{"", "GIPHY", "tenor", "finergifs "} {
		if _, ok := ParseID(s); ok {
			t.Errorf("ParseID(%q) accepted", s)
		}
	}
}
