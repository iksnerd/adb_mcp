package android

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// downscalePNG shrinks a PNG so its largest dimension is at most maxDim, using a
// box-average filter for readable text. If the image already fits (or maxDim is
// non-positive, or decoding fails) the original bytes are returned unchanged.
// The returned width/height describe the bytes actually returned.
func downscalePNG(data []byte, maxDim int) (out []byte, w, h int) {
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return data, 0, 0
	}
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if maxDim <= 0 || (sw <= maxDim && sh <= maxDim) {
		return data, sw, sh
	}

	scale := float64(maxDim) / float64(max(sw, sh))
	dw := max(1, int(float64(sw)*scale))
	dh := max(1, int(float64(sh)*scale))

	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	for dy := 0; dy < dh; dy++ {
		sy0 := b.Min.Y + dy*sh/dh
		sy1 := b.Min.Y + (dy+1)*sh/dh
		if sy1 <= sy0 {
			sy1 = sy0 + 1
		}
		for dx := 0; dx < dw; dx++ {
			sx0 := b.Min.X + dx*sw/dw
			sx1 := b.Min.X + (dx+1)*sw/dw
			if sx1 <= sx0 {
				sx1 = sx0 + 1
			}
			var r, g, bl, a, n uint64
			for sy := sy0; sy < sy1; sy++ {
				for sx := sx0; sx < sx1; sx++ {
					cr, cg, cb, ca := src.At(sx, sy).RGBA()
					r += uint64(cr)
					g += uint64(cg)
					bl += uint64(cb)
					a += uint64(ca)
					n++
				}
			}
			if n == 0 {
				n = 1
			}
			dst.SetNRGBA(dx, dy, color.NRGBA{
				R: uint8((r / n) >> 8),
				G: uint8((g / n) >> 8),
				B: uint8((bl / n) >> 8),
				A: uint8((a / n) >> 8),
			})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return data, sw, sh
	}
	return buf.Bytes(), dw, dh
}
