// Package imaging is lossless geometry on animated GIFs: frames are cut, never composited or
// re-quantized, so every palette, delay and disposal survives. Stdlib image/gif only.
package imaging

import (
	"image"
	"image/color"
	"image/gif"
)

// Square returns g center-cropped to a min(w, h) square. Each frame keeps its palette and is cut
// to the window, re-based to the new canvas (EncodeAll refuses out-of-bounds frames). A frame that
// misses the window becomes a 1x1 transparent frame: dropping it, or folding its delay into the
// previous frame, would change when that frame's disposal runs.
func Square(g *gif.GIF) *gif.GIF {
	w, h := g.Config.Width, g.Config.Height
	s := min(w, h)
	win := image.Rect((w-s)/2, (h-s)/2, (w-s)/2+s, (h-s)/2+s)
	out := &gif.GIF{LoopCount: g.LoopCount, BackgroundIndex: g.BackgroundIndex,
		Config: image.Config{ColorModel: g.Config.ColorModel, Width: s, Height: s}}
	for i, fr := range g.Image {
		var p *image.Paletted
		if r := fr.Rect.Intersect(win); r.Empty() {
			p = image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.RGBA{}})
		} else {
			p = fr.SubImage(r).(*image.Paletted)
			p.Rect = p.Rect.Sub(win.Min)
		}
		out.Image = append(out.Image, p)
		out.Delay = append(out.Delay, g.Delay[i])
		if g.Disposal != nil {
			out.Disposal = append(out.Disposal, g.Disposal[i])
		}
	}
	return out
}
