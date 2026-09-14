package main

/*
#cgo CFLAGS: -I${SRCDIR}/skia/csrc
#include "goskia.h"
*/
import "C"

func main() {
	println("TOP_LEFT_GR_SURFACE_ORIGIN =", C.TOP_LEFT_GR_SURFACE_ORIGIN)
	println("BOTTOM_LEFT_GR_SURFACE_ORIGIN =", C.BOTTOM_LEFT_GR_SURFACE_ORIGIN)
}
