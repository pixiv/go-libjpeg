// Package util contains utility code for demosntration of go-libjpeg.
package util

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// SubsampledImages contains example images which have subsampling colors.
var SubsampledImages = []string{
	"checkerboard_444.jpg",
	"checkerboard_440.jpg",
	"checkerboard_422.jpg",
	"checkerboard_420.jpg",
}

// GetExFilePath returns the path of specified example file.
func GetExFilePath(name string) string {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		path := filepath.Join(gopath, "src/github.com/pixiv/go-libjpeg/test/images", name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	panic(fmt.Errorf("%v does not exist in any directory which contains in $GOPATH", name))
}

// GetOutFilePath returns the path of specified out file.
func GetOutFilePath(name string) string {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		path := filepath.Join(gopath, "src/github.com/pixiv/go-libjpeg/test/out")
		if _, err := os.Stat(path); err == nil {
			return filepath.Join(path, name)
		}
	}
	panic(fmt.Errorf("out directory does not exist in any directory which contains in $GOPATH"))
}

// OpenFile opens specified example file
func OpenFile(name string) (io io.Reader) {
	io, err := os.Open(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// ReadFile reads and returns data bytes of specified example file.
func ReadFile(name string) (data []byte) {
	data, err := ioutil.ReadFile(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// CreateFile opens specified example file
func CreateFile(name string) (f *os.File) {
	f, err := os.Create(GetOutFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// WritePNG encodes and writes image into PNG file.
func WritePNG(img image.Image, name string) {
	f, err := os.Create(GetOutFilePath(name))
	if err != nil {
		panic(err)
	}
	b := bufio.NewWriter(f)
	defer func() {
		b.Flush()
		f.Close()
	}()

	if err := png.Encode(b, img); err != nil {
		panic(err)
	}
	return
}
