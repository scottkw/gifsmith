//go:build server

package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"gifsmith/internal/platform"
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
		FinerDown:   os.Getenv("FAKE_FINER") == "down",
		ClipsDenied: os.Getenv("FAKE_CLIPS") == "denied",
		SlowDelay:   5 * time.Second, // past the 3 s client timeout below, so "slow" times out quickly
	})
	go func() { log.Fatal(http.Serve(ln, h)) }()
	host := ln.Addr().String()
	log.Printf("fixture server: http://%s", host)
	return deps{
		HTTP:         &http.Client{Timeout: 3 * time.Second},
		FinerBase:    "http://" + host + "/finer",
		FinerMedia:   "http://" + host + "/media",
		GiphyBase:    "http://" + host + "/giphy",
		KlipyBase:    "http://" + host + "/klipy",
		Secrets:      settings.NewMemoryStore(), // never the user's Keychain
		ConfigDir:    "bin/uat-config",
		StagingDir:   "bin/uat-cache/staging",
		DownloadsDir: "bin/uat-downloads", // the smoke script creates it; never the user's ~/Downloads
		// Vendor hosts, plus plain http to the fixture listener only.
		Allow: func(p provider.ID, u *url.URL) bool {
			return staging.VendorAllow(p, u) || (u.Scheme == "http" && u.Host == host)
		},
		// Never the real pasteboard or browser; Trusted false, so paste reports needsAccessibility.
		Platform: &platform.Fake{},
		// Never the real login item; FAKE_AUTOSTART=fail makes Enable/Disable fail (UAT G2).
		Autostart: &memAutostart{fail: os.Getenv("FAKE_AUTOSTART") == "fail"},
	}
}

// memAutostart is the UAT build's login item: a flag in memory.
type memAutostart struct {
	mu            sync.Mutex
	enabled, fail bool
}

func (a *memAutostart) set(on bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.fail {
		return errors.New("autostart fixture failure")
	}
	a.enabled = on
	return nil
}

func (a *memAutostart) Enable() error  { return a.set(true) }
func (a *memAutostart) Disable() error { return a.set(false) }

func (a *memAutostart) IsEnabled() (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.enabled, nil
}
