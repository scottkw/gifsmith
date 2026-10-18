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
	if len(out.Image) != len(in.Image) || !slices.Equal(out.Delay, in.Delay) || !slices.Equal(out.Disposal, in.Disposal) || out.LoopCount != in.LoopCount {
		t.Errorf("frames %d delays %v disposal %v loop %d, want %d %v %v %d",
			len(out.Image), out.Delay, out.Disposal, out.LoopCount, len(in.Image), in.Delay, in.Disposal, in.LoopCount)
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
