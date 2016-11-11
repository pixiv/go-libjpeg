package util

import (
	"fmt"
	"image"
	"image/color"
)

func delta(a, b uint8) int {
	d := int(a) - int(b)
	if d < 0 {
		return -d
	}
	return d
}

// MatchColor returns whteher the difference between two colors is smaller than
// the given tolerance. If the two colors a and b assume to the same, it returns
// true.
func MatchColor(a, b color.Color, tolerance int) (matched bool) {
	switch ca := a.(type) {
	case color.CMYK:
		if cb, ok := b.(color.CMYK); ok {
			dC, dM, dY, dK := delta(ca.C, cb.C), delta(ca.M, cb.M), delta(ca.Y, cb.Y), delta(ca.K, cb.K)
			if dC > tolerance || dM > tolerance || dY > tolerance || dK > tolerance {
				return false
			}
			return true
		}
	case color.YCbCr:
		if cb, ok := b.(color.YCbCr); ok {
			dY, dCb, dCr := delta(ca.Y, cb.Y), delta(ca.Cb, cb.Cb), delta(ca.Cr, cb.Cr)
			if dY > tolerance || dCb > tolerance || dCr > tolerance {
				return false
			}
			return true
		}
	case color.NRGBA:
		if cb, ok := b.(color.NRGBA); ok {
			dR, dG, dB, dA := delta(ca.R, cb.R), delta(ca.G, cb.G), delta(ca.B, cb.B), delta(ca.A, cb.A)
			if ca.A == 0 && cb.A == 0 {
				return true
			}
			if dR > tolerance || dG > tolerance || dB > tolerance || dA > tolerance {
				return false
			}
			return true
		}
	}

	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	dr := delta(uint8(ar>>8), uint8(br>>8))
	dg := delta(uint8(ag>>8), uint8(bg>>8))
	db := delta(uint8(ab>>8), uint8(bb>>8))
	da := delta(uint8(aa>>8), uint8(ba>>8))
	if dr > tolerance || dg > tolerance || db > tolerance || da > tolerance {
		return false
	}
	return true
}

// MatchImage matches by pixel-by-pixel. If any one of pixel does not matched,
// it returns an error with image difference (a - b).
func MatchImage(a, b image.Image, tolerance int) (diff image.Image, err error) {
	if a == nil {
		return nil, fmt.Errorf("first image is nil")
	} else if b == nil {
		return nil, fmt.Errorf("second image is nil")
	}
	if a.Bounds().Dx() != b.Bounds().Dx() || a.Bounds().Dy() != b.Bounds().Dy() {
		return nil, fmt.Errorf("unmatched bounds: %v != %v\n", a.Bounds(), b.Bounds())
	}
	rgba := image.NewRGBA(a.Bounds())
	dp := 0
	for x := 0; x < a.Bounds().Dx(); x++ {
		for y := 0; y < a.Bounds().Dy(); y++ {
			aC := a.At(a.Bounds().Min.X+x, a.Bounds().Min.Y+y)
			bC := b.At(b.Bounds().Min.X+x, b.Bounds().Min.Y+y)
			if !MatchColor(aC, bC, tolerance) {
				dp++
				aR, aG, aB, _ := aC.RGBA()
				bR, bG, bB, _ := bC.RGBA()
				dR, dG, dB := delta(uint8(aR>>8), uint8(bR>>8)), delta(uint8(aG>>8), uint8(bG>>8)), delta(uint8(aB>>8), uint8(bB>>8))
				rgba.SetRGBA(x, y, color.RGBA{uint8(dR), uint8(dG), uint8(dB), 255})
			}
		}
	}
	if dp > 0 {
		return rgba, fmt.Errorf("image unmatched")
	}
	return
}
