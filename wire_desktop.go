//go:build !server

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gifsmith/internal/platform"
	"gifsmith/internal/settings"
	"gifsmith/internal/staging"
)

func wire() deps {
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}
	config, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return deps{
		HTTP:         &http.Client{Timeout: 15 * time.Second},
		FinerBase:    "https://api.thefinergifs.club",
		FinerMedia:   "https://media.thefinergifs.club",
		GiphyBase:    "https://api.giphy.com",
		KlipyBase:    "https://api.klipy.com",
		Secrets:      settings.KeyringStore{Service: settings.KeychainService},
		ConfigDir:    filepath.Join(config, "com.kenscott.gifsmith"),
		StagingDir:   filepath.Join(cache, "com.kenscott.gifsmith", "staging"),
		DownloadsDir: filepath.Join(home, "Downloads"),
		Allow:        staging.VendorAllow,
		Platform:     platform.New(),
	}
}
