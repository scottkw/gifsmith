package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
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

// WR-02: vendor text is clamped to the library limits, never splitting a rune.
func TestClamp(t *testing.T) {
	long := strings.Repeat("ü", MaxTitleRunes+7)
	if got := ClampTitle(long); utf8.RuneCountInString(got) != MaxTitleRunes || !utf8.ValidString(got) {
		t.Errorf("ClampTitle: %d runes, valid %v; want %d", utf8.RuneCountInString(got), utf8.ValidString(got), MaxTitleRunes)
	}
	if got := ClampTitle("Cat GIF"); got != "Cat GIF" {
		t.Errorf("ClampTitle(short) = %q", got)
	}
	odd := "a" + strings.Repeat("é", MaxIDBytes) // 2-byte runes straddle the byte limit
	if got := ClampID(odd); len(got) != MaxIDBytes-1 || !utf8.ValidString(got) || !strings.HasPrefix(odd, got) {
		t.Errorf("ClampID: %d bytes, valid %v; want %d valid bytes", len(got), utf8.ValidString(got), MaxIDBytes-1)
	}
	if got := ClampID(strings.Repeat("a", MaxIDBytes)); len(got) != MaxIDBytes {
		t.Errorf("ClampID(at limit) = %d bytes, want %d", len(got), MaxIDBytes)
	}
}

func TestDefaultParams(t *testing.T) {
	if got, want := DefaultParams(), (Params{PageSize: 24, Rating: "pg-13", Lang: "en", Country: "us"}); got != want {
		t.Errorf("DefaultParams() = %+v, want %+v", got, want)
	}
	if fmt.Sprint(PageSizes) != "[24 36 48]" || strings.Join(Ratings, " ") != "g pg pg-13 r" {
		t.Errorf("PageSizes %v, Ratings %v", PageSizes, Ratings)
	}
}
