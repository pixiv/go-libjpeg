// Package rgb provides RGB image which implements image.Image interface.
package rgb

import (
	"image"
	"image/color"
)

// Image represent image data which has RGB colors.
// Image is compatible with image.RGBA, but does not have alpha channel to reduce using memory.
type Image struct {
	// Pix holds the image's stream, in R, G, B order.
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewImage allocates and returns RGB image
func NewImage(r image.Rectangle) *Image {
	w, h := r.Dx(), r.Dy()
	return &Image{Pix: make([]uint8, 3*w*h), Stride: 3 * w, Rect: r}
}

// ColorModel returns RGB color model.
func (p *Image) ColorModel() color.Model {
	return ColorModel
}

// Bounds implements image.Image.At
func (p *Image) Bounds() image.Rectangle {
	return p.Rect
}

// At implements image.Image.At
func (p *Image) At(x, y int) color.Color {
	return p.RGBAAt(x, y)
}

// RGBAAt returns the color of the pixel at (x, y) as RGBA.
func (p *Image) RGBAAt(x, y int) color.RGBA {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA{}
	}
	i := (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
	return color.RGBA{p.Pix[i+0], p.Pix[i+1], p.Pix[i+2], 0xFF}
}

// ColorModel is RGB color model instance
var ColorModel = color.ModelFunc(rgbModel)

func rgbModel(c color.Color) color.Color {
	if _, ok := c.(RGB); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}

// RGB color
type RGB struct {
	R, G, B uint8
}

// RGBA implements Color.RGBA
func (c RGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(0xFFFF)
	return
}

// Make sure Image implements image.Image.
// See https://golang.org/doc/effective_go.html#blank_implements.
var _ image.Image = new(Image)
