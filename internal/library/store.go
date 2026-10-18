// Package library is a skeleton (RED).
package library

import (
	"errors"
	"net/url"
	"time"

	"gifsmith/internal/provider"
)

const (
	RecentsMax   = 50
	FavoritesMax = 1000
	SearchesMax  = 20
)

var (
	ErrUnreadable = errors.New("library file unreadable")
	ErrInvalid    = errors.New("invalid library item")
)

type Entry struct {
	Gif provider.Gif `json:"gif"`
	At  int64        `json:"at"`
}

type Snapshot struct {
	Favorites  []Entry             `json:"favorites"`
	Recents    []Entry             `json:"recents"`
	Searches   map[string][]string `json:"searches"`
	Unreadable []string            `json:"unreadable"`
}

type Store struct {
	Dir   string
	Allow func(provider.ID, *url.URL) bool
	Now   func() time.Time
}

func Validate(g provider.Gif, allow func(provider.ID, *url.URL) bool) error { return nil }

func (s *Store) Snapshot() Snapshot         { return Snapshot{} }
func (s *Store) Touch(g provider.Gif) error { return nil }
