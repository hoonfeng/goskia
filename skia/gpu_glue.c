// C trampoline bridging Skia's GL proc-address loader callback to a Go
// function. See gpu.go.
#include "goskia.h"
#include "_cgo_export.h"

// goskia_get_proc is the gr_gl_get_proc callback handed to Skia. It forwards to
// the Go side, keyed by an integer id smuggled through the void* ctx slot.
static gr_gl_func_ptr goskia_get_proc(void* ctx, const char* name) {
	return (gr_gl_func_ptr)goskiaGLGetProc((int)(intptr_t)ctx, (char*)name);
}

const gr_glinterface_t* goskia_assemble_gl(int id) {
	return gr_glinterface_assemble_gl_interface((void*)(intptr_t)id, goskia_get_proc);
}
