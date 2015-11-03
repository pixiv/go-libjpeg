package jpeg

/*
#include <stdio.h>
#include <stdlib.h>
#include "jpeglib.h"

void error_panic(j_common_ptr cinfo);

static struct jpeg_compress_struct *new_compress() {
	struct jpeg_compress_struct *cinfo = (struct jpeg_compress_struct *) malloc(sizeof(struct jpeg_compress_struct));
	struct jpeg_error_mgr *jerr = (struct jpeg_error_mgr *)malloc(sizeof(struct jpeg_error_mgr));

  jpeg_std_error(jerr);
  jerr->error_exit = (void *)error_panic;
	jpeg_create_compress(cinfo);
  cinfo->err = jerr;

	return cinfo;
}

static void destroy_compress(struct jpeg_compress_struct *cinfo) {
	free(cinfo->err);
	jpeg_destroy_compress(cinfo);
}

static int DCT_v_scaled_size(j_decompress_ptr cinfo, int component) {
#if JPEG_LIB_VERSION >= 70
	return cinfo->comp_info[component].DCT_v_scaled_size;
#else
	return cinfo->comp_info[component].DCT_scaled_size;
#endif
}

*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"io"
	"unsafe"
)

// EncoderOptions specifies which settings to use during Compression.
type EncoderOptions struct {
	Quality        int
	OptimizeCoding bool
	DCTMethod      DCTMethod
}

// Encode encodes src image and writes into w as JPEG format data.
func Encode(w io.Writer, src image.Image, opt *EncoderOptions) (err error) {
	// Recover panic
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("JPEG error: %v", r)
			}
		}
	}()

	var cinfo *C.struct_jpeg_compress_struct = C.new_compress()
	defer C.destroy_compress(cinfo)

	dstManager := makeDestinationManager(w, cinfo)
	defer C.free(unsafe.Pointer(dstManager))

	switch s := src.(type) {
	case *image.YCbCr:
		err = encodeYCbCr(cinfo, s, opt)
	case *image.Gray:
		err = encodeGray(cinfo, s, opt)
	default:
		return errors.New("unsupported image type")
	}

	return
}

// encode image.YCbCr
func encodeYCbCr(cinfo *C.struct_jpeg_compress_struct, src *image.YCbCr, p *EncoderOptions) (err error) {
	// Set up compression parameters
	cinfo.image_width = C.JDIMENSION(src.Bounds().Dx())
	cinfo.image_height = C.JDIMENSION(src.Bounds().Dy())
	cinfo.input_components = 3
	cinfo.in_color_space = C.JCS_YCbCr

	C.jpeg_set_defaults(cinfo)
	setupEncoderOptions(cinfo, p)

	compInfo := (*[3]C.jpeg_component_info)(unsafe.Pointer(cinfo.comp_info))
	colorVDiv := 1
	switch src.SubsampleRatio {
	case image.YCbCrSubsampleRatio444:
		// 1x1,1x1,1x1
		compInfo[Y].h_samp_factor, compInfo[Y].v_samp_factor = 1, 1
		compInfo[Cb].h_samp_factor, compInfo[Cb].v_samp_factor = 1, 1
		compInfo[Cr].h_samp_factor, compInfo[Cr].v_samp_factor = 1, 1
	case image.YCbCrSubsampleRatio440:
		// 1x2,1x1,1x1
		compInfo[Y].h_samp_factor, compInfo[Y].v_samp_factor = 1, 2
		compInfo[Cb].h_samp_factor, compInfo[Cb].v_samp_factor = 1, 1
		compInfo[Cr].h_samp_factor, compInfo[Cr].v_samp_factor = 1, 1
		colorVDiv = 2
	case image.YCbCrSubsampleRatio422:
		// 2x1,1x1,1x1
		compInfo[Y].h_samp_factor, compInfo[Y].v_samp_factor = 2, 1
		compInfo[Cb].h_samp_factor, compInfo[Cb].v_samp_factor = 1, 1
		compInfo[Cr].h_samp_factor, compInfo[Cr].v_samp_factor = 1, 1
	case image.YCbCrSubsampleRatio420:
		// 2x2,1x1,1x1
		compInfo[Y].h_samp_factor, compInfo[Y].v_samp_factor = 2, 2
		compInfo[Cb].h_samp_factor, compInfo[Cb].v_samp_factor = 1, 1
		compInfo[Cr].h_samp_factor, compInfo[Cr].v_samp_factor = 1, 1
		colorVDiv = 2
	}

	// libjpeg raw data in is in planar format, which avoids unnecessary
	// planar->packed->planar conversions.
	cinfo.raw_data_in = C.TRUE

	// Start compression
	C.jpeg_start_compress(cinfo, C.TRUE)

	// Allocate JSAMPIMAGE to hold pointers to one iMCU worth of image data
	// this is a safe overestimate; we use the return value from
	// jpeg_read_raw_data to figure out what is the actual iMCU row count.
	var yRowPtr [AlignSize]C.JSAMPROW
	var cbRowPtr [AlignSize]C.JSAMPROW
	var crRowPtr [AlignSize]C.JSAMPROW
	yCbCrPtr := [3]C.JSAMPARRAY{
		C.JSAMPARRAY(unsafe.Pointer(&yRowPtr[0])),
		C.JSAMPARRAY(unsafe.Pointer(&cbRowPtr[0])),
		C.JSAMPARRAY(unsafe.Pointer(&crRowPtr[0])),
	}

	var rows C.JDIMENSION
	for rows = 0; rows < cinfo.image_height; {
		// First fill in the pointers into the plane data buffers
		for j := 0; j < int(C.DCTSIZE*compInfo[Y].v_samp_factor); j++ {
			yRowPtr[j] = C.JSAMPROW(unsafe.Pointer(&src.Y[src.YStride*(int(rows)+j)]))
		}
		for j := 0; j < int(C.DCTSIZE*compInfo[Cb].v_samp_factor); j++ {
			cbRowPtr[j] = C.JSAMPROW(unsafe.Pointer(&src.Cb[src.CStride*(int(rows)/colorVDiv+j)]))
			crRowPtr[j] = C.JSAMPROW(unsafe.Pointer(&src.Cr[src.CStride*(int(rows)/colorVDiv+j)]))
		}
		// Get the data
		rows += C.jpeg_write_raw_data(cinfo, C.JSAMPIMAGE(unsafe.Pointer(&yCbCrPtr[0])), C.JDIMENSION(C.DCTSIZE*compInfo[0].v_samp_factor))
	}

	C.jpeg_finish_compress(cinfo)
	return
}

// encode image.Gray
func encodeGray(cinfo *C.struct_jpeg_compress_struct, src *image.Gray, p *EncoderOptions) (err error) {
	// Set up compression parameters
	cinfo.image_width = C.JDIMENSION(src.Bounds().Dx())
	cinfo.image_height = C.JDIMENSION(src.Bounds().Dy())
	cinfo.input_components = 1
	cinfo.in_color_space = C.JCS_GRAYSCALE

	C.jpeg_set_defaults(cinfo)
	setupEncoderOptions(cinfo, p)

	compInfo := (*C.jpeg_component_info)(unsafe.Pointer(cinfo.comp_info))
	compInfo.h_samp_factor, compInfo.v_samp_factor = 1, 1

	// libjpeg raw data in is in planar format, which avoids unnecessary
	// planar->packed->planar conversions.
	cinfo.raw_data_in = C.TRUE

	// Start compression
	C.jpeg_start_compress(cinfo, C.TRUE)

	// Allocate JSAMPIMAGE to hold pointers to one iMCU worth of image data
	// this is a safe overestimate; we use the return value from
	// jpeg_read_raw_data to figure out what is the actual iMCU row count.
	var rowPtr [AlignSize]C.JSAMPROW
	arrayPtr := [1]C.JSAMPARRAY{
		C.JSAMPARRAY(unsafe.Pointer(&rowPtr[0])),
	}

	var rows C.JDIMENSION
	for rows = 0; rows < cinfo.image_height; {
		// First fill in the pointers into the plane data buffers
		for j := 0; j < int(C.DCTSIZE*compInfo.v_samp_factor); j++ {
			rowPtr[j] = C.JSAMPROW(unsafe.Pointer(&src.Pix[src.Stride*(int(rows)+j)]))
		}
		// Get the data
		rows += C.jpeg_write_raw_data(cinfo, C.JSAMPIMAGE(unsafe.Pointer(&arrayPtr[0])), C.JDIMENSION(C.DCTSIZE*compInfo.v_samp_factor))
	}

	C.jpeg_finish_compress(cinfo)
	return
}

func setupEncoderOptions(cinfo *C.struct_jpeg_compress_struct, opt *EncoderOptions) {
	C.jpeg_set_quality(cinfo, C.int(opt.Quality), C.TRUE)
	if opt.OptimizeCoding {
		cinfo.optimize_coding = C.TRUE
	} else {
		cinfo.optimize_coding = C.FALSE
	}
	cinfo.dct_method = C.J_DCT_METHOD(opt.DCTMethod)
}
