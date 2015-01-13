package rgb_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/pixiv/go-libjpeg/rgb"
)

func TestImageInterface(t *testing.T) {
	rect := image.Rect(0, 0, 100, 100)
	img := rgb.NewImage(rect)

	if got := img.ColorModel(); got != rgb.ColorModel {
		t.Errorf("ColorModel() should return rgb.ColorModel, got: %v", got)
	}

	if got := img.Bounds(); got != rect {
		t.Errorf("Bounds() should return %v, got: %v", rect, got)
	}

	black := color.RGBA{0x00, 0x00, 0x00, 0xFF}
	if got := img.At(0, 0); got != black {
		t.Errorf("At(0, 0) should return %v, got: %v", black, got)
	}

	blank := color.RGBA{}
	if got := img.At(-1, -1); got != blank {
		t.Errorf("At(0, 0) should return %v, got: %v", blank, got)
	}
}

func TestConvertFromRGBA(t *testing.T) {
	rgba := color.RGBA{0x11, 0x22, 0x33, 0xFF}
	expect := rgb.RGB{0x11, 0x22, 0x33}
	if got := rgb.ColorModel.Convert(rgba); got != expect {
		t.Errorf("got: %v, expect: %v", got, expect)
	}
}

func TestConvertFromRGB(t *testing.T) {
	c := rgb.RGB{0x11, 0x22, 0x33}
	if got := rgb.ColorModel.Convert(c); got != c {
		t.Errorf("got: %v, expect: %v", got, c)
	}
}

func TestColorRGBA(t *testing.T) {
	c := rgb.RGB{0x11, 0x22, 0x33}
	r, g, b, a := uint32(0x1111), uint32(0x2222), uint32(0x3333), uint32(0xFFFF)

	gotR, gotG, gotB, gotA := c.RGBA()
	if gotR != r {
		t.Errorf("got R: %v, expect R: %v", gotR, r)
	}
	if gotG != g {
		t.Errorf("got G: %v, expect G: %v", gotG, g)
	}
	if gotB != b {
		t.Errorf("got B: %v, expect B: %v", gotB, b)
	}
	if gotA != a {
		t.Errorf("got A: %v, expect A: %v", gotA, a)
	}
}
