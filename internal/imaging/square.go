// Package imaging is lossless geometry on animated GIFs: frames are cut, never composited or
// re-quantized, so every palette, delay and disposal survives. Stdlib image/gif only.
package imaging

import "image/gif"

// Square returns g center-cropped to a min(w, h) square.
func Square(g *gif.GIF) *gif.GIF { return g } // RED skeleton
