//go:build !server

package main

import (
	"net/http"
	"time"
)

func wire() deps {
	return deps{
		HTTP:       &http.Client{Timeout: 15 * time.Second},
		FinerBase:  "https://api.thefinergifs.club",
		FinerMedia: "https://media.thefinergifs.club",
	}
}
