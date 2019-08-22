package jpeg_test

import (
	"bytes"
	"testing"

	"github.com/pixiv/go-libjpeg/jpeg"
)

// https://github.com/pixiv/go-libjpeg/issues/51
func TestIssue51(t *testing.T) {
	data := []byte("\xff\xd8\xff\xdb\x00C\x000000000000000" +
		"00000000000000000000" +
		"00000000000000000000" +
		"00000000000\xff\xc9\x00\v\b00\x000" +
		"\x01\x01\x14\x00\xff\xda\x00\b\x01\x010\x00?\x0000")

	jpeg.Decode(bytes.NewReader(data), &jpeg.DecoderOptions{})
}
