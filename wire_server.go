//go:build server

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"gifsmith/internal/provider/fakeapi"
)

// wire points every provider at an in-process fixture server on 127.0.0.1 (browser UAT only, never shipped).
func wire() deps {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	h := fakeapi.Handler(fakeapi.Options{FinerDown: os.Getenv("FAKE_FINER") == "down"})
	go func() { log.Fatal(http.Serve(ln, h)) }()
	base := "http://" + ln.Addr().String()
	log.Printf("fixture server: %s", base)
	return deps{
		HTTP:       &http.Client{Timeout: 3 * time.Second},
		FinerBase:  base + "/finer",
		FinerMedia: base + "/media",
	}
}
