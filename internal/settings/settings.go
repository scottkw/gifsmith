// Package settings holds API keys (RED skeleton).
package settings

import (
	"errors"

	"gifsmith/internal/provider"
)

const KeychainService = "com.kenscott.gifsmith"

var ErrNoKey = errors.New("no key stored")

type SecretStore interface {
	Get(account string) (string, error)
	Set(account, secret string) error
}

type KeyringStore struct{ Service string }

func (s KeyringStore) Get(account string) (string, error) { return "", nil }
func (s KeyringStore) Set(account, secret string) error   { return nil }

type MemoryStore struct{}

func NewMemoryStore() *MemoryStore                        { return &MemoryStore{} }
func (m *MemoryStore) Get(account string) (string, error) { return "", nil }
func (m *MemoryStore) Set(account, secret string) error   { return nil }

type Keys struct{}

func NewKeys(s SecretStore) *Keys                   { return &Keys{} }
func (k *Keys) Load()                               {}
func (k *Keys) Get(p provider.ID) (string, bool)    { return "", false }
func (k *Keys) Set(p provider.ID, key string) error { return nil }
func (k *Keys) Func(p provider.ID) func() (string, bool) {
	return func() (string, bool) { return "", false }
}

type Prefs struct {
	LastProvider string `json:"lastProvider"`
}

func LoadPrefs(dir string) (Prefs, error) { return Prefs{}, nil }
func SavePrefs(dir string, p Prefs) error { return nil }
