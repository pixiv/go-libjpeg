#pragma once

#include <stdio.h>
#include <stdlib.h>
#include <setjmp.h>
#include "jpeglib.h"
#include "jerror.h"

// the dimension multiple to which data buffers should be aligned.
#define ALIGN_SIZE 16

struct my_error_mgr {
	struct jpeg_error_mgr pub;
	jmp_buf jmpbuf;
};

#if defined(_WIN32) && !defined(__CYGWIN__)
// setjmp/longjmp occasionally crashes on Windows
// see https://github.com/golang/go/issues/13672
// use __builtin_setjmp/longjmp for workaround
#undef setjmp
#define setjmp(b)		__builtin_setjmp(b)
#undef longjmp
#define longjmp(b, c)	__builtin_longjmp((b), 1)	// __builtin_longjmp accepts only `1` as the secound argument
#endif

void error_longjmp(j_common_ptr cinfo);
