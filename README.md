go-libjpeg
==========

[![GoDoc](https://godoc.org/github.com/pixiv/go-libjpeg/jpeg?status.svg)](https://godoc.org/github.com/pixiv/go-libjpeg/jpeg)
[![Build Status](https://travis-ci.org/pixiv/go-libjpeg.svg?branch=master)](https://travis-ci.org/pixiv/go-libjpeg)

An implementation of Go binding for LibJpeg (preferably libjpeg-turbo).

The core codes are picked from [go-thumber](http://github.com/pixiv/go-thumber)
and rewritten to compatible with image.Image interface.

## Usage

```
import "github.com/pixiv/go-libjpeg/jpeg"

func main() {
    // Decoding JPEG into image.Image
    io, err := os.Open("in.jpg")
    if err != nil {
        log.Fatal(err)
    }
    img, err := jpeg.Decode(io, &jpeg.DecoderOptions{})
    if err != nil {
        log.Fatalf("Decode returns error: %v\n", err)
    }

    // Encode JPEG
    f, err := os.Create("out.jpg")
    if err != nil {
        panic(err)
    }
    w := bufio.NewWriter(f)
    if err := jpeg.Encode(w, img, &jpeg.EncoderOptions{Quality: 90}); err != nil {
        log.Printf("Encode returns error: %v\n", err)
        return
    }
    w.Flush()
    f.Close()
}
```

See [test code](./jpeg/jpeg_test.go) to read full features.

## Features

- Raw JPEG decoding in YCbCr color.
- Decoding with color conversion into RGB/RGBA (RGBA conversion is only supported with libjpeg-turbo).
- Scaled decoding.
- Encoding from some color models (YCbCr, RGB and RGBA).

## Benchmark

```
$ go test -bench . -benchtime 10s
...
BenchmarkDecode                     1000          26345730 ns/op
BenchmarkDecodeIntoRGB               500          30886383 ns/op
BenchmarkDecodeWithNativeJPEG        300          49815928 ns/op
...
```

With libjpeg-turbo:
```
BenchmarkDecode                     2000           9557646 ns/op
BenchmarkDecodeIntoRGB              1000          12676414 ns/op
BenchmarkDecodeWithNativeJPEG        300          45836153 ns/op
```

go-libjpeg is about 1.9x faster than image/jpeg. 
With libjpeg-turbo, it can make more faster (about 4.8x faster than image/jpeg).

### Dependencies

* Go 1.6 or later.
* libjpeg (preferably libjpeg-turbo)

    DecodeIntoRGBA can only work if go-libjpeg is built with libjpeg-turbo.
    Because DecdeIntoRGBA uses `JCS_ALPHA_EXTENSIONS`. You can use
    DecodeIntoRGB and convert to image.RGBA if can not use libjpeg-turbo.

## License

Copyright (c) 2014 pixiv Inc. All rights reserved.

See [LICENSE](./LICENSE).
