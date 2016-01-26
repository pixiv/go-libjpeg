// Package jpeg decodes JPEG image to image.YCbCr using libjpeg (or libjpeg-turbo).
package jpeg

//
// Original codes are bollowed from go-thumber.
// Copyright (c) 2014 pixiv Inc. All rights reserved.
//
// See: https://github.com/pixiv/go-thumber
//

/*
#cgo LDFLAGS: -ljpeg
#include <stdlib.h>
#include <stdio.h>
#include <jpeglib.h>

*/
import "C"

// Y/Cb/Cr Planes
const (
	Y  = 0
	Cb = 1
	Cr = 2
)

// DCTMethod is the DCT/IDCT method type.
type DCTMethod C.J_DCT_METHOD

const (
	// DCTISlow is slow but accurate integer algorithm
	DCTISlow DCTMethod = C.JDCT_ISLOW
	// DCTIFast is faster, less accurate integer method
	DCTIFast DCTMethod = C.JDCT_IFAST
	// DCTFloat is floating-point: accurate, fast on fast HW
	DCTFloat DCTMethod = C.JDCT_FLOAT
)
