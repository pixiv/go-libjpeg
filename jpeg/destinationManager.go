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
void destinationInit(struct jpeg_compress_struct*);
boolean destinationEmpty(struct jpeg_compress_struct*);
void destinationTerm(struct jpeg_compress_struct*);

*/
import "C"

import (
	"io"
	"unsafe"
)

const writeBufferSize = 16384

type destinationManager struct {
	magic  uint32
	pub    C.struct_jpeg_destination_mgr
	buffer [writeBufferSize]byte
	dest   io.Writer
}

func getDestinationManager(cinfo *C.struct_jpeg_compress_struct) (ret *destinationManager) {
	// unsafe upcast magic to get the destinationManager associated with a cinfo
	ret = (*destinationManager)(unsafe.Pointer(uintptr(unsafe.Pointer(cinfo.dest)) - unsafe.Offsetof(destinationManager{}.pub)))
	// just in case this ever breaks in a future release for some reason,
	// check the magic
	if ret.magic != magic {
		panic("Invalid destinationManager magic; upcast failed.")
	}
	return
}

//export destinationInit
func destinationInit(cinfo *C.struct_jpeg_compress_struct) {
	// do nothing
}

func flushBuffer(mgr *destinationManager, inBuffer int) {
	wrote := 0
	for wrote != inBuffer {
		bytes, err := mgr.dest.Write(mgr.buffer[wrote:inBuffer])
		if err != nil {
			releaseDestinationManager(mgr)
			panic(err)
		}
		wrote += int(bytes)
	}
	mgr.pub.free_in_buffer = writeBufferSize
	mgr.pub.next_output_byte = (*C.JOCTET)(&mgr.buffer[0])
}

//export destinationEmpty
func destinationEmpty(cinfo *C.struct_jpeg_compress_struct) C.boolean {
	// need to write *entire* buffer, not subtracting free_in_buffer
	mgr := getDestinationManager(cinfo)
	flushBuffer(mgr, writeBufferSize)
	return C.TRUE
}

//export destinationTerm
func destinationTerm(cinfo *C.struct_jpeg_compress_struct) {
	// just empty buffer
	mgr := getDestinationManager(cinfo)
	inBuffer := int(writeBufferSize - mgr.pub.free_in_buffer)
	flushBuffer(mgr, inBuffer)
}

func makeDestinationManager(dest io.Writer, cinfo *C.struct_jpeg_compress_struct) (mgr *destinationManager) {
	mgr = (*destinationManager)(C.malloc(C.size_t(unsafe.Sizeof(destinationManager{}))))
	if mgr == nil {
		panic("Failed to allocate destinationManager")
	}
	mgr.magic = magic
	mgr.dest = dest
	mgr.pub.init_destination = (*[0]byte)(C.destinationInit)
	mgr.pub.empty_output_buffer = (*[0]byte)(C.destinationEmpty)
	mgr.pub.term_destination = (*[0]byte)(C.destinationTerm)
	mgr.pub.free_in_buffer = writeBufferSize
	mgr.pub.next_output_byte = (*C.JOCTET)(&mgr.buffer[0])
	cinfo.dest = &mgr.pub
	return
}

func releaseDestinationManager(mgr *destinationManager) {
	C.free(unsafe.Pointer(mgr))
}
