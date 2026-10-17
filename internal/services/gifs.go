package services

import (
	"context"
	"errors"
	"log"

	"gifsmith/internal/provider"
)

var errUnknownProvider = errors.New("unknown provider")

// GifService exposes search and trending to the webview. All provider HTTP stays in Go.
type GifService struct {
	Providers map[provider.ID]provider.Provider
}

func (s *GifService) get(p string) (provider.Provider, error) {
	id, ok := provider.ParseID(p)
	if !ok || s.Providers[id] == nil {
		return nil, errUnknownProvider
	}
	return s.Providers[id], nil
}

func (s *GifService) Search(ctx context.Context, p, q, cursor string) (provider.Page, error) {
	pr, err := s.get(p)
	if err != nil {
		return provider.Page{}, err
	}
	page, err := pr.Search(ctx, q, cursor)
	logProviderErr("search", p, err)
	return page, err
}

func (s *GifService) Trending(ctx context.Context, p, cursor string) (provider.Page, error) {
	pr, err := s.get(p)
	if err != nil {
		return provider.Page{}, err
	}
	page, err := pr.Trending(ctx, cursor)
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
