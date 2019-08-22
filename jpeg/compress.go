package jpeg

/*
#include <stdio.h>
#include <stdlib.h>
#include "jpeglib.h"
#include "jpeg.h"

static struct jpeg_compress_struct *new_compress(void) {
	struct jpeg_compress_struct *cinfo = (struct jpeg_compress_struct *) calloc(sizeof(struct jpeg_compress_struct), 1);
	if (!cinfo) {
		return NULL;
	}

	struct my_error_mgr *jerr = (struct my_error_mgr *)calloc(sizeof(struct my_error_mgr), 1);
	if (!jerr) {
		free(cinfo);
		return NULL;
	}

	cinfo->err = jpeg_std_error(&jerr->pub);
	jerr->pub.error_exit = (void *)error_longjmp;
	if (setjmp(jerr->jmpbuf) != 0) {
		free(jerr);
		free(cinfo);
		return NULL;
	}
	jpeg_create_compress(cinfo);

	return cinfo;
}

static void destroy_compress(struct jpeg_compress_struct *cinfo) {
	free(cinfo->err);
	jpeg_destroy_compress(cinfo);
	free(cinfo);
}

static JDIMENSION write_scanlines(j_compress_ptr cinfo, JSAMPROW row, JDIMENSION max_lines, int *msg_code) {
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)cinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		*msg_code = err->pub.msg_code;
		return 0;
	}

	*msg_code = 0;
	return jpeg_write_scanlines(cinfo, &row, max_lines);
}

static JDIMENSION write_mcu_gray(struct jpeg_compress_struct *cinfo, JSAMPROW pix, int stride, int *msg_code) {
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)cinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		*msg_code = err->pub.msg_code;
		return 0;
	}

	// Set height to one MCU size
	// because jpeg_write_raw_data processes just one MCU row per call.
	int height = DCTSIZE * cinfo->comp_info[0].v_samp_factor;

	// Allocate JSAMPIMAGE to hold pointers to one iMCU worth of image data
	// this is a safe overestimate; we use the return value from
	// jpeg_read_raw_data to figure out what is the actual iMCU row count.
	JSAMPROW *rows = alloca(sizeof(JSAMPROW *) * height);

	// First fill in the pointers into the plane data buffers
	int h = 0;
	for (h = 0; h < height; h++) {
		rows[h] = &pix[stride * h];
	}

	// Get the data
	*msg_code = 0;
	return jpeg_write_raw_data(cinfo, &rows, height);
}

static JDIMENSION write_mcu_ycbcr(struct jpeg_compress_struct *cinfo, JSAMPROW y_row, JSAMPROW cb_row, JSAMPROW cr_row, int y_stride, int c_stride, int *msg_code) {
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)cinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		*msg_code = err->pub.msg_code;
		return 0;
	}

	// Allocate JSAMPIMAGE to hold pointers to one iMCU worth of image data
	// this is a safe overestimate; we use the return value from
	// jpeg_read_raw_data to figure out what is the actual iMCU row count.
	int y_h = DCTSIZE * cinfo->comp_info[0].v_samp_factor;
	int c_h = DCTSIZE * cinfo->comp_info[1].v_samp_factor;
	JSAMPROW *y_rows = alloca(sizeof(JSAMPROW) * y_h);
	JSAMPROW *cb_rows = alloca(sizeof(JSAMPROW) * c_h);
	JSAMPROW *cr_rows = alloca(sizeof(JSAMPROW) * c_h);
	JSAMPARRAY image[] = {y_rows, cb_rows, cr_rows};
	int h = 0;

	// First fill in the pointers into the plane data buffers
	for (h = 0; h < y_h; h++) {
		y_rows[h] = &y_row[y_stride * h];
	}

	for (h = 0; h < c_h; h++) {
		cb_rows[h] = &cb_row[c_stride * h];
		cr_rows[h] = &cr_row[c_stride * h];
	}

	// Get the data
	*msg_code = 0;
	return jpeg_write_raw_data(cinfo, image, y_h);
}

static int start_compress(j_compress_ptr cinfo, boolean write_all_tables)
{
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)cinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		return err->pub.msg_code;
	}

	jpeg_start_compress(cinfo, write_all_tables);

	return 0;
}

static int finish_compress(j_compress_ptr cinfo)
{
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)cinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		return err->pub.msg_code;
	}

	jpeg_finish_compress(cinfo);

	return 0;
}

*/
import "C"

import (
	"errors"
	"image"
	"io"
	"unsafe"

	"github.com/pixiv/go-libjpeg/rgb"
)

// EncoderOptions specifies which settings to use during Compression.
type EncoderOptions struct {
	Quality         int
	OptimizeCoding  bool
	ProgressiveMode bool
	DCTMethod       DCTMethod
}

func newCompress(w io.Writer) (cinfo *C.struct_jpeg_compress_struct, err error) {
	cinfo = C.new_compress()
	if cinfo == nil {
		err = errors.New("failed to allocate jpeg encoder")
		return
	}

	_, err = makeDestinationManager(w, cinfo)
	return
}

func startCompress(cinfo *C.struct_jpeg_compress_struct) error {
	code := C.start_compress(cinfo, C.TRUE)
	if code != 0 {
		return errors.New(jpegErrorMessage(unsafe.Pointer(cinfo)))
	}
	return nil
}

func destroyCompress(cinfo *C.struct_jpeg_compress_struct) {
	if cinfo == nil {
		return
	}
	destinationManager := getDestinationManager(cinfo)
	if destinationManager != nil {
		releaseDestinationManager(destinationManager)
	}
	C.destroy_compress(cinfo)
}

func finishCompress(cinfo *C.struct_jpeg_compress_struct) error {
	code := C.finish_compress(cinfo)
	if code != 0 {
		return errors.New(jpegErrorMessage(unsafe.Pointer(cinfo)))
	}
	return nil
}

func writeScanline(cinfo *C.struct_jpeg_compress_struct, row C.JSAMPROW, maxLines C.JDIMENSION) (line int, err error) {
	code := C.int(0)
	line = int(C.write_scanlines(cinfo, row, maxLines, &code))
	if code != 0 {
		err = errors.New(jpegErrorMessage(unsafe.Pointer(cinfo)))
	}
	return
}

func writeMCUGray(cinfo *C.struct_jpeg_compress_struct, row C.JSAMPROW, stride int) (line int, err error) {
	code := C.int(0)
	line = int(C.write_mcu_gray(cinfo, row, C.int(stride), &code))
	if code != 0 {
		err = errors.New(jpegErrorMessage(unsafe.Pointer(cinfo)))
	}
	return
}

func writeMCUYCbCr(cinfo *C.struct_jpeg_compress_struct, y, cb, cr C.JSAMPROW, yStride, cStride int) (line int, err error) {
	code := C.int(0)
	line = int(C.write_mcu_ycbcr(cinfo, y, cb, cr, C.int(yStride), C.int(cStride), &code))
	if code != 0 {
		err = errors.New(jpegErrorMessage(unsafe.Pointer(cinfo)))
	}
	return
}

// Encode encodes src image and writes into w as JPEG format data.
func Encode(w io.Writer, src image.Image, opt *EncoderOptions) (err error) {
	var cinfo *C.struct_jpeg_compress_struct
	cinfo, err = newCompress(w)
	if err != nil {
		return
	}
	defer destroyCompress(cinfo)

	switch s := src.(type) {
	case *image.YCbCr:
		err = encodeYCbCr(cinfo, s, opt)
	case *image.Gray:
		err = encodeGray(cinfo, s, opt)
	case *image.RGBA:
		err = encodeRGBA(cinfo, s, opt)
	case *rgb.Image:
		err = encodeRGB(cinfo, s, opt)
	default:
		return errors.New("unsupported image type")
	}

	return
}

// encode image.YCbCr
func encodeYCbCr(cinfo *C.struct_jpeg_compress_struct, src *image.YCbCr, p *EncoderOptions) (err error) {
	// Set up compression parameters
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	cinfo.image_width = C.JDIMENSION(w)
	cinfo.image_height = C.JDIMENSION(h)
	cinfo.input_components = 3
	cinfo.in_color_space = C.JCS_YCbCr

	setupEncoderOptions(cinfo, p)

	compInfo := (*[3]C.jpeg_component_info)(unsafe.Pointer(cinfo.comp_info))
	cVDiv := 1
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
		cVDiv = 2
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
		cVDiv = 2
	}

	// libjpeg raw data in is in planar format, which avoids unnecessary
	// planar->packed->planar conversions.
	cinfo.raw_data_in = C.TRUE

	// Start compression
	err = startCompress(cinfo)
	if err != nil {
		return
	}
	defer func() {
		ferr := finishCompress(cinfo)
		if ferr != nil && err == nil {
			err = ferr
		}
	}()

	for v := 0; v < h; {
		yOff, cOff := v*src.YStride, v/cVDiv*src.CStride
		line, err := writeMCUYCbCr(
			cinfo,
			C.JSAMPROW(unsafe.Pointer(&src.Y[yOff])),
			C.JSAMPROW(unsafe.Pointer(&src.Cb[cOff])),
			C.JSAMPROW(unsafe.Pointer(&src.Cr[cOff])),
			src.YStride,
			src.CStride,
		)
		if err != nil {
			return err
		}
		v += line
	}
	return
}

// encode image.RGBA
func encodeRGBA(cinfo *C.struct_jpeg_compress_struct, src *image.RGBA, p *EncoderOptions) (err error) {
	// Set up compression parameters
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	cinfo.image_width = C.JDIMENSION(w)
	cinfo.image_height = C.JDIMENSION(h)
	cinfo.input_components = 4
	cinfo.in_color_space = getJCS_EXT_RGBA()
	if cinfo.in_color_space == C.JCS_UNKNOWN {
		return errors.New("JCS_EXT_RGBA is not supported (probably built without libjpeg-turbo)")
	}

	setupEncoderOptions(cinfo, p)

	// Start compression
	err = startCompress(cinfo)
	if err != nil {
		return
	}
	defer func() {
		ferr := finishCompress(cinfo)
		if ferr != nil && err == nil {
			err = ferr
		}
	}()

	for v := 0; v < h; {
		line, err := writeScanline(cinfo, C.JSAMPROW(unsafe.Pointer(&src.Pix[v*src.Stride])), C.JDIMENSION(1))
		if err != nil {
			return err
		}
		v += line
	}
	return
}

// encode rgb.Image.
func encodeRGB(cinfo *C.struct_jpeg_compress_struct, src *rgb.Image, p *EncoderOptions) (err error) {
	// Set up compression parameters
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	cinfo.image_width = C.JDIMENSION(w)
	cinfo.image_height = C.JDIMENSION(h)
	cinfo.input_components = 3
	cinfo.in_color_space = C.JCS_RGB

	setupEncoderOptions(cinfo, p)

	// Start compression
	err = startCompress(cinfo)
	if err != nil {
		return
	}
	defer func() {
		ferr := finishCompress(cinfo)
		if ferr != nil && err == nil {
			err = ferr
		}
	}()

	for v := 0; v < h; {
		line, err := writeScanline(cinfo, C.JSAMPROW(unsafe.Pointer(&src.Pix[v*src.Stride])), C.JDIMENSION(1))
		if err != nil {
			return err
		}
		v += line
	}
	return
}

// encode image.Gray
func encodeGray(cinfo *C.struct_jpeg_compress_struct, src *image.Gray, p *EncoderOptions) (err error) {
	// Set up compression parameters
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	cinfo.image_width = C.JDIMENSION(w)
	cinfo.image_height = C.JDIMENSION(h)
	cinfo.input_components = 1
	cinfo.in_color_space = C.JCS_GRAYSCALE

	setupEncoderOptions(cinfo, p)

	compInfo := (*C.jpeg_component_info)(unsafe.Pointer(cinfo.comp_info))
	compInfo.h_samp_factor, compInfo.v_samp_factor = 1, 1

	// libjpeg raw data in is in planar format, which avoids unnecessary
	// planar->packed->planar conversions.
	cinfo.raw_data_in = C.TRUE

	// Start compression
	err = startCompress(cinfo)
	if err != nil {
		return
	}
	defer func() {
		ferr := finishCompress(cinfo)
		if ferr != nil && err == nil {
			err = ferr
		}
	}()

	for v := 0; v < h; {
		line, err := writeMCUGray(cinfo, C.JSAMPROW(unsafe.Pointer(&src.Pix[v*src.Stride])), src.Stride)
		if err != nil {
			return err
		}
		v += line
	}
	return
}

func setupEncoderOptions(cinfo *C.struct_jpeg_compress_struct, opt *EncoderOptions) {
	C.jpeg_set_defaults(cinfo)
	C.jpeg_set_quality(cinfo, C.int(opt.Quality), C.TRUE)
	if opt.OptimizeCoding {
		cinfo.optimize_coding = C.TRUE
	} else {
		cinfo.optimize_coding = C.FALSE
	}
	if opt.ProgressiveMode {
		C.jpeg_simple_progression(cinfo)
	}
	cinfo.dct_method = C.J_DCT_METHOD(opt.DCTMethod)
}
