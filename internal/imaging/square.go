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
// misses the window becomes a 1x1 transparent frame with disposal None: dropping it, or folding its
// delay into the previous frame, would change the timing, and its own disposal touched only pixels
// outside the window, while Background on the placeholder would clear the visible pixel (0,0).
func Square(g *gif.GIF) *gif.GIF {
	w, h := g.Config.Width, g.Config.Height
	s := min(w, h)
	win := image.Rect((w-s)/2, (h-s)/2, (w-s)/2+s, (h-s)/2+s)
	out := &gif.GIF{LoopCount: g.LoopCount, BackgroundIndex: g.BackgroundIndex,
		Config: image.Config{ColorModel: g.Config.ColorModel, Width: s, Height: s}}
	for i, fr := range g.Image {
		var p *image.Paletted
		disposal := byte(0)
		if g.Disposal != nil {
			disposal = g.Disposal[i]
		}
		if r := fr.Rect.Intersect(win); r.Empty() {
			p, disposal = image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.RGBA{}}), gif.DisposalNone
		} else {
			p = fr.SubImage(r).(*image.Paletted)
			p.Rect = p.Rect.Sub(win.Min)
		}
		out.Image = append(out.Image, p)
		out.Delay = append(out.Delay, g.Delay[i])
		out.Disposal = append(out.Disposal, disposal)
	}
	return out
}
