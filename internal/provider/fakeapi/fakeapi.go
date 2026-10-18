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
	"path"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	FinerDown   bool          // /finer/search answers like the real host does when it is down (530, Cloudflare 1033)
	SlowDelay   time.Duration // used by the "slow" scenario word (plan 02-03)
	ClipsDenied bool          // /giphy/v1/clips/* answer 403, as for a key GIPHY has not approved for Clips
}

const (
	finerTotal = 60
	giphyTotal = 60
	klipyTotal = 60
)

func Handler(o Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /finer/search", func(w http.ResponseWriter, r *http.Request) { finerSearch(w, r, o) })
	mux.HandleFunc("GET /giphy/v1/gifs/{kind}", func(w http.ResponseWriter, r *http.Request) { giphyGifs(w, r, o, false) })
	mux.HandleFunc("GET /giphy/v1/clips/{kind}", func(w http.ResponseWriter, r *http.Request) { giphyGifs(w, r, o, true) })
	mux.HandleFunc("GET /klipy/api/v1/{key}/gifs/{kind}", func(w http.ResponseWriter, r *http.Request) { klipyGifs(w, r, o) })
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
	// Failing-media words: fileids whose /media/ file answers 500 or HTML (copy error rows).
	prefix := "office-"
	if hasWord(q.Get("q"), "brokenmedia") {
		prefix = "broken-"
	} else if hasWord(q.Get("q"), "htmlmedia") {
		prefix = "html-"
	} else if hasWord(q.Get("q"), "widemedia") { // 48x32 media (square crop)
		prefix = "wide-"
	} else if hasWord(q.Get("q"), "tallmedia") { // 32x48 media
		prefix = "tall-"
	}
	results := []result{}
	for n := max(start, 0) + 1; n <= found && len(results) < size; n++ {
		id := prefix + strconv.Itoa(n)
		results = append(results, result{ID: strconv.Itoa(n), Fields: fields{Name: id, Text: "Office moment " + strconv.Itoa(n), Fileid: id}})
	}
	writeJSON(w, map[string]any{"hits": map[string]int{"found": found, "start": start}, "results": results})
}

// giphyGifs serves /v1/gifs/{search,trending} and, with clips, /v1/clips/{search,trending}
// (clips_grid_picker-shaped: fixed_width, original GIF and 720p/480p/360p MP4 video assets). Scenario
// words are read from q and the key, so trending (no q) can be driven by a key such as "ratelimit-key".
func giphyGifs(w http.ResponseWriter, r *http.Request, o Options, clips bool) {
	kind := r.PathValue("kind")
	if kind != "search" && kind != "trending" {
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query()
	key := q.Get("api_key")
	// "revoked" in q stands for a stored key the vendor no longer accepts (S13 UAT).
	if key == "" || strings.HasPrefix(key, "bad") || hasWord(q.Get("q"), "revoked") {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "0") // what the real host sends on 401
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"data": [], "meta": {"status": 401, "msg": "Unauthorized", "response_id": ""}}`)
		return
	}
	if clips && o.ClipsDenied {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden) // GIPHY does not document the unapproved-key status; 403 stands in
		_, _ = fmt.Fprint(w, `{"data": [], "meta": {"status": 403, "msg": "Forbidden", "response_id": ""}}`)
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
		URL    string               `json:"url"`
		Source string               `json:"source,omitempty"`
		Images map[string]rendition `json:"images"`
		Video  map[string]any       `json:"video,omitempty"`
	}
	media := "http://" + r.Host + "/media/"
	originalSize := "5678"
	if hasWord(words, "hugesize") { // over the 15 MB square crop limit
		originalSize = "16000000"
	}
	// Non-default request parameters show in titles, so UAT can see what Go sent.
	tag := ""
	if r := q.Get("rating"); r != "" && r != "pg-13" {
		tag = "[rating=" + r + "] "
	}
	data := []item{}
	for n := max(offset, 0) + 1; n <= total && len(data) < limit; n++ {
		if clips {
			id := "clip-" + strconv.Itoa(n)
			data = append(data, item{Type: "video", ID: id, Slug: "clip-" + id, Title: tag + "GIPHY clip " + strconv.Itoa(n), URL: "https://giphy.com/clips/clip-" + id, Images: map[string]rendition{
				"fixed_width": {URL: media + id + "-200w.gif", Width: "200", Height: "112", Size: "1234"},
				"original":    {URL: media + id + ".gif", Width: "480", Height: "270", Size: "4321"},
			}, Video: map[string]any{"assets": clipAssets(media+id, hasWord(words, "no720"))}})
			continue
		}
		id := "giphy-" + strconv.Itoa(n)
		source := "" // odd items carry a source page, even ones none
		if n%2 == 1 {
			source = "https://example.org/source/" + id
		}
		data = append(data, item{Type: "gif", ID: id, Slug: "moment-" + id, Title: tag + "GIPHY moment " + strconv.Itoa(n), URL: "https://giphy.com/gifs/moment-" + id, Source: source, Images: map[string]rendition{
			"fixed_width":  {URL: media + id + "-200w.gif", Width: "200", Height: "200", Size: "1234"},
			"fixed_height": {URL: media + id + "-200h.gif", Width: "200", Height: "200", Size: "1234"},
			"original":     {URL: media + id + ".gif", Width: "480", Height: "480", Size: originalSize},
		}})
	}
	writeJSON(w, map[string]any{
		"data":       data,
		"pagination": map[string]int{"total_count": total, "count": len(data), "offset": offset},
		"meta":       map[string]any{"status": 200, "msg": "OK", "response_id": "fixture"},
	})
}

// klipyGifs serves the native v1 search and trending routes (key in the path).
// Items 1 and 2 share the slug "klipy-dup" so a dedupe anywhere is visible end to end.
func klipyGifs(w http.ResponseWriter, r *http.Request, o Options) {
	kind := r.PathValue("kind")
	if kind != "search" && kind != "trending" {
		http.NotFound(w, r)
		return
	}
	key := r.PathValue("key")
	if strings.HasPrefix(key, "bad") || hasWord(r.URL.Query().Get("q"), "revoked") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // what the real host sends for an invalid key
		_, _ = fmt.Fprint(w, `{"result":false,"errors":{"message":["The provided API key is invalid."]}}`)
		return
	}
	q := r.URL.Query()
	page, err := strconv.Atoi(q.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	words := q.Get("q") + " " + key
	if Scenario(w, r, words, page == 1, o) {
		return
	}
	per, err := strconv.Atoi(q.Get("per_page"))
	if err != nil || per <= 0 {
		per = 24
	}
	total := klipyTotal
	if hasWord(words, "empty") {
		total = 0
	}
	type gifFile struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Size   int    `json:"size"`
	}
	type item struct {
		ID    int64                         `json:"id"`
		Slug  string                        `json:"slug"`
		Title string                        `json:"title"`
		Type  string                        `json:"type"`
		Tags  []any                         `json:"tags"`
		File  map[string]map[string]gifFile `json:"file"`
	}
	media := "http://" + r.Host + "/media/"
	data := []item{}
	first := (page-1)*per + 1
	for n := first; n <= total && n < first+per; n++ {
		name := "klipy-" + strconv.Itoa(n)
		slug := name
		if n <= 2 {
			slug = "klipy-dup"
		}
		tags := []any{"funny", "cat"}
		if n == 3 { // an object element the client must skip without failing the page
			tags = []any{map[string]string{"name": "obj"}, "funny"}
		}
		data = append(data, item{ID: 8041071659142944 + int64(n), Slug: slug, Title: "Klipy moment " + strconv.Itoa(n), Type: "gif", Tags: tags, File: map[string]map[string]gifFile{
			"sm": {"gif": {URL: media + name + "-sm.gif", Width: 220, Height: 220, Size: 1234}},
			"md": {"gif": {URL: media + name + "-md.gif", Width: 320, Height: 320, Size: 3456}},
			"hd": {"gif": {URL: media + name + "-hd.gif", Width: 498, Height: 498, Size: 5678}},
		}})
	}
	writeJSON(w, map[string]any{"result": true, "data": map[string]any{
		"data": data, "current_page": page, "per_page": per, "has_next": page*per < total,
	}})
}

// clipAssets are a clip's MP4 renditions (no size, like the real wire format); no720 drops 720p.
func clipAssets(base string, no720 bool) map[string]map[string]string {
	a := map[string]map[string]string{
		"720p": {"url": base + "-720p.mp4", "width": "1280", "height": "720"},
		"480p": {"url": base + "-480p.mp4", "width": "854", "height": "480"},
		"360p": {"url": base + "-360p.mp4", "width": "640", "height": "360"},
	}
	if no720 {
		delete(a, "720p")
	}
	return a
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

// mp4 is the smallest file the MP4 magic check accepts: one ftyp box (isom, minor 0x200, isom mp41).
var mp4 = []byte("\x00\x00\x00\x18ftypisom\x00\x00\x02\x00isommp41")

// media serves a tiny 2-frame animated GIF whose colour is derived from the file name, or for
// .mp4 names the ftyp fixture. Names starting "broken" answer 500; "html" an HTML page with 200.
// GIFs are 32x32; names starting "wide" are 48x32 and "tall" 32x48 (square crop fixtures).
func media(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	ext := path.Ext(file)
	name := strings.TrimSuffix(file, ext)
	if (ext != ".gif" && ext != ".mp4") || name == "" {
		http.NotFound(w, r)
		return
	}
	if strings.HasPrefix(name, "broken") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if strings.HasPrefix(name, "html") {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprint(w, "<html>error</html>")
		return
	}
	if ext == ".mp4" {
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write(mp4)
		return
	}
	h := fnv.New32a()
	h.Write([]byte(name))
	sum := h.Sum32()
	c1 := color.RGBA{uint8(sum), uint8(sum >> 8), uint8(sum >> 16), 255}
	c2 := color.RGBA{255 - c1.R, 255 - c1.G, 255 - c1.B, 255}
	wd, ht := 32, 32
	if strings.HasPrefix(name, "wide") {
		wd = 48
	} else if strings.HasPrefix(name, "tall") {
		ht = 48
	}
	anim := &gif.GIF{}
	for _, fg := range []color.RGBA{c1, c2} {
		img := image.NewPaletted(image.Rect(0, 0, wd, ht), color.Palette{fg, color.White})
		for y := 12; y < 20; y++ { // a white bar so the frames differ visibly even for similar colours
			for x := 0; x < wd; x++ {
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
