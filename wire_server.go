//go:build server

package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"gifsmith/internal/provider"
	"gifsmith/internal/provider/fakeapi"
	"gifsmith/internal/settings"
	"gifsmith/internal/staging"
)

// wire points every provider at an in-process fixture server on 127.0.0.1 (browser UAT only, never shipped).
func wire() deps {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	h := fakeapi.Handler(fakeapi.Options{
		FinerDown: os.Getenv("FAKE_FINER") == "down",
		SlowDelay: 5 * time.Second, // past the 3 s client timeout below, so "slow" times out quickly
	})
	go func() { log.Fatal(http.Serve(ln, h)) }()
	host := ln.Addr().String()
	log.Printf("fixture server: http://%s", host)
	return deps{
		HTTP:       &http.Client{Timeout: 3 * time.Second},
		FinerBase:  "http://" + host + "/finer",
		FinerMedia: "http://" + host + "/media",
		GiphyBase:  "http://" + host + "/giphy",
		KlipyBase:  "http://" + host + "/klipy",
		Secrets:    settings.NewMemoryStore(), // never the user's Keychain
		ConfigDir:  "bin/uat-config",
		StagingDir: "bin/uat-cache/staging",
		// Vendor hosts, plus plain http to the fixture listener only.
		Allow: func(p provider.ID, u *url.URL) bool {
			return staging.VendorAllow(p, u) || (u.Scheme == "http" && u.Host == host)
		},
	}
}
