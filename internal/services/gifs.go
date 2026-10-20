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
	Svc   *search.Service
	Keys  *settings.Keys
	Clips *search.Clips
}

// ProviderStatus crosses to the webview: booleans and state strings only, never a key.
type ProviderStatus struct {
	ID         string `json:"id"`
	Configured bool   `json:"configured"`
	Clips      string `json:"clips"`
}

// Providers reports the four fixed slots; clips is the Clips probe state
// (no_key, checking, enabled, disabled or unknown).
func (s *GifService) Providers(ctx context.Context) []ProviderStatus {
	_, giphy := s.Keys.Get(provider.GIPHY)
	_, klipy := s.Keys.Get(provider.Klipy)
	clips := string(s.Clips.State())
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
	return s.call("search", p, func(id provider.ID) (provider.Page, error) { return s.Svc.Search(ctx, id, q, cursor) })
}

func (s *GifService) Trending(ctx context.Context, p, cursor string) (provider.Page, error) {
	return s.call("trending", p, func(id provider.ID) (provider.Page, error) { return s.Svc.Trending(ctx, id, cursor) })
}

// call refuses Clips until its probe passed, and lets the first GIPHY GIFs
// success start the launch's Clips probe (the key is now proven on GIFs).
func (s *GifService) call(op, p string, get func(provider.ID) (provider.Page, error)) (provider.Page, error) {
	id, err := parse(p)
	if err != nil {
		return provider.Page{}, err
	}
	if id == provider.Clips && s.Clips.State() != search.ClipsEnabled {
		return provider.Page{}, &provider.Error{Kind: provider.ClipsDisabled, Provider: provider.Clips}
	}
	page, err := get(id)
	logProviderErr(op, p, err)
	var pe *provider.Error
	if errors.As(err, &pe) && pe.Kind == provider.KeyRejected {
		kp := id
		if id == provider.Clips {
			kp = provider.GIPHY // Clips use the GIPHY key
		}
		s.Keys.MarkRejected(kp) // the Keys pane shows Rejected, as the panel does
	}
	if err == nil && id == provider.GIPHY {
		s.Clips.Trigger()
	}
	return page, err
}

// CheckClips re-probes once when the Clips result is unknown (the user chose
// Clips). It returns the state, and the probe's error while still unknown.
func (s *GifService) CheckClips(ctx context.Context) (string, error) {
	st, err := s.Clips.Recheck(ctx)
	logProviderErr("clips probe", string(provider.Clips), err)
	return string(st), err
}

// logProviderErr logs only typed provider errors (URL-free by construction); cancellations are routine.
func logProviderErr(op, p string, err error) {
	var pe *provider.Error
	if errors.As(err, &pe) {
		log.Printf("%s %s: %v", op, p, pe)
	}
}
