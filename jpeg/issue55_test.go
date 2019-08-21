package jpeg_test

import (
	"bytes"
	"testing"

	"github.com/pixiv/go-libjpeg/jpeg"
)

var data55 = []byte("\xff\xd8\xff\xdb\x00C\x000000000000000" +
	"00000000000000000000" +
	"00000000000000000000" +
	"00000000000\xff\xc0\x00\x11\b\x00000" +
	"\x03R\"\x00G\x11\x00B\x11\x00\xff\xda\x00\f\x03R\x00G\x00B" +
	"\x00")

// https://github.com/pixiv/go-libjpeg/issues/55
func TestIssue55(t *testing.T) {
	img, err := jpeg.Decode(bytes.NewReader(data55), &jpeg.DecoderOptions{})
	if err != nil {
		return
	}

	var w bytes.Buffer
	err = jpeg.Encode(&w, img, &jpeg.EncoderOptions{})
	if err != nil {
		t.Errorf("encoding after decoding failed: %v", err)
	}
}
