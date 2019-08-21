package jpeg

/*
#include <stdio.h>
#include <stdlib.h>
#include "jpeglib.h"
#include "jerror.h"
#include "jpeg.h"

void error_panic(j_common_ptr dinfo);

static struct jpeg_decompress_struct *new_decompress(void) {
	struct jpeg_decompress_struct *dinfo = (struct jpeg_decompress_struct *)calloc(sizeof(struct jpeg_decompress_struct), 1);
	if (!dinfo) {
		return NULL;
	}

	struct my_error_mgr *jerr = (struct my_error_mgr *)calloc(sizeof(struct my_error_mgr), 1);
	if (!jerr) {
		free(dinfo);
		return NULL;
	}

	dinfo->err = jpeg_std_error(&jerr->pub);
	jerr->pub.error_exit = (void *)error_longjmp;
	if (setjmp(jerr->jmpbuf) != 0) {
		free(jerr);
		free(dinfo);
		return NULL;
	}
	jpeg_create_decompress(dinfo);

	return dinfo;
}

static int start_decompress(j_decompress_ptr dinfo)
{
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)dinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		return err->pub.msg_code;
	}

	jpeg_start_decompress(dinfo);
	return 0;
}

static int finish_decompress(j_decompress_ptr dinfo)
{
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)dinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		return err->pub.msg_code;
	}

	jpeg_finish_decompress(dinfo);
	return 0;
}

static void destroy_decompress(struct jpeg_decompress_struct *dinfo) {
	free(dinfo->err);
	jpeg_destroy_decompress(dinfo);
	free(dinfo);
}

static int read_header(struct jpeg_decompress_struct *dinfo, int req_img)
{
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)dinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		return err->pub.msg_code;
	}

	jpeg_read_header(dinfo, req_img);
	return 0;
}

static JDIMENSION read_scanlines(j_decompress_ptr dinfo, unsigned char *buf, int stride, int height, int *msg_code) {
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)dinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		*msg_code = err->pub.msg_code;
		return 0;
	}

	JSAMPROW *rows = alloca(sizeof(JSAMPROW) * height);
	int i;
	for (i = 0; i < height; i++) {
		rows[i] = &buf[i * stride];
	}
	*msg_code = 0;
	return jpeg_read_scanlines(dinfo, rows, height);
}

static int DCT_v_scaled_size(j_decompress_ptr dinfo, int component) {
#if JPEG_LIB_VERSION >= 70
	return dinfo->comp_info[component].DCT_v_scaled_size;
#else
	return dinfo->comp_info[component].DCT_scaled_size;
#endif
}

static JDIMENSION read_mcu_gray(struct jpeg_decompress_struct *dinfo, JSAMPROW pix, int stride, int imcu_rows, int *msg_code) {
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)dinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		*msg_code = err->pub.msg_code;
		return 0;
	}

	JSAMPROW *rows = alloca(sizeof(JSAMPROW) * imcu_rows);
	int h = 0;
	for (h = 0; h < imcu_rows; h++) {
		rows[h] = &pix[stride * h];
	}

	// Get the data
	*msg_code = 0;
	return jpeg_read_raw_data(dinfo, &rows, imcu_rows);
}

static JDIMENSION read_mcu_ycbcr(struct jpeg_decompress_struct *dinfo, JSAMPROW y_row, JSAMPROW cb_row, JSAMPROW cr_row, int y_stride, int c_stride, int imcu_rows, int *msg_code) {
	// handle error
	struct my_error_mgr *err = (struct my_error_mgr *)dinfo->err;
	if (setjmp(err->jmpbuf) != 0) {
		*msg_code = err->pub.msg_code;
		return 0;
	}

	// Allocate JSAMPIMAGE to hold pointers to one iMCU worth of image data
	// this is a safe overestimate; we use the return value from
	// jpeg_read_raw_data to figure out what is the actual iMCU row count.
	JSAMPROW *y_rows = alloca(sizeof(JSAMPROW) * imcu_rows);
	JSAMPROW *cb_rows = alloca(sizeof(JSAMPROW) * imcu_rows);
	JSAMPROW *cr_rows = alloca(sizeof(JSAMPROW) * imcu_rows);
	JSAMPARRAY image[] = {y_rows, cb_rows, cr_rows};
	int x = 0;

	// First fill in the pointers into the plane data buffers
	int h = 0;
	for (h = 0; h < imcu_rows; h++) {
		y_rows[h] = &y_row[y_stride * h];
		cb_rows[h] = &cb_row[c_stride * h];
		cr_rows[h] = &cr_row[c_stride * h];
	}

	// Get the data
	*msg_code = 0;
	return jpeg_read_raw_data(dinfo, image, imcu_rows);
}

*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"unsafe"

	"github.com/pixiv/go-libjpeg/rgb"
)

func newDecompress(r io.Reader) *C.struct_jpeg_decompress_struct {
	dinfo := C.new_decompress()
	if dinfo == nil {
		return nil
	}
	makeSourceManager(r, dinfo)
	return dinfo
}

func destroyDecompress(dinfo *C.struct_jpeg_decompress_struct) {
	if dinfo == nil {
		return
	}
	sourceManager := getSourceManager(dinfo)
	if sourceManager != nil {
		releaseSourceManager(sourceManager)
	}
	C.destroy_decompress(dinfo)
}

func readHeader(dinfo *C.struct_jpeg_decompress_struct) error {
	if C.read_header(dinfo, C.TRUE) != 0 {
		return errors.New(jpegErrorMessage(unsafe.Pointer(dinfo)))
	}
	return nil
}

func startDecompress(dinfo *C.struct_jpeg_decompress_struct) error {
	if C.start_decompress(dinfo) != 0 {
		return errors.New(jpegErrorMessage(unsafe.Pointer(dinfo)))
	}
	return nil
}

func finishDecompress(dinfo *C.struct_jpeg_decompress_struct) error {
	if C.finish_decompress(dinfo) != 0 {
		return errors.New(jpegErrorMessage(unsafe.Pointer(dinfo)))
	}
	return nil
}

func readScanlines(dinfo *C.struct_jpeg_decompress_struct, row *C.uchar, stride, height C.int) (lines C.JDIMENSION, err error) {
	code := C.int(0)
	lines = C.read_scanlines(dinfo, row, stride, height, &code)
	if code != 0 {
		err = errors.New(jpegErrorMessage(unsafe.Pointer(dinfo)))
	} else if lines == 0 {
		err = errors.New("unexpected EOF")
	}
	return
}

func readMCUGray(dinfo *C.struct_jpeg_decompress_struct, pix C.JSAMPROW, stride, iMCURows int) (line C.JDIMENSION, err error) {
	code := C.int(0)
	line = C.read_mcu_gray(dinfo, pix, C.int(stride), C.int(iMCURows), &code)
	if code != 0 {
		err = errors.New(jpegErrorMessage(unsafe.Pointer(dinfo)))
	}
	return
}

func readMCUYCbCr(dinfo *C.struct_jpeg_decompress_struct, y, cb, cr C.JSAMPROW, yStride, cStride int, iMCURows int) (line C.JDIMENSION, err error) {
	code := C.int(0)
	line = C.read_mcu_ycbcr(dinfo, y, cb, cr, C.int(yStride), C.int(cStride), C.int(iMCURows), &code)
	if code != 0 {
		err = errors.New(jpegErrorMessage(unsafe.Pointer(dinfo)))
	}
	return
}

// DecoderOptions specifies JPEG decoding parameters.
type DecoderOptions struct {
	ScaleTarget            image.Rectangle // ScaleTarget is the target size to scale image.
	DCTMethod              DCTMethod       // DCTMethod is DCT Algorithm method.
	DisableFancyUpsampling bool            // If true, disable fancy upsampling
	DisableBlockSmoothing  bool            // If true, disable block smoothing
}

// SupportRGBA returns whether RGBA decoding is supported.
func SupportRGBA() bool {
	if getJCS_EXT_RGBA() == C.JCS_UNKNOWN {
		return false
	}
	return true
}

// Decode reads a JPEG data stream from r and returns decoded image as an image.Image.
// Output image has YCbCr colors or 8bit Grayscale.
func Decode(r io.Reader, options *DecoderOptions) (dest image.Image, err error) {
	dinfo := newDecompress(r)
	if dinfo == nil {
		return nil, errors.New("allocation failed")
	}
	defer destroyDecompress(dinfo)

	err = readHeader(dinfo)
	if err != nil {
		return nil, err
	}

	setupDecoderOptions(dinfo, options)

	switch dinfo.num_components {
	case 1:
		if dinfo.jpeg_color_space != C.JCS_GRAYSCALE {
			return nil, errors.New("unsupported colorspace")
		}
		dest, err = decodeGray(dinfo)
	case 3:
		switch dinfo.jpeg_color_space {
		case C.JCS_YCbCr:
			dest, err = decodeYCbCr(dinfo)
		case C.JCS_RGB:
			dest, err = decodeRGB(dinfo)
		default:
			return nil, errors.New("unsupported colorspace")
		}
	default:
		return nil, fmt.Errorf("unsupported number of components: %d", dinfo.num_components)
	}
	return
}

func decodeGray(dinfo *C.struct_jpeg_decompress_struct) (dest *image.Gray, err error) {
	// output dawnsampled raw data before starting decompress
	dinfo.raw_data_out = C.TRUE

	err = startDecompress(dinfo)
	if err != nil {
		return nil, err
	}
	defer func() {
		ferr := finishDecompress(dinfo)
		if ferr != nil && err == nil {
			err = ferr
		}
	}()

	compInfo := (*[1]C.jpeg_component_info)(unsafe.Pointer(dinfo.comp_info))
	dest = NewGrayAligned(image.Rect(0, 0, int(compInfo[0].downsampled_width), int(compInfo[0].downsampled_height)))

	iMCURows := int(C.DCT_v_scaled_size(dinfo, C.int(0)) * compInfo[0].v_samp_factor)

	for dinfo.output_scanline < dinfo.output_height {
		_, err = readMCUGray(dinfo, C.JSAMPROW(unsafe.Pointer(&dest.Pix[dest.Stride*int(dinfo.output_scanline)])), dest.Stride, iMCURows)
		if err != nil {
			return
		}
	}
	return
}

func decodeYCbCr(dinfo *C.struct_jpeg_decompress_struct) (dest *image.YCbCr, err error) {
	// output dawnsampled raw data before starting decompress
	dinfo.raw_data_out = C.TRUE

	err = startDecompress(dinfo)
	if err != nil {
		return nil, err
	}

	compInfo := (*[3]C.jpeg_component_info)(unsafe.Pointer(dinfo.comp_info))

	dwY := compInfo[Y].downsampled_width
	dhY := compInfo[Y].downsampled_height
	dwC := compInfo[Cb].downsampled_width
	dhC := compInfo[Cb].downsampled_height
	//fmt.Printf("%d %d %d %d\n", dwY, dhY, dwC, dhC)
	if dwC != compInfo[Cr].downsampled_width || dhC != compInfo[Cr].downsampled_height {
		return nil, errors.New("Unsupported color subsampling (Cb and Cr differ)")
	}

	// Since the decisions about which DCT size and subsampling mode
	// to use, if any, are complex, instead just check the calculated
	// output plane sizes and infer the subsampling mode from that.
	var subsampleRatio image.YCbCrSubsampleRatio
	cVDiv := 1
	switch {
	case dwY == dwC && dhY == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio444
	case dwY == dwC && (dhY+1)/2 == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio440
		cVDiv = 2
	case (dwY+1)/2 == dwC && dhY == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio422
	case (dwY+1)/2 == dwC && (dhY+1)/2 == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio420
		cVDiv = 2
	default:
		return nil, errors.New("Unsupported color subsampling")
	}

	// Allocate distination iamge
	dest = NewYCbCrAligned(image.Rect(0, 0, int(dinfo.output_width), int(dinfo.output_height)), subsampleRatio)

	var iMCURows int
	for i := 0; i < int(dinfo.num_components); i++ {
		compRows := int(C.DCT_v_scaled_size(dinfo, C.int(i)) * compInfo[i].v_samp_factor)
		if compRows > iMCURows {
			iMCURows = compRows
		}
	}
	yStride, cStride := dest.YStride, dest.CStride

	for dinfo.output_scanline < dinfo.output_height {
		y := C.JSAMPROW(unsafe.Pointer(&dest.Y[yStride*int(dinfo.output_scanline)]))
		cb := C.JSAMPROW(unsafe.Pointer(&dest.Cb[cStride*int(dinfo.output_scanline)/cVDiv]))
		cr := C.JSAMPROW(unsafe.Pointer(&dest.Cr[cStride*int(dinfo.output_scanline)/cVDiv]))
		_, err = readMCUYCbCr(dinfo, y, cb, cr, yStride, cStride, iMCURows)
		if err != nil {
			return
		}
	}
	return
}

func readRGBScanlines(dinfo *C.struct_jpeg_decompress_struct, pix []uint8, stride int) (err error) {
	err = startDecompress(dinfo)
	if err != nil {
		return
	}
	defer func() {
		ferr := finishDecompress(dinfo)
		if ferr != nil && err == nil {
			err = ferr
		}
	}()

	for dinfo.output_scanline < dinfo.output_height {
		pbuf := (*C.uchar)(unsafe.Pointer(&pix[stride*int(dinfo.output_scanline)]))
		_, err = readScanlines(dinfo, pbuf, C.int(stride), dinfo.rec_outbuf_height)
		if err != nil {
			return
		}
	}
	return
}

// TODO: supports decoding into image.RGBA instead of rgb.Image.
func decodeRGB(dinfo *C.struct_jpeg_decompress_struct) (dest *rgb.Image, err error) {
	C.jpeg_calc_output_dimensions(dinfo)
	dest = rgb.NewImage(image.Rect(0, 0, int(dinfo.output_width), int(dinfo.output_height)))

	dinfo.out_color_space = C.JCS_RGB
	err = readRGBScanlines(dinfo, dest.Pix, dest.Stride)
	return
}

// DecodeIntoRGB reads a JPEG data stream from r and returns decoded image as an rgb.Image with RGB colors.
func DecodeIntoRGB(r io.Reader, options *DecoderOptions) (dest *rgb.Image, err error) {
	dinfo := newDecompress(r)
	if dinfo == nil {
		return nil, errors.New("allocation failed")
	}
	defer destroyDecompress(dinfo)

	err = readHeader(dinfo)
	if err != nil {
		return nil, err
	}

	setupDecoderOptions(dinfo, options)
	return decodeRGB(dinfo)
}

// DecodeIntoRGBA reads a JPEG data stream from r and returns decoded image as an image.RGBA with RGBA colors.
// This function only works with libjpeg-turbo, not libjpeg.
func DecodeIntoRGBA(r io.Reader, options *DecoderOptions) (dest *image.RGBA, err error) {
	dinfo := newDecompress(r)
	if dinfo == nil {
		return nil, errors.New("allocation failed")
	}
	defer destroyDecompress(dinfo)

	// Recover panic
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(error); !ok {
				err = fmt.Errorf("JPEG error: %v", r)
			}
		}
	}()

	err = readHeader(dinfo)
	if err != nil {
		return nil, err
	}

	setupDecoderOptions(dinfo, options)

	C.jpeg_calc_output_dimensions(dinfo)
	dest = image.NewRGBA(image.Rect(0, 0, int(dinfo.output_width), int(dinfo.output_height)))

	colorSpace := getJCS_EXT_RGBA()
	if colorSpace == C.JCS_UNKNOWN {
		return nil, errors.New("JCS_EXT_RGBA is not supported (probably built without libjpeg-turbo)")
	}
	dinfo.out_color_space = colorSpace
	err = readRGBScanlines(dinfo, dest.Pix, dest.Stride)

	return
}

// DecodeConfig returns the color model and dimensions of a JPEG image without decoding the entire image.
func DecodeConfig(r io.Reader) (config image.Config, err error) {
	dinfo := newDecompress(r)
	if dinfo == nil {
		err = errors.New("allocation failed")
		return
	}
	defer destroyDecompress(dinfo)

	// Recover panic
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(error); !ok {
				err = fmt.Errorf("JPEG error: %v", r)
			}
		}
	}()

	err = readHeader(dinfo)
	if err != nil {
		return
	}

	config = image.Config{
		ColorModel: color.YCbCrModel,
		Width:      int(dinfo.image_width),
		Height:     int(dinfo.image_height),
	}
	return
}

func setupDecoderOptions(dinfo *C.struct_jpeg_decompress_struct, opt *DecoderOptions) {
	tw, th := opt.ScaleTarget.Dx(), opt.ScaleTarget.Dy()
	if tw > 0 && th > 0 {
		var scaleFactor int
		for scaleFactor = 1; scaleFactor <= 8; scaleFactor++ {
			if ((scaleFactor*int(dinfo.image_width)+7)/8) >= tw &&
				((scaleFactor*int(dinfo.image_height)+7)/8) >= th {
				break
			}
		}
		if scaleFactor < 8 {
			dinfo.scale_num = C.uint(scaleFactor)
			dinfo.scale_denom = 8
		}
	}

	dinfo.dct_method = C.J_DCT_METHOD(opt.DCTMethod)
	if opt.DisableFancyUpsampling {
		dinfo.do_fancy_upsampling = C.FALSE
	} else {
		dinfo.do_fancy_upsampling = C.TRUE
	}
	if opt.DisableBlockSmoothing {
		dinfo.do_block_smoothing = C.FALSE
	} else {
		dinfo.do_block_smoothing = C.TRUE
	}
}

const alignSize int = C.ALIGN_SIZE

// NewYCbCrAligned Allocates YCbCr image with padding.
// Because LibJPEG needs extra padding to decoding buffer, This func add an
// extra alignSize (16) padding to cover overflow from any such modes.
func NewYCbCrAligned(r image.Rectangle, subsampleRatio image.YCbCrSubsampleRatio) *image.YCbCr {
	w, h, cw, ch := r.Dx(), r.Dy(), 0, 0
	switch subsampleRatio {
	case image.YCbCrSubsampleRatio422:
		cw = (r.Max.X+1)/2 - r.Min.X/2
		ch = h
	case image.YCbCrSubsampleRatio420:
		cw = (r.Max.X+1)/2 - r.Min.X/2
		ch = (r.Max.Y+1)/2 - r.Min.Y/2
	case image.YCbCrSubsampleRatio440:
		cw = w
		ch = (r.Max.Y+1)/2 - r.Min.Y/2
	default:
		cw = w
		ch = h
	}

	// TODO: check the padding size to minimize memory allocation.
	yStride := pad(w, alignSize) + alignSize
	cStride := pad(cw, alignSize) + alignSize
	yHeight := pad(h, alignSize) + alignSize
	cHeight := pad(ch, alignSize) + alignSize

	b := make([]byte, yStride*yHeight+2*cStride*cHeight)
	return &image.YCbCr{
		Y:              b[:yStride*yHeight],
		Cb:             b[yStride*yHeight+0*cStride*cHeight : yStride*yHeight+1*cStride*cHeight],
		Cr:             b[yStride*yHeight+1*cStride*cHeight : yStride*yHeight+2*cStride*cHeight],
		SubsampleRatio: subsampleRatio,
		YStride:        yStride,
		CStride:        cStride,
		Rect:           r,
	}
}

func pad(a int, b int) int {
	return (a + (b - 1)) & (^(b - 1))
}

// NewGrayAligned Allocates Grey image with padding.
// This func add an extra padding to cover overflow from decoding image.
func NewGrayAligned(r image.Rectangle) *image.Gray {
	w, h := r.Dx(), r.Dy()

	// TODO: check the padding size to minimize memory allocation.
	stride := pad(w, alignSize) + alignSize
	ph := pad(h, alignSize) + alignSize

	pix := make([]uint8, stride*ph)
	return &image.Gray{
		Pix:    pix,
		Stride: stride,
		Rect:   r,
	}
}
