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

static struct jpeg_destination_mgr *calloc_jpeg_destination_mgr(void) {
	return calloc(sizeof(struct jpeg_destination_mgr), 1);
}

static void free_jpeg_destination_mgr(struct jpeg_destination_mgr *p) {
	free(p);
}

*/
import "C"

import (
	"errors"
	"io"
	"sync"
	"unsafe"
)

const writeBufferSize = 16384

var destinationManagerMapMutex sync.RWMutex
var destinationManagerMap = make(map[uintptr]*destinationManager)

// DestinationManagerMapLen returns the number of globally working destinationManagers for debug.
func DestinationManagerMapLen() int {
	return len(destinationManagerMap)
}

type destinationManager struct {
	pub    *C.struct_jpeg_destination_mgr
	buffer unsafe.Pointer
	dest   io.Writer
}

func getDestinationManager(cinfo *C.struct_jpeg_compress_struct) (ret *destinationManager) {
	destinationManagerMapMutex.RLock()
	defer destinationManagerMapMutex.RUnlock()
	return destinationManagerMap[uintptr(unsafe.Pointer(cinfo.dest))]
}

//export destinationInit
func destinationInit(cinfo *C.struct_jpeg_compress_struct) {
	// do nothing
}

func flushBuffer(mgr *destinationManager, inBuffer int) error {
	wrote := 0
	for wrote != inBuffer {
		slice := C.GoBytes(unsafe.Pointer(uintptr(mgr.buffer)+uintptr(wrote)), C.int(inBuffer-wrote))
		bytes, err := mgr.dest.Write(slice)
		if err != nil {
			return err
		}
		wrote += bytes
	}
	mgr.pub.free_in_buffer = writeBufferSize
	mgr.pub.next_output_byte = (*C.JOCTET)(mgr.buffer)
	return nil
}

//export destinationEmpty
func destinationEmpty(cinfo *C.struct_jpeg_compress_struct) C.boolean {
	// need to write *entire* buffer, not subtracting free_in_buffer
	mgr := getDestinationManager(cinfo)
	err := flushBuffer(mgr, writeBufferSize)
	if err != nil {
		return C.FALSE
	}
	return C.TRUE
}

//export destinationTerm
func destinationTerm(cinfo *C.struct_jpeg_compress_struct) {
	// just empty buffer
	mgr := getDestinationManager(cinfo)
	inBuffer := int(writeBufferSize - mgr.pub.free_in_buffer)
	flushBuffer(mgr, inBuffer) // can ignore error here
}

func makeDestinationManager(dest io.Writer, cinfo *C.struct_jpeg_compress_struct) (mgr *destinationManager, err error) {
	mgr = new(destinationManager)
	mgr.dest = dest
	mgr.pub = C.calloc_jpeg_destination_mgr()
	if mgr.pub == nil {
		err = errors.New("failed to allocate C.struct_jpeg_destination_mgr")
		return
	}
	mgr.buffer = C.calloc(writeBufferSize, 1)
	if mgr.buffer == nil {
		C.free_jpeg_destination_mgr(mgr.pub)
		err = errors.New("failed to allocate buffer")
		return
	}
	mgr.pub.init_destination = (*[0]byte)(C.destinationInit)
	mgr.pub.empty_output_buffer = (*[0]byte)(C.destinationEmpty)
	mgr.pub.term_destination = (*[0]byte)(C.destinationTerm)
	mgr.pub.free_in_buffer = writeBufferSize
	mgr.pub.next_output_byte = (*C.JOCTET)(mgr.buffer)
	cinfo.dest = mgr.pub

	destinationManagerMapMutex.Lock()
	defer destinationManagerMapMutex.Unlock()
	destinationManagerMap[uintptr(unsafe.Pointer(mgr.pub))] = mgr

	return
}

func releaseDestinationManager(mgr *destinationManager) {
	destinationManagerMapMutex.Lock()
	defer destinationManagerMapMutex.Unlock()
	var key = uintptr(unsafe.Pointer(mgr.pub))
	if _, ok := destinationManagerMap[key]; ok {
		delete(destinationManagerMap, key)
		C.free_jpeg_destination_mgr(mgr.pub)
		C.free(mgr.buffer)
	}
}
