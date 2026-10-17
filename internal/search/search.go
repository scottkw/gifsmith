// Package search sits in front of every provider: TTL cache, 429 budget and
// the dedupe policy. Domain code only: no Wails imports, nothing on disk.
package search

import (
	"context"
	"errors"
	"time"

	"gifsmith/internal/provider"
)

const (
	TrendingTTL = 10 * time.Minute
	SearchTTL   = 5 * time.Minute
	MinBlock    = 60 * time.Second
)

var ErrUnknownProvider = errors.New("unknown provider")

type Service struct {
	providers map[provider.ID]provider.Provider
	now       func() time.Time
}

func New(providers map[provider.ID]provider.Provider, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{providers: providers, now: now}
}

func (s *Service) Search(ctx context.Context, p provider.ID, q, cursor string) (provider.Page, error) {
	pr := s.providers[p]
	if pr == nil {
		return provider.Page{}, ErrUnknownProvider
	}
	return pr.Search(ctx, q, cursor)
}

func (s *Service) Trending(ctx context.Context, p provider.ID, cursor string) (provider.Page, error) {
	pr := s.providers[p]
	if pr == nil {
		return provider.Page{}, ErrUnknownProvider
	}
	return pr.Trending(ctx, cursor)
}

func (s *Service) Seed(p provider.ID, cursor string, pg provider.Page) {}

func (s *Service) Invalidate(p provider.ID) {}
