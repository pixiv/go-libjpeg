package jpeg

/*
#include <stdio.h>
#include <stdlib.h>
#include "jpeglib.h"

static void error_message(j_common_ptr cinfo, char *buf, size_t capa) {
	if (cinfo != NULL && cinfo->err != NULL) {
		if (cinfo->err->format_message != NULL) {
			(*cinfo->err->format_message)(cinfo, buf);
		}
		else {
			snprintf(buf, capa, "JPEG error code %d", cinfo->err->msg_code);
		}
	}
	else {
		snprintf(buf, capa, "JPEG unknown error");
	}
}
*/
import "C"

import (
	"unsafe"
)

func jpegErrorMessage(cinfo unsafe.Pointer) string {
	buf := C.calloc(1, C.JMSG_LENGTH_MAX)
	if buf == nil {
		return "cannot allocate memory"
	}
	defer C.free(buf)
	msg := (*C.char)(buf)
	C.error_message(C.j_common_ptr(cinfo), msg, C.JMSG_LENGTH_MAX)
	return C.GoString(msg)
}
