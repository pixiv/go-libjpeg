package jpeg

/*
#include <stdio.h>
#include <stdlib.h>
#include "jpeglib.h"

// export from golang
void goPanic(char*);

void error_panic(j_common_ptr dinfo);

*/
import "C"

//export goPanic
func goPanic(msg *C.char) {
	panic(C.GoString(msg))
}
