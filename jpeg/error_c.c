#ifndef __INTELLISENSE__
#include "_cgo_export.h"
#endif
#include "jpeg.h"

/* must not return */
void error_longjmp(j_common_ptr cinfo) {
	struct my_error_mgr *err = (struct my_error_mgr *)cinfo->err;
	longjmp(err->jmpbuf, err->pub.msg_code);
}
