#pragma once

#include <stdio.h>
#include <stdlib.h>
#include <setjmp.h>
#include "jpeglib.h"

int jpegtran_get_orientation (j_decompress_ptr cinfo);
