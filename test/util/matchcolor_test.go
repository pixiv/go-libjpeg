package util

import (
	"image"
	"image/color"
	"testing"
)

var colorMatches = []struct {
	a, b      color.Color
	tolerance int
	match     bool
}{
	{
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{0, 0, 0, 0},
		0,
		true,
	},
	{
		color.NRGBA{1, 1, 1, 1},
		color.NRGBA{1, 1, 1, 1},
		0,
		true,
	},
	{
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{0, 0, 0, 1},
		0,
		false,
	},
	{
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{1, 0, 0, 0},
		0,
		true,
	},
	{
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{1, 0, 0, 1},
		0,
		false,
	},
	{
		color.NRGBA{0, 1, 0, 1},
		color.NRGBA{1, 0, 0, 1},
		0,
		false,
	},
	{
		color.NRGBA{0, 0, 0, 1},
		color.NRGBA{1, 0, 0, 1},
		0,
		false,
	},
	{
		color.NRGBA{1, 0, 0, 2},
		color.NRGBA{2, 0, 0, 1},
		0,
		false,
	},
	{
		color.NRGBA{0, 0, 0, 127},
		color.NRGBA{2, 0, 0, 127},
		0,
		false,
	},
	{
		color.NRGBA{126, 0, 0, 127},
		color.NRGBA{127, 0, 0, 126},
		0,
		false,
	},
	{
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{0, 0, 0, 1},
		1,
		true,
	},
	{
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{1, 0, 0, 0},
		1,
		true,
	},
	{
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{1, 0, 0, 1},
		1,
		true,
	},
	{
		color.NRGBA{0, 0, 0, 255},
		color.NRGBA{0, 0, 0, 255},
		0,
		true,
	},
	{
		color.NRGBA{127, 0, 0, 255},
		color.NRGBA{126, 0, 0, 255},
		1,
		true,
	},
	{
		color.NRGBA{127, 0, 0, 126},
		color.NRGBA{126, 0, 0, 127},
		1,
		true,
	},
	{
		color.YCbCr{76, 85, 255},
		color.NRGBA{255, 0, 0, 255},
		1,
		true,
	},
}

func TestMatchImage(t *testing.T) {
	for _, x := range colorMatches {
		a := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		b := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		a.Set(0, 0, x.a)
		b.Set(0, 0, x.b)

		if _, err := MatchImage(a, b, x.tolerance); (err == nil) != x.match {
			t.Errorf("MatchImage(a:%v b:%v, tolerance: %v) err:%v but want:%v", x.a, x.b, x.tolerance, err, x.match)
		}
	}
}

func TestMatchColor(t *testing.T) {
	for _, x := range colorMatches {
		if got := MatchColor(x.a, x.b, x.tolerance); x.match != got {
			t.Errorf("MatchColor(a:%v b:%v, tolerance: %v) got:%v but want:%v", x.a, x.b, x.tolerance, got, x.match)
		}
	}
}
