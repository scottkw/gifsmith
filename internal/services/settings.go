package services

import (
	"context"

	"gifsmith/internal/provider"
	"gifsmith/internal/search"
	"gifsmith/internal/settings"
)

// SettingsService (RED skeleton).
type SettingsService struct {
	Keys      *settings.Keys
	Dir       string
	Svc       *search.Service
	NewClient func(p provider.ID, key string) provider.Provider
	Emit      func(name string)
}

func (s *SettingsService) SaveKey(ctx context.Context, p, key string) error    { return nil }
func (s *SettingsService) LastProvider(ctx context.Context) string             { return "" }
func (s *SettingsService) SetLastProvider(ctx context.Context, p string) error { return nil }
