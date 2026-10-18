package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"slices"
	"testing"
)

var pal = color.Palette{
	color.RGBA{0, 0, 0, 0}, // transparent
	color.RGBA{255, 0, 0, 255},
	color.RGBA{0, 255, 0, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{255, 255, 255, 255},
}

// frame is an opaque patterned frame over r; seed varies the pattern per frame.
func frame(r image.Rectangle, seed int) *image.Paletted {
	p := image.NewPaletted(r, pal)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			p.SetColorIndex(x, y, uint8(1+(x/3+y/2+seed)%4))
		}
	}
	return p
}

// roundTrip encodes and decodes g, as the real pipeline does (DecodeAll fills Config and Disposal).
func roundTrip(t *testing.T, g *gif.GIF) *gif.GIF {
	t.Helper()
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatalf("EncodeAll: %v", err)
	}
	out, err := gif.DecodeAll(&buf)
	if err != nil {
		t.Fatalf("DecodeAll: %v", err)
	}
	return out
}

func anim(w, h int, frames []*image.Paletted, delays []int, disposal []byte) *gif.GIF {
	return &gif.GIF{Image: frames, Delay: delays, Disposal: disposal, LoopCount: 3, Config: image.Config{Width: w, Height: h}}
}

// checkSquare crops in, round-trips the result and checks size, timing, loop and frame bounds.
func checkSquare(t *testing.T, in *gif.GIF, side int) *gif.GIF {
	t.Helper()
	out := roundTrip(t, Square(in))
	if out.Config.Width != side || out.Config.Height != side {
		t.Fatalf("canvas %dx%d, want %dx%d", out.Config.Width, out.Config.Height, side, side)
	}
	// A frame outside the window is a placeholder whose disposal is None (see Square).
	w, h := in.Config.Width, in.Config.Height
	win := image.Rect((w-side)/2, (h-side)/2, (w-side)/2+side, (h-side)/2+side)
	disposal := slices.Clone(in.Disposal)
	for i, fr := range in.Image {
		if fr.Rect.Intersect(win).Empty() {
			disposal[i] = gif.DisposalNone
		}
	}
	if len(out.Image) != len(in.Image) || !slices.Equal(out.Delay, in.Delay) || !slices.Equal(out.Disposal, disposal) || out.LoopCount != in.LoopCount {
		t.Errorf("frames %d delays %v disposal %v loop %d, want %d %v %v %d",
			len(out.Image), out.Delay, out.Disposal, out.LoopCount, len(in.Image), in.Delay, disposal, in.LoopCount)
	}
	for i, fr := range out.Image {
		if !fr.Bounds().In(image.Rect(0, 0, side, side)) {
			t.Errorf("frame %d bounds %v outside the %dx%d canvas", i, fr.Bounds(), side, side)
		}
	}
	return out
}

func TestSquareWide(t *testing.T) {
	r := image.Rect(0, 0, 120, 60)
	in := roundTrip(t, anim(120, 60, []*image.Paletted{frame(r, 0), frame(r, 1)}, []int{10, 20}, []byte{gif.DisposalNone, gif.DisposalBackground}))
	checkSquare(t, in, 60)
}

func TestSquareTall(t *testing.T) {
	r := image.Rect(0, 0, 32, 48)
	in := roundTrip(t, anim(32, 48, []*image.Paletted{frame(r, 0), frame(r, 2)}, []int{5, 5}, []byte{gif.DisposalNone, gif.DisposalNone}))
	checkSquare(t, in, 32)
}

func TestSquareAlreadySquareIsIdentical(t *testing.T) {
	r := image.Rect(0, 0, 40, 40)
	in := roundTrip(t, anim(40, 40, []*image.Paletted{frame(r, 0), frame(r, 3)}, []int{7, 9}, []byte{gif.DisposalNone, gif.DisposalPrevious}))
	out := checkSquare(t, in, 40)
	for i := range in.Image {
		for y := 0; y < 40; y++ {
			for x := 0; x < 40; x++ {
				if a, b := in.Image[i].At(x, y), out.Image[i].At(x, y); a != b {
					t.Fatalf("frame %d pixel (%d,%d) = %v, want %v", i, x, y, b, a)
				}
			}
		}
	}
}

// render composites frames 0..upto onto an RGBA canvas the way a GIF viewer does, applying each
// earlier frame's disposal (None keeps it, Background clears its rect, Previous restores).
func render(g *gif.GIF, upto int) *image.RGBA {
	b := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(b)
	for i := 0; i <= upto; i++ {
		fr := g.Image[i]
		var saved *image.RGBA
		if g.Disposal[i] == gif.DisposalPrevious {
			saved = image.NewRGBA(b)
			copy(saved.Pix, canvas.Pix)
		}
		for y := fr.Rect.Min.Y; y < fr.Rect.Max.Y; y++ {
			for x := fr.Rect.Min.X; x < fr.Rect.Max.X; x++ {
				if c := fr.Palette[fr.ColorIndexAt(x, y)].(color.RGBA); c.A != 0 {
					canvas.SetRGBA(x, y, c)
				}
			}
		}
		if i == upto {
			break
		}
		switch g.Disposal[i] {
		case gif.DisposalBackground:
			for y := fr.Rect.Min.Y; y < fr.Rect.Max.Y; y++ {
				for x := fr.Rect.Min.X; x < fr.Rect.Max.X; x++ {
					canvas.SetRGBA(x, y, color.RGBA{})
				}
			}
		case gif.DisposalPrevious:
			canvas = saved
		}
	}
	return canvas
}

// transparent is a frame over r that is transparent except a patterned square in its middle.
func transparent(r image.Rectangle, seed int) *image.Paletted {
	p := frame(r, seed)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if (x+y+seed)%5 == 0 {
				p.SetColorIndex(x, y, 0)
			}
		}
	}
	return p
}

// TestSquareGolden: decode -> Square -> EncodeAll -> DecodeAll keeps frames, delays and loop, and
// every pixel of every rendered output frame equals the source render inside the crop window.
func TestSquareGolden(t *testing.T) {
	full := image.Rect(0, 0, 120, 60)
	fixtures := map[string]*gif.GIF{
		"wide every disposal": anim(120, 60, []*image.Paletted{
			frame(full, 0),
			transparent(image.Rect(20, 10, 100, 50), 1),
			transparent(image.Rect(40, 0, 80, 30), 2),
			transparent(image.Rect(0, 0, 50, 60), 3), // partially overlaps the window
			frame(image.Rect(0, 0, 25, 60), 1),       // entirely left of the 30..90 window
			transparent(image.Rect(85, 20, 120, 60), 4),
		}, []int{10, 20, 30, 40, 50, 60}, []byte{gif.DisposalNone, gif.DisposalBackground, gif.DisposalPrevious, gif.DisposalNone, gif.DisposalBackground, gif.DisposalNone}),
		"tall": anim(60, 120, []*image.Paletted{
			frame(image.Rect(0, 0, 60, 120), 0),
			transparent(image.Rect(0, 100, 60, 120), 2), // entirely below the 30..90 window
			transparent(image.Rect(10, 20, 50, 100), 3),
		}, []int{5, 6, 7}, []byte{gif.DisposalNone, gif.DisposalPrevious, gif.DisposalBackground}),
		"square": anim(40, 40, []*image.Paletted{frame(image.Rect(0, 0, 40, 40), 0), transparent(image.Rect(5, 5, 30, 30), 1)},
			[]int{8, 9}, []byte{gif.DisposalNone, gif.DisposalNone}),
	}
	for name, src := range fixtures {
		in := roundTrip(t, src)
		w, h := in.Config.Width, in.Config.Height
		s := min(w, h)
		win := image.Rect((w-s)/2, (h-s)/2, (w-s)/2+s, (h-s)/2+s)
		out := checkSquare(t, in, s)
		mismatches := 0
		for i := range in.Image {
			a, b := render(in, i), render(out, i)
			for y := 0; y < s; y++ {
				for x := 0; x < s; x++ {
					if a.RGBAAt(x+win.Min.X, y+win.Min.Y) != b.RGBAAt(x, y) {
						mismatches++
					}
				}
			}
		}
		t.Logf("%s: %d frames, %d mismatches", name, len(in.Image), mismatches)
		if mismatches != 0 {
			t.Errorf("%s: %d pixel mismatches inside the window", name, mismatches)
		}
	}
}

// A frame entirely outside the window becomes 1x1 transparent and keeps its own delay. Its
// disposal becomes None: Background would clear the placeholder's pixel at (0,0), which the
// original frame never covered.
func TestSquareOutsideFramePlaceholder(t *testing.T) {
	in := roundTrip(t, anim(120, 60, []*image.Paletted{frame(image.Rect(0, 0, 120, 60), 0), frame(image.Rect(0, 0, 25, 60), 1)},
		[]int{10, 70}, []byte{gif.DisposalNone, gif.DisposalBackground}))
	sq := Square(in)
	p := sq.Image[1]
	if p.Bounds() != image.Rect(0, 0, 1, 1) || len(p.Palette) != 1 {
		t.Fatalf("outside frame = %v with %d colours, want a 1x1 one-entry frame", p.Bounds(), len(p.Palette))
	}
	if _, _, _, a := p.Palette[0].RGBA(); a != 0 {
		t.Error("placeholder colour is not transparent")
	}
	if !slices.Equal(sq.Delay, []int{10, 70}) {
		t.Errorf("delays = %v, want [10 70] (not folded into the previous frame)", sq.Delay)
	}
	if sq.Disposal[1] != gif.DisposalNone {
		t.Errorf("placeholder disposal = %d, want DisposalNone", sq.Disposal[1])
	}
}
