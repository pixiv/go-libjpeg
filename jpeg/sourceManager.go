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
#include <jpeglib.h>

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

static struct jpeg_source_mgr *malloc_jpeg_source_mgr(void) {
	return malloc(sizeof(struct jpeg_source_mgr));
}

static void free_jpeg_source_mgr(struct jpeg_source_mgr *p) {
	free(p);
}

*/
import "C"

import (
	"io"
	"sync"
	"unsafe"
)

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
			sourceFill(dinfo)
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
	buffer := [readBufferSize]byte{}
	bytes, err := mgr.src.Read(buffer[:])
	C.memcpy(mgr.buffer, unsafe.Pointer(&buffer[0]), C.size_t(bytes))
	mgr.pub.bytes_in_buffer = C.size_t(bytes)
	mgr.currentSize = bytes
	mgr.pub.next_input_byte = (*C.JOCTET)(mgr.buffer)
	if err == io.EOF {
		if bytes == 0 {
			if mgr.startOfFile {
				releaseSourceManager(mgr)
				panic("input is empty")
			}
			// EOF and need more data. Fill in a fake EOI to get a partial image.
			footer := []byte{0xff, C.JPEG_EOI}
			C.memcpy(mgr.buffer, unsafe.Pointer(&footer[0]), C.size_t(len(footer)))
			mgr.pub.bytes_in_buffer = 2
		}
	} else if err != nil {
		releaseSourceManager(mgr)
		panic(err)
	}
	mgr.startOfFile = false

	return C.TRUE
}

func makeSourceManager(src io.Reader, dinfo *C.struct_jpeg_decompress_struct) (mgr *sourceManager) {
	mgr = new(sourceManager)
	mgr.src = src
	mgr.pub = C.malloc_jpeg_source_mgr()
	if mgr.pub == nil {
		panic("Failed to allocate C.struct_jpeg_source_mgr")
	}
	mgr.buffer = C.malloc(readBufferSize)
	if mgr.buffer == nil {
		panic("Failed to allocate buffer")
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
