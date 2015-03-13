package jpeg_test

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	nativeJPEG "image/jpeg"
	"os"
	"testing"

	"github.com/pixiv/go-libjpeg/jpeg"
	"github.com/pixiv/go-libjpeg/test/util"
)

var naturalImageFiles = []string{
	"cosmos.jpg",
	"kinkaku.jpg",
}

var subsampledImageFiles = []string{
	"checkerboard_444.jpg",
	"checkerboard_440.jpg",
	"checkerboard_422.jpg",
	"checkerboard_420.jpg",
}

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, file := range naturalImageFiles {
			io := util.OpenFile(file)
			img, err := jpeg.Decode(io, &jpeg.DecoderOptions{})
			if img == nil {
				b.Error("Got nil")
			}
			if err != nil {
				b.Errorf("Got Error: %v", err)
			}
		}
	}
}

func BenchmarkDecodeIntoRGB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, file := range naturalImageFiles {
			io := util.OpenFile(file)
			img, err := jpeg.DecodeIntoRGB(io, &jpeg.DecoderOptions{})
			if img == nil {
				b.Error("Got nil")
			}
			if err != nil {
				b.Errorf("Got Error: %v", err)
			}
		}
	}
}

func BenchmarkDecodeWithNativeJPEG(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, file := range naturalImageFiles {
			io := util.OpenFile(file)
			img, err := nativeJPEG.Decode(io)
			if img == nil {
				b.Error("Got nil")
			}
			if err != nil {
				b.Errorf("Got Error: %v", err)
			}
		}
	}
}

func TestDecode(t *testing.T) {
	for _, file := range naturalImageFiles {
		io := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.Decode(io, &jpeg.DecoderOptions{})
		if err != nil {
			t.Errorf("Got Error: %v", err)
		}

		util.WritePNG(img, fmt.Sprintf("TestDecode_%s.png", file))
	}
}

func TestDecodeScaled(t *testing.T) {
	for _, file := range naturalImageFiles {
		io := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.Decode(io, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, 100, 100)})
		if err != nil {
			t.Errorf("Got Error: %v", err)
		}
		if got := img.Bounds().Dx(); got != 256 {
			t.Errorf("Wrong scaled width: %v, expect: 128 (=1024/8)", got)
		}
		if got := img.Bounds().Dy(); got != 192 {
			t.Errorf("Wrong scaled height: %v, expect: 192 (=768/8)", got)
		}

		util.WritePNG(img, fmt.Sprintf("TestDecodeScaled_%s.png", file))
	}
}

func TestDecodeIntoRGBA(t *testing.T) {
	if jpeg.SupportRGBA() != true {
		t.Skipf("This build is not support DecodeIntoRGBA.")
		return
	}
	for _, file := range naturalImageFiles {
		io := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.DecodeIntoRGBA(io, &jpeg.DecoderOptions{})
		if err != nil {
			t.Errorf("Got Error: %v", err)
			continue
		}

		util.WritePNG(img, fmt.Sprintf("TestDecodeIntoRGBA_%s.png", file))
	}
}

func TestDecodeScaledIntoRGBA(t *testing.T) {
	if jpeg.SupportRGBA() != true {
		t.Skipf("This build is not support DecodeIntoRGBA.")
		return
	}
	for _, file := range naturalImageFiles {
		io := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.DecodeIntoRGBA(io, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, 100, 100)})
		if err != nil {
			t.Errorf("Got Error: %v", err)
			continue
		}
		if got := img.Bounds().Dx(); got != 256 {
			t.Errorf("Wrong scaled width: %v, expect: 128 (=1024/8)", got)
		}
		if got := img.Bounds().Dy(); got != 192 {
			t.Errorf("Wrong scaled height: %v, expect: 192 (=768/8)", got)
		}

		util.WritePNG(img, fmt.Sprintf("TestDecodeIntoRGBA_%s.png", file))
	}
}

func TestDecodeScaledIntoRGB(t *testing.T) {
	for _, file := range naturalImageFiles {
		io := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.DecodeIntoRGB(io, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, 100, 100)})
		if err != nil {
			t.Errorf("Got Error: %v", err)
		}
		if got := img.Bounds().Dx(); got != 256 {
			t.Errorf("Wrong scaled width: %v, expect: 128 (=1024/8)", got)
		}
		if got := img.Bounds().Dy(); got != 192 {
			t.Errorf("Wrong scaled height: %v, expect: 192 (=768/8)", got)
		}

		util.WritePNG(img, fmt.Sprintf("TestDecodeIntoRGB_%s.png", file))
	}
}

func TestDecodeSubsampledImage(t *testing.T) {
	for _, file := range subsampledImageFiles {
		io := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.Decode(io, &jpeg.DecoderOptions{})
		if err != nil {
			t.Errorf("Got Error: %v", err)
		}

		util.WritePNG(img, fmt.Sprintf("TestDecodeSubsampledImage_%s.png", file))
	}
}

func TestDecodeAndEncode(t *testing.T) {
	for _, file := range naturalImageFiles {
		io := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.Decode(io, &jpeg.DecoderOptions{})
		if err != nil {
			t.Errorf("Decode returns error: %v", err)
		}

		// Create output file
		f, err := os.Create(util.GetOutFilePath(fmt.Sprintf("TestDecodeAndEncode_%s", file)))
		if err != nil {
			panic(err)
		}
		w := bufio.NewWriter(f)
		defer func() {
			w.Flush()
			f.Close()
		}()

		if err := jpeg.Encode(w, img, &jpeg.EncoderOptions{Quality: 90}); err != nil {
			t.Errorf("Encode returns error: %v", err)
		}
	}
}

func TestDecodeAndEncodeSubsampledImages(t *testing.T) {
	for _, file := range subsampledImageFiles {
		r := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		img, err := jpeg.Decode(r, &jpeg.DecoderOptions{})
		if err != nil {
			t.Errorf("Decode returns error: %v", err)
		}

		// Create output file
		f, err := os.Create(util.GetOutFilePath(fmt.Sprintf("TestDecodeAndEncodeSubsampledImages_%s", file)))
		if err != nil {
			panic(err)
		}
		w := bufio.NewWriter(f)
		defer func() {
			w.Flush()
			f.Close()
		}()

		if err := jpeg.Encode(w, img, &jpeg.EncoderOptions{Quality: 90}); err != nil {
			t.Errorf("Encode returns error: %v", err)
		}
	}
}

func TestDecodeConfig(t *testing.T) {
	for _, file := range naturalImageFiles {
		r := util.OpenFile(file)
		fmt.Printf(" - test: %s\n", file)

		config, err := jpeg.DecodeConfig(r)
		if err != nil {
			t.Errorf("Got error: %v", err)
		}

		if got := config.ColorModel; got != color.YCbCrModel {
			t.Errorf("got wrong ColorModel: %v, expect: color.YCbCrModel", got)
		}
		if got := config.Width; got != 1024 {
			t.Errorf("got wrong width: %d, expect: 1024", got)
		}
		if got := config.Height; got != 768 {
			t.Errorf("got wrong height: %d, expect: 768", got)
		}
	}
}

func TestNewYCbCrAlignedWithLandscape(t *testing.T) {
	got := jpeg.NewYCbCrAligned(image.Rect(0, 0, 125, 25), image.YCbCrSubsampleRatio444)

	if len(got.Y) != 6912 {
		t.Errorf("wrong array size Y: %d, expect: 6912", len(got.Y))
	}
	if len(got.Cb) != 6912 {
		t.Errorf("wrong array size Cb: %d, expect: 6912", len(got.Cb))
	}
	if len(got.Cr) != 6912 {
		t.Errorf("wrong array size Cr: %d, expect: 6912", len(got.Cr))
	}
	if got.YStride != 144 {
		t.Errorf("got wrong YStride: %d, expect: 128", got.YStride)
	}
	if got.CStride != 144 {
		t.Errorf("got wrong CStride: %d, expect: 128", got.CStride)
	}
}

func TestNewYCbCrAlignedWithPortrait(t *testing.T) {
	got := jpeg.NewYCbCrAligned(image.Rect(0, 0, 25, 125), image.YCbCrSubsampleRatio444)

	if len(got.Y) != 6912 {
		t.Errorf("wrong array size Y: %d, expect: 6912", len(got.Y))
	}
	if len(got.Cb) != 6912 {
		t.Errorf("wrong array size Cb: %d, expect: 6912", len(got.Cb))
	}
	if len(got.Cr) != 6912 {
		t.Errorf("wrong array size Cr: %d, expect: 6912", len(got.Cr))
	}
	if got.YStride != 48 {
		t.Errorf("got wrong YStride: %d, expect: 128", got.YStride)
	}
	if got.CStride != 48 {
		t.Errorf("got wrong CStride: %d, expect: 128", got.CStride)
	}
}
