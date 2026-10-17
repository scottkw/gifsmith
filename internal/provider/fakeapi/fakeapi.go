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

const (
	finerTotal = 60
	giphyTotal = 60
)

func Handler(o Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /finer/search", func(w http.ResponseWriter, r *http.Request) { finerSearch(w, r, o) })
	mux.HandleFunc("GET /giphy/v1/gifs/{kind}", func(w http.ResponseWriter, r *http.Request) { giphyGifs(w, r, o) })
	mux.HandleFunc("GET /media/{file}", media)
	return mux
}

func finerSearch(w http.ResponseWriter, r *http.Request, o Options) {
	if o.FinerDown {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(530)
		_, _ = fmt.Fprint(w, "error code: 1033")
		return
	}
	q := r.URL.Query()
	start, _ := strconv.Atoi(q.Get("start"))
	if Scenario(w, r, q.Get("q"), start <= 0, o) {
		return
	}
	size, err := strconv.Atoi(q.Get("size"))
	if err != nil || size <= 0 {
		size = 10
	}
	found := finerTotal
	if hasWord(q.Get("q"), "empty") {
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

// giphyGifs serves /v1/gifs/search and /v1/gifs/trending. Scenario words are
// read from q and the key, so trending (no q) can be driven by a key such as "ratelimit-key".
func giphyGifs(w http.ResponseWriter, r *http.Request, o Options) {
	kind := r.PathValue("kind")
	if kind != "search" && kind != "trending" {
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query()
	key := q.Get("api_key")
	if key == "" || strings.HasPrefix(key, "bad") {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "0") // what the real host sends on 401
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"data": [], "meta": {"status": 401, "msg": "Unauthorized", "response_id": ""}}`)
		return
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	words := q.Get("q") + " " + key
	if Scenario(w, r, words, offset <= 0, o) {
		return
	}
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 25
	}
	total := giphyTotal
	if hasWord(words, "empty") {
		total = 0
	}
	type rendition struct {
		URL    string `json:"url"`
		Width  string `json:"width"` // strings, like the real wire format
		Height string `json:"height"`
		Size   string `json:"size"`
	}
	type item struct {
		Type   string               `json:"type"`
		ID     string               `json:"id"`
		Slug   string               `json:"slug"`
		Title  string               `json:"title"`
		Images map[string]rendition `json:"images"`
	}
	media := "http://" + r.Host + "/media/"
	data := []item{}
	for n := max(offset, 0) + 1; n <= total && len(data) < limit; n++ {
		id := "giphy-" + strconv.Itoa(n)
		data = append(data, item{Type: "gif", ID: id, Slug: "moment-" + id, Title: "GIPHY moment " + strconv.Itoa(n), Images: map[string]rendition{
			"fixed_width": {URL: media + id + "-200w.gif", Width: "200", Height: "200", Size: "1234"},
			"original":    {URL: media + id + ".gif", Width: "480", Height: "480", Size: "5678"},
		}})
	}
	writeJSON(w, map[string]any{
		"data":       data,
		"pagination": map[string]int{"total_count": total, "count": len(data), "offset": offset},
		"meta":       map[string]any{"status": 200, "msg": "OK", "response_id": "fixture"},
	})
}

func hasWord(q, word string) bool { return strings.Contains(strings.ToLower(q), word) }

// Scenario answers the vendor-neutral scenario words found in the query
// (case-insensitive) and reports whether it wrote the response. "empty" is
// per route because the zero-result body is vendor-specific.
//
//	ratelimit -> 429 + Retry-After: 120     noretry  -> 429 without Retry-After
//	slow      -> sleep o.SlowDelay, then normal   boom -> 500
//	blank     -> 200 with an empty body
//	pagelimit -> later pages 429 + Retry-After: 120; pageboom -> later pages 500
func Scenario(w http.ResponseWriter, r *http.Request, q string, firstPage bool, o Options) bool {
	switch {
	case hasWord(q, "pageboom"):
		if !firstPage {
			w.WriteHeader(http.StatusInternalServerError)
			return true
		}
	case hasWord(q, "pagelimit"):
		if !firstPage {
			w.Header().Set("Retry-After", "120")
			w.WriteHeader(http.StatusTooManyRequests)
			return true
		}
	case hasWord(q, "ratelimit"):
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
		return true
	case hasWord(q, "noretry"):
		w.WriteHeader(http.StatusTooManyRequests)
		return true
	case hasWord(q, "boom"):
		w.WriteHeader(http.StatusInternalServerError)
		return true
	case hasWord(q, "blank"):
		w.WriteHeader(http.StatusOK)
		return true
	case hasWord(q, "slow"):
		select {
		case <-time.After(o.SlowDelay):
		case <-r.Context().Done(): // the client gave up; do not hold the handler
			return true
		}
	}
	return false
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
	_, _ = w.Write(buf.Bytes())
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
