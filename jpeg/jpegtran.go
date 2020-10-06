package jpeg

/*
#include <stdio.h>
#include <stdlib.h>
#include "jpeglib.h"
#include "transupp.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
)

type Transform int

const (
	TransformNone Transform = iota
	TransformFlipHorizontal
	TransformFlipVertical
	TransformTranspose
	TransformTransverse
	TransformRotate90
	TransformRotate180
	TransformRotate270
)

type JpegTranOptions struct {
	// Create progressive JPEG file
	Progressive bool
	// Fail if there is non-transformable edge blocks
	Perfect   bool
	Transform Transform
}

func NewJpegTranOptions() *JpegTranOptions {
	return &JpegTranOptions{
		Progressive: true,
		Perfect:     true,
	}
}

//
// Based on https://github.com/cloudflare/jpegtran/blob/master/jpegtran.c implementation.
//
func JpegTran(r io.Reader, w io.Writer, options *JpegTranOptions) error {
	if options == nil {
		options = NewJpegTranOptions()
	}

	srcInfo := newDecompress(r)
	if srcInfo == nil {
		return errors.New("allocation failed")
	}
	defer destroyDecompress(srcInfo)

	dstInfo, err := newCompress(w)
	if err != nil {
		return err
	}
	defer destroyCompress(dstInfo)

	err = readHeader(srcInfo)
	if err != nil {
		return err
	}

	var transformOption C.jpeg_transform_info
	if options.Perfect {
		transformOption.perfect = 1
	}

	switch options.Transform {
	case TransformNone:
		transformOption.transform = C.JXFORM_NONE
	case TransformFlipHorizontal:
		transformOption.transform = C.JXFORM_FLIP_H
	case TransformFlipVertical:
		transformOption.transform = C.JXFORM_FLIP_V
	case TransformTranspose:
		transformOption.transform = C.JXFORM_TRANSPOSE
	case TransformTransverse:
		transformOption.transform = C.JXFORM_TRANSVERSE
	case TransformRotate90:
		transformOption.transform = C.JXFORM_ROT_90
	case TransformRotate180:
		transformOption.transform = C.JXFORM_ROT_180
	case TransformRotate270:
		transformOption.transform = C.JXFORM_ROT_270
	default:
		return errors.New(fmt.Sprintf("unknown transform: %v", options.Transform))
	}

	//transformOption.transform = C.JXFORM_FLIP_H
	if C.jtransform_request_workspace(srcInfo, &transformOption) == 0 {
		return errors.New("transformation is not perfect")
	}

	srcCoefArrays, err := readCoefficients(srcInfo)
	if err != nil {
		return err
	}

	C.jpeg_copy_critical_parameters(srcInfo, dstInfo)

	if options.Progressive {
		C.jpeg_simple_progression(dstInfo)
	}

	dstCoefArrays := C.jtransform_adjust_parameters(srcInfo, dstInfo, srcCoefArrays, &transformOption)

	C.jpeg_write_coefficients(dstInfo, dstCoefArrays)

	C.jtransform_execute_transformation(srcInfo, dstInfo,
		srcCoefArrays,
		&transformOption)

	C.jpeg_finish_compress(dstInfo)

	return nil
}
