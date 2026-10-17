// Package fakeapi is a fixture HTTP server that speaks the vendors' wire formats.
// Go tests use it through httptest; the -tags server build points every client
// at it for browser UAT. It never sees a real key.
package fakeapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/gif"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	FinerDown bool          // /finer/search answers like the real host does when it is down (530, Cloudflare 1033)
	SlowDelay time.Duration // used by the "slow" scenario word (plan 02-03)
}

const finerTotal = 60

func Handler(o Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /finer/search", func(w http.ResponseWriter, r *http.Request) { finerSearch(w, r, o) })
	mux.HandleFunc("GET /media/{file}", media)
	return mux
}

func finerSearch(w http.ResponseWriter, r *http.Request, o Options) {
	if o.FinerDown {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(530)
		fmt.Fprint(w, "error code: 1033")
		return
	}
	q := r.URL.Query()
	start, _ := strconv.Atoi(q.Get("start"))
	size, err := strconv.Atoi(q.Get("size"))
	if err != nil || size <= 0 {
		size = 10
	}
	found := finerTotal
	if strings.Contains(q.Get("q"), "empty") {
		found = 0
	}
	type fields struct {
		Name   string `json:"name"`
		Text   string `json:"text"`
		Fileid string `json:"fileid"`
	}
	type result struct {
		ID     string `json:"id"`
		Fields fields `json:"fields"`
	}
	results := []result{}
	for n := max(start, 0) + 1; n <= found && len(results) < size; n++ {
		id := "office-" + strconv.Itoa(n)
		results = append(results, result{ID: strconv.Itoa(n), Fields: fields{Name: id, Text: "Office moment " + strconv.Itoa(n), Fileid: id}})
	}
	writeJSON(w, map[string]any{"hits": map[string]int{"found": found, "start": start}, "results": results})
}

// media serves a tiny 2-frame animated GIF whose colour is derived from the file name.
func media(w http.ResponseWriter, r *http.Request) {
	name, ok := strings.CutSuffix(r.PathValue("file"), ".gif")
	if !ok || name == "" {
		http.NotFound(w, r)
		return
	}
	h := fnv.New32a()
	h.Write([]byte(name))
	sum := h.Sum32()
	c1 := color.RGBA{uint8(sum), uint8(sum >> 8), uint8(sum >> 16), 255}
	c2 := color.RGBA{255 - c1.R, 255 - c1.G, 255 - c1.B, 255}
	anim := &gif.GIF{}
	for _, fg := range []color.RGBA{c1, c2} {
		img := image.NewPaletted(image.Rect(0, 0, 32, 32), color.Palette{fg, color.White})
		for y := 12; y < 20; y++ { // a white bar so the frames differ visibly even for similar colours
			for x := 0; x < 32; x++ {
				img.SetColorIndex(x, y, 1)
			}
		}
		anim.Image = append(anim.Image, img)
		anim.Delay = append(anim.Delay, 50)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, anim); err != nil {
		http.Error(w, "encode", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/gif")
	w.Write(buf.Bytes())
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
