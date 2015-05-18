package jpeg

//
// Original codes are bollowed from go-thumber.
// Copyright (c) 2014 pixiv Inc. All rights reserved.
//
// See: https://github.com/pixiv/go-thumber
//

/*
#include <stdlib.h>
#include <stdio.h>
#include <jpeglib.h>

// exported from golang
void sourceInit(struct jpeg_decompress_struct*);
void sourceSkip(struct jpeg_decompress_struct*, long);
boolean sourceFill(struct jpeg_decompress_struct*);
void sourceTerm(struct jpeg_decompress_struct*);

static void* get_jpeg_resync_to_restart() {
	return jpeg_resync_to_restart;
}

*/
import "C"

import (
	"io"
	"unsafe"
)

const readBufferSize = 16384

type sourceManager struct {
	magic       uint32
	pub         C.struct_jpeg_source_mgr
	buffer      [readBufferSize]byte
	src         io.Reader
	startOfFile bool
	currentSize int
}

func getSourceManager(dinfo *C.struct_jpeg_decompress_struct) (ret *sourceManager) {
	// unsafe upcast magic to get the sourceManager associated with a dinfo
	ret = (*sourceManager)(unsafe.Pointer(uintptr(unsafe.Pointer(dinfo.src)) - unsafe.Offsetof(sourceManager{}.pub)))
	// just in case this ever breaks in a future release for some reason,
	// check the magic
	if ret.magic != magic {
		panic("Invalid sourceManager magic; upcast failed.")
	}
	return
}

//export sourceInit
func sourceInit(dinfo *C.struct_jpeg_decompress_struct) {
	mgr := getSourceManager(dinfo)
	mgr.startOfFile = true
}

//export sourceSkip
func sourceSkip(dinfo *C.struct_jpeg_decompress_struct, bytes C.long) {
	mgr := getSourceManager(dinfo)
	if bytes > 0 {
		for bytes >= C.long(mgr.pub.bytes_in_buffer) {
			bytes -= C.long(mgr.pub.bytes_in_buffer)
			sourceFill(dinfo)
		}
	}
	mgr.pub.bytes_in_buffer -= C.size_t(bytes)
	if mgr.pub.bytes_in_buffer != 0 {
		mgr.pub.next_input_byte = (*C.JOCTET)(&mgr.buffer[mgr.currentSize-int(mgr.pub.bytes_in_buffer)])
	}
}

//export sourceTerm
func sourceTerm(dinfo *C.struct_jpeg_decompress_struct) {
	// do nothing
}

//export sourceFill
func sourceFill(dinfo *C.struct_jpeg_decompress_struct) C.boolean {
	mgr := getSourceManager(dinfo)
	bytes, err := mgr.src.Read(mgr.buffer[:])
	mgr.pub.bytes_in_buffer = C.size_t(bytes)
	mgr.currentSize = bytes
	mgr.pub.next_input_byte = (*C.JOCTET)(&mgr.buffer[0])
	if err == io.EOF {
		if bytes == 0 {
			if mgr.startOfFile {
				panic("input is empty")
			}
			// EOF and need more data. Fill in a fake EOI to get a partial image.
			mgr.buffer[0] = 0xff
			mgr.buffer[1] = C.JPEG_EOI
			mgr.pub.bytes_in_buffer = 2
		}
	} else if err != nil {
		panic(err)
	}
	mgr.startOfFile = false

	return C.TRUE
}

func makeSourceManager(src io.Reader, dinfo *C.struct_jpeg_decompress_struct) (mgr sourceManager) {
	mgr.magic = magic
	mgr.src = src
	mgr.pub.init_source = (*[0]byte)(C.sourceInit)
	mgr.pub.fill_input_buffer = (*[0]byte)(C.sourceFill)
	mgr.pub.skip_input_data = (*[0]byte)(C.sourceSkip)
	mgr.pub.resync_to_restart = (*[0]byte)(C.get_jpeg_resync_to_restart()) // default implementation
	mgr.pub.term_source = (*[0]byte)(C.sourceTerm)
	mgr.pub.bytes_in_buffer = 0
	mgr.pub.next_input_byte = nil
	dinfo.src = &mgr.pub
	return
}
