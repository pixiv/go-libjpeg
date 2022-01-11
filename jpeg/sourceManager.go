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
#include <string.h>
#include "jpeglib.h"

// exported from golang
void sourceInit(struct jpeg_decompress_struct*);
void sourceSkip(struct jpeg_decompress_struct*, long);
boolean sourceFill(struct jpeg_decompress_struct*);
void sourceTerm(struct jpeg_decompress_struct*);

// _get_jpeg_resync_to_restart returns the pointer of jpeg_resync_to_restart.
// see https://github.com/golang/go/issues/9411.
static void* _get_jpeg_resync_to_restart() {
	return jpeg_resync_to_restart;
}

static struct jpeg_source_mgr *calloc_jpeg_source_mgr(void) {
	return calloc(sizeof(struct jpeg_source_mgr), 1);
}

static void free_jpeg_source_mgr(struct jpeg_source_mgr *p) {
	free(p);
}

*/
import "C"

import (
	"errors"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

func makePseudoSlice(ptr unsafe.Pointer) []byte {
	var buffer []byte
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&buffer))
	slice.Cap = readBufferSize
	slice.Len = readBufferSize
	slice.Data = uintptr(ptr)
	return buffer
}

const readBufferSize = 16384

var sourceManagerMapMutex sync.RWMutex
var sourceManagerMap = make(map[uintptr]*sourceManager)

// SourceManagerMapLen returns the number of globally working sourceManagers for debug.
func SourceManagerMapLen() int {
	return len(sourceManagerMap)
}

type sourceManager struct {
	pub         *C.struct_jpeg_source_mgr
	buffer      unsafe.Pointer
	src         io.Reader
	startOfFile bool
	currentSize int
}

func getSourceManager(dinfo *C.struct_jpeg_decompress_struct) (ret *sourceManager) {
	sourceManagerMapMutex.RLock()
	defer sourceManagerMapMutex.RUnlock()
	return sourceManagerMap[uintptr(unsafe.Pointer(dinfo.src))]
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
			if sourceFill(dinfo) != C.TRUE {
				break
			}
		}
	}
	mgr.pub.bytes_in_buffer -= C.size_t(bytes)
	if mgr.pub.bytes_in_buffer != 0 {
		next := unsafe.Pointer(uintptr(mgr.buffer) + uintptr(mgr.currentSize-int(mgr.pub.bytes_in_buffer)))
		mgr.pub.next_input_byte = (*C.JOCTET)(next)
	}
}

//export sourceTerm
func sourceTerm(dinfo *C.struct_jpeg_decompress_struct) {
	// do nothing
}

//export sourceFill
func sourceFill(dinfo *C.struct_jpeg_decompress_struct) C.boolean {
	mgr := getSourceManager(dinfo)
	buffer := makePseudoSlice(mgr.buffer)
	bytes, err := mgr.src.Read(buffer)
	mgr.pub.bytes_in_buffer = C.size_t(bytes)
	mgr.currentSize = bytes
	mgr.pub.next_input_byte = (*C.JOCTET)(mgr.buffer)
	if err == io.EOF {
		if bytes == 0 {
			if mgr.startOfFile {
				return C.FALSE
			}
			// EOF and need more data. Fill in a fake EOI to get a partial image.
			mgr.pub.bytes_in_buffer = C.size_t(copy(buffer, []byte{0xff, C.JPEG_EOI}))
		}
	} else if err != nil {
		return C.FALSE
	}
	mgr.startOfFile = false

	return C.TRUE
}

func makeSourceManager(src io.Reader, dinfo *C.struct_jpeg_decompress_struct) (mgr *sourceManager, err error) {
	mgr = new(sourceManager)
	mgr.src = src
	mgr.pub = C.calloc_jpeg_source_mgr()
	if mgr.pub == nil {
		err = errors.New("failed to allocate C.struct_jpeg_source_mgr")
		return
	}
	mgr.buffer = C.calloc(readBufferSize, 1)
	if mgr.buffer == nil {
		C.free_jpeg_source_mgr(mgr.pub)
		err = errors.New("failed to allocate buffer")
		return
	}
	mgr.pub.init_source = (*[0]byte)(C.sourceInit)
	mgr.pub.fill_input_buffer = (*[0]byte)(C.sourceFill)
	mgr.pub.skip_input_data = (*[0]byte)(C.sourceSkip)
	mgr.pub.resync_to_restart = (*[0]byte)(C._get_jpeg_resync_to_restart())
	mgr.pub.term_source = (*[0]byte)(C.sourceTerm)
	mgr.pub.bytes_in_buffer = 0
	mgr.pub.next_input_byte = nil
	dinfo.src = mgr.pub

	sourceManagerMapMutex.Lock()
	defer sourceManagerMapMutex.Unlock()
	sourceManagerMap[uintptr(unsafe.Pointer(mgr.pub))] = mgr

	return
}

func releaseSourceManager(mgr *sourceManager) {
	sourceManagerMapMutex.Lock()
	defer sourceManagerMapMutex.Unlock()
	var key = uintptr(unsafe.Pointer(mgr.pub))
	if _, ok := sourceManagerMap[key]; ok {
		delete(sourceManagerMap, key)
		C.free_jpeg_source_mgr(mgr.pub)
		C.free(mgr.buffer)
	}
}
