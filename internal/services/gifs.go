package services

import (
	"context"
	"errors"
	"log"

	"gifsmith/internal/provider"
	"gifsmith/internal/search"
	"gifsmith/internal/settings"
)

var errUnknownProvider = errors.New("unknown provider")

// GifService exposes search and trending to the webview. All provider HTTP stays
// in Go, behind search.Service (cache, 429 budget, dedupe policy).
type GifService struct {
	Svc  *search.Service
	Keys *settings.Keys
}

// ProviderStatus crosses to the webview: booleans and state strings only, never a key.
type ProviderStatus struct {
	ID         string `json:"id"`
	Configured bool   `json:"configured"`
	Clips      string `json:"clips"`
}

// Providers reports the four fixed slots. Clips needs the GIPHY key; plan 02-06
// adds its probe states (checking/enabled/disabled) in place of "unknown".
func (s *GifService) Providers(ctx context.Context) []ProviderStatus {
	_, giphy := s.Keys.Get(provider.GIPHY)
	_, klipy := s.Keys.Get(provider.Klipy)
	clips := "no_key"
	if giphy {
		clips = "unknown"
	}
	return []ProviderStatus{
		{ID: string(provider.GIPHY), Configured: giphy, Clips: clips},
		{ID: string(provider.Clips), Configured: giphy, Clips: clips},
		{ID: string(provider.Klipy), Configured: klipy, Clips: clips},
		{ID: string(provider.FinerGifs), Configured: true, Clips: clips},
	}
}

func parse(p string) (provider.ID, error) {
	id, ok := provider.ParseID(p)
	if !ok {
		return "", errUnknownProvider
	}
	return id, nil
}

func (s *GifService) Search(ctx context.Context, p, q, cursor string) (provider.Page, error) {
	id, err := parse(p)
	if err != nil {
		return provider.Page{}, err
	}
	page, err := s.Svc.Search(ctx, id, q, cursor)
	logProviderErr("search", p, err)
	return page, err
}

func (s *GifService) Trending(ctx context.Context, p, cursor string) (provider.Page, error) {
	id, err := parse(p)
	if err != nil {
		return provider.Page{}, err
	}
	page, err := s.Svc.Trending(ctx, id, cursor)
	logProviderErr("trending", p, err)
	return page, err
}

// logProviderErr logs only typed provider errors (URL-free by construction); cancellations are routine.
func logProviderErr(op, p string, err error) {
	var pe *provider.Error
	if errors.As(err, &pe) {
		log.Printf("%s %s: %v", op, p, pe)
	}
}
