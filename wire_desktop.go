//go:build !server

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gifsmith/internal/staging"
)

func wire() deps {
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}
	return deps{
		HTTP:       &http.Client{Timeout: 15 * time.Second},
		FinerBase:  "https://api.thefinergifs.club",
		FinerMedia: "https://media.thefinergifs.club",
		StagingDir: filepath.Join(cache, "com.kenscott.gifsmith", "staging"),
		Allow:      staging.VendorAllow,
	}
}
