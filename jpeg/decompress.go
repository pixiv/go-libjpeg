package jpeg

/*
#include <stdio.h>
#include <stdlib.h>
#include "jpeglib.h"

void error_panic(j_common_ptr dinfo);

static struct jpeg_decompress_struct *new_decompress() {
	struct jpeg_decompress_struct *dinfo = (struct jpeg_decompress_struct *)malloc(sizeof(struct jpeg_decompress_struct));
	struct jpeg_error_mgr *jerr = (struct jpeg_error_mgr *)malloc(sizeof(struct jpeg_error_mgr));

  jpeg_std_error(jerr);
  jerr->error_exit = (void *)error_panic;
	jpeg_create_decompress(dinfo);
  dinfo->err = jerr;

	return dinfo;
}

static void destroy_decompress(struct jpeg_decompress_struct *dinfo) {
	free(dinfo->err);
	jpeg_destroy_decompress(dinfo);
}

static int DCT_v_scaled_size(j_decompress_ptr dinfo, int component) {
#if JPEG_LIB_VERSION >= 70
	return dinfo->comp_info[component].DCT_v_scaled_size;
#else
	return dinfo->comp_info[component].DCT_scaled_size;
#endif
}

static J_COLOR_SPACE getJCS_EXT_RGBA() {
#ifdef JCS_EXT_RGBA
	return JCS_EXT_RGBA;
#endif
  return JCS_UNKNOWN;
}

*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"unsafe"

	"github.com/pixiv/go-libjpeg/rgb"
)

// DecoderOptions specifies JPEG decoding parameters.
type DecoderOptions struct {
	ScaleTarget            image.Rectangle // ScaleTarget is the target size to scale image.
	DCTMethod              DCTMethod       // DCTMethod is DCT Algorithm method.
	DisableFancyUpsampling bool            // If true, disable fancy upsampling
	DisableBlockSmoothing  bool            // If true, disable block smoothing
}

// SupportRGBA returns whether RGBA decoding is supported.
func SupportRGBA() bool {
	if C.getJCS_EXT_RGBA() == C.JCS_UNKNOWN {
		return false
	}
	return true
}

// Decode reads a JPEG data stream from r and returns decoded image as an image.Image.
// Output image has YCbCr colors or 8bit Grayscale.
func Decode(r io.Reader, options *DecoderOptions) (dest image.Image, err error) {
	// Recover panic
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			if _, ok := r.(error); !ok {
				err = fmt.Errorf("JPEG error: %v", r)
			}
		}
	}()

	var dinfo *C.struct_jpeg_decompress_struct = C.new_decompress()
	defer C.destroy_decompress(dinfo)

	srcManager := makeSourceManager(r, dinfo)
	defer C.free(unsafe.Pointer(srcManager))
	C.jpeg_read_header(dinfo, C.TRUE)

	setupDecoderOptions(dinfo, options)

	switch dinfo.num_components {
	case 1:
		if dinfo.jpeg_color_space != C.JCS_GRAYSCALE {
			return nil, errors.New("Image has unsupported colorspace")
		}
		dest, err = decodeGray(dinfo)
	case 3:
		switch dinfo.jpeg_color_space {
		case C.JCS_YCbCr:
			dest, err = decodeYCbCr(dinfo)
		case C.JCS_RGB:
			dest, err = decodeRGB(dinfo)
		default:
			return nil, errors.New("Image has unsupported colorspace")
		}
	}
	return
}

func decodeGray(dinfo *C.struct_jpeg_decompress_struct) (dest *image.Gray, err error) {
	// output dawnsampled raw data before starting decompress
	dinfo.raw_data_out = C.TRUE

	C.jpeg_start_decompress(dinfo)

	compInfo := (*[1]C.jpeg_component_info)(unsafe.Pointer(dinfo.comp_info))
	dest = NewGrayAligned(image.Rect(0, 0, int(compInfo[0].downsampled_width), int(compInfo[0].downsampled_height)))

	var rowPtr [AlignSize]C.JSAMPROW
	arrayPtr := [1]C.JSAMPARRAY{
		C.JSAMPARRAY(unsafe.Pointer(&rowPtr[0])),
	}

	iMCURows := int(C.DCT_v_scaled_size(dinfo, C.int(0)) * compInfo[0].v_samp_factor)

	for dinfo.output_scanline < dinfo.output_height {
		for j := 0; j < iMCURows; j++ {
			rowPtr[j] = C.JSAMPROW(unsafe.Pointer(&dest.Pix[dest.Stride*(int(dinfo.output_scanline)+j)]))
		}
		// Get the data
		C.jpeg_read_raw_data(dinfo, C.JSAMPIMAGE(unsafe.Pointer(&arrayPtr[0])), C.JDIMENSION(2*iMCURows))
	}

	return
}

func decodeYCbCr(dinfo *C.struct_jpeg_decompress_struct) (dest *image.YCbCr, err error) {
	// output dawnsampled raw data before starting decompress
	dinfo.raw_data_out = C.TRUE

	C.jpeg_start_decompress(dinfo)

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
	colorVDiv := 1
	switch {
	case dwY == dwC && dhY == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio444
	case dwY == dwC && (dhY+1)/2 == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio440
		colorVDiv = 2
	case (dwY+1)/2 == dwC && dhY == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio422
	case (dwY+1)/2 == dwC && (dhY+1)/2 == dhC:
		subsampleRatio = image.YCbCrSubsampleRatio420
		colorVDiv = 2
	default:
		return nil, errors.New("Unsupported color subsampling")
	}

	// Allocate distination iamge
	dest = NewYCbCrAligned(image.Rect(0, 0, int(dinfo.output_width), int(dinfo.output_height)), subsampleRatio)

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

	var iMCURows int
	for i := 0; i < int(dinfo.num_components); i++ {
		compRows := int(C.DCT_v_scaled_size(dinfo, C.int(i)) * compInfo[i].v_samp_factor)
		if compRows > iMCURows {
			iMCURows = compRows
		}
	}
	//fmt.Printf("iMCU_rows: %d (div: %d)\n", iMCURows, colorVDiv)
	for dinfo.output_scanline < dinfo.output_height {
		// First fill in the pointers into the plane data buffers
		for j := 0; j < iMCURows; j++ {
			yRowPtr[j] = C.JSAMPROW(unsafe.Pointer(&dest.Y[dest.YStride*(int(dinfo.output_scanline)+j)]))
			cbRowPtr[j] = C.JSAMPROW(unsafe.Pointer(&dest.Cb[dest.CStride*(int(dinfo.output_scanline)/colorVDiv+j)]))
			crRowPtr[j] = C.JSAMPROW(unsafe.Pointer(&dest.Cr[dest.CStride*(int(dinfo.output_scanline)/colorVDiv+j)]))
		}
		// Get the data
		C.jpeg_read_raw_data(dinfo, C.JSAMPIMAGE(unsafe.Pointer(&yCbCrPtr[0])), C.JDIMENSION(2*iMCURows))
	}

	C.jpeg_finish_decompress(dinfo)
	return
}

// TODO: supports decoding into image.RGBA instead of rgb.Image.
func decodeRGB(dinfo *C.struct_jpeg_decompress_struct) (dest *rgb.Image, err error) {
	C.jpeg_calc_output_dimensions(dinfo)
	dest = rgb.NewImage(image.Rect(0, 0, int(dinfo.output_width), int(dinfo.output_height)))

	dinfo.out_color_space = C.JCS_RGB
	readScanLines(dinfo, dest.Pix, dest.Stride)
	return
}

// DecodeIntoRGB reads a JPEG data stream from r and returns decoded image as an rgb.Image with RGB colors.
func DecodeIntoRGB(r io.Reader, options *DecoderOptions) (dest *rgb.Image, err error) {
	// Recover panic
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			if _, ok := r.(error); !ok {
				err = fmt.Errorf("JPEG error: %v", r)
			}
		}
	}()

	var dinfo *C.struct_jpeg_decompress_struct = C.new_decompress()
	defer C.destroy_decompress(dinfo)

	srcManager := makeSourceManager(r, dinfo)
	defer C.free(unsafe.Pointer(srcManager))
	C.jpeg_read_header(dinfo, C.TRUE)
	setupDecoderOptions(dinfo, options)

	C.jpeg_calc_output_dimensions(dinfo)
	dest = rgb.NewImage(image.Rect(0, 0, int(dinfo.output_width), int(dinfo.output_height)))

	dinfo.out_color_space = C.JCS_RGB
	readScanLines(dinfo, dest.Pix, dest.Stride)
	return
}

// DecodeIntoRGBA reads a JPEG data stream from r and returns decoded image as an image.RGBA with RGBA colors.
// This function only works with libjpeg-trubo, not libjpeg.
func DecodeIntoRGBA(r io.Reader, options *DecoderOptions) (dest *image.RGBA, err error) {
	// Recover panic
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			if _, ok := r.(error); !ok {
				err = fmt.Errorf("JPEG error: %v", r)
			}
		}
	}()

	var dinfo *C.struct_jpeg_decompress_struct = C.new_decompress()
	defer C.destroy_decompress(dinfo)

	srcManager := makeSourceManager(r, dinfo)
	defer C.free(unsafe.Pointer(srcManager))
	C.jpeg_read_header(dinfo, C.TRUE)
	setupDecoderOptions(dinfo, options)

	C.jpeg_calc_output_dimensions(dinfo)
	dest = image.NewRGBA(image.Rect(0, 0, int(dinfo.output_width), int(dinfo.output_height)))

	colorSpace := C.getJCS_EXT_RGBA()
	if colorSpace == C.JCS_UNKNOWN {
		return nil, errors.New("JCS_EXT_RGBA is not supported (probably built without libjpeg-trubo)")
	}

	dinfo.out_color_space = colorSpace
	readScanLines(dinfo, dest.Pix, dest.Stride)
	return
}

func readScanLines(dinfo *C.struct_jpeg_decompress_struct, buf []uint8, stride int) {
	var rowPtr C.JSAMPROW                             // pointer to pixel lines
	arrayPtr := C.JSAMPARRAY(unsafe.Pointer(&rowPtr)) // pointer to row pointers

	C.jpeg_start_decompress(dinfo)
	for dinfo.output_scanline < dinfo.output_height {
		rowPtr = C.JSAMPROW(unsafe.Pointer(&buf[stride*int(dinfo.output_scanline)]))
		C.jpeg_read_scanlines(dinfo, arrayPtr, C.JDIMENSION(dinfo.rec_outbuf_height))
	}
	C.jpeg_finish_decompress(dinfo)
}

// DecodeConfig returns the color model and dimensions of a JPEG image without decoding the entire image.
func DecodeConfig(r io.Reader) (config image.Config, err error) {
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

	var dinfo *C.struct_jpeg_decompress_struct = C.new_decompress()
	defer C.destroy_decompress(dinfo)

	srcManager := makeSourceManager(r, dinfo)
	defer C.free(unsafe.Pointer(srcManager))
	C.jpeg_read_header(dinfo, C.TRUE)

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

// AlignSize is the dimension multiple to which data buffers should be aligned.
const AlignSize int = 16

// NewYCbCrAligned Allocates YCbCr image with padding.
// Because LibJPEG needs extra padding to decoding buffer, This func add an
// extra AlignSize (16) padding to cover overflow from any such modes.
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
	yStride := pad(w, AlignSize) + AlignSize
	cStride := pad(cw, AlignSize) + AlignSize
	yHeight := pad(h, AlignSize) + AlignSize
	cHeight := pad(ch, AlignSize) + AlignSize

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
	stride := pad(w, AlignSize) + AlignSize
	ph := pad(h, AlignSize) + AlignSize

	pix := make([]uint8, stride*ph)
	return &image.Gray{
		Pix:    pix,
		Stride: stride,
		Rect:   r,
	}
}
