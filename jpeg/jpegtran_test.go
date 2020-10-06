package jpeg_test

import (
	"fmt"
	"image"

	"github.com/pixiv/go-libjpeg/jpeg"
	"github.com/pixiv/go-libjpeg/test/util"

	"bytes"
	"path"
	"strings"
	"testing"
)

func TestJpegTran(t *testing.T) {
	opts := []*jpeg.JpegTranOptions{
		{Progressive: true, Perfect: true},
		{Progressive: true, Perfect: true},
		{Progressive: true, Perfect: true, Transform: jpeg.TransformFlipHorizontal},
		{Progressive: true, Perfect: true, Transform: jpeg.TransformRotate180},
		{Progressive: true, Perfect: true, Transform: jpeg.TransformFlipVertical},
		{Progressive: true, Perfect: true, Transform: jpeg.TransformTranspose},
		{Progressive: true, Perfect: true, Transform: jpeg.TransformRotate90},
		{Progressive: true, Perfect: true, Transform: jpeg.TransformTransverse},
		{Progressive: true, Perfect: true, Transform: jpeg.TransformRotate270},
	}

	expected, err := jpeg.DecodeIntoRGBA(bytes.NewReader(util.ReadFile("lossless_0.jpg")), &jpeg.DecoderOptions{})
	if err != nil {
		t.Fatalf("can't decode expected image: %v", err)
	}

	for i, opt := range opts {
		testJpegTranImage(t, fmt.Sprintf("lossless_%d.jpg", i), expected, opt)
	}
}

func testJpegTranImage(t *testing.T, source string, expected *image.RGBA, opt *jpeg.JpegTranOptions) {
	base := strings.TrimSuffix(path.Base(source), path.Ext(source))
	pngName := strings.TrimSuffix(source, path.Ext(source)) + ".png"
	t.Run(base, func(t *testing.T) {
		src := util.ReadFile(source)

		var buf bytes.Buffer
		if err := jpeg.JpegTran(bytes.NewReader(src), &buf, opt); err != nil {
			t.Fatalf("can't transform image: %v", err)
		}

		actual, err := jpeg.DecodeIntoRGBA(&buf, &jpeg.DecoderOptions{})
		if err != nil {
			t.Fatalf("can't decode created image: %v", err)
		}
		util.WritePNG(actual, pngName)

		ensureSameImage(t, actual, expected)
	})
}

func ensureSameImage(t *testing.T, a *image.RGBA, b *image.RGBA) {
	if a.Rect.Size() != b.Rect.Size() {
		t.Fatalf("image has differ size")
	}
	dy := a.Rect.Dy()
	dx := a.Rect.Dx()
	for y := 0; y < dy; y++ {
		al := a.Pix[y*a.Stride : y*a.Stride+dx*4]
		bl := b.Pix[y*b.Stride : y*b.Stride+dx*4]
		if !bytes.Equal(al, bl) {
			t.Fatalf("image has differ pixels")
		}
	}
}
