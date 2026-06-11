package skia

/*
#include "goskia.h"

// implemented in gpu_glue.c
const gr_glinterface_t* goskia_assemble_gl(int id);
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Common OpenGL sized internal formats, for use with NewGPUSurfaceFromFBO.
const (
	GLRGBA8   = 0x8058 // GL_RGBA8
	GLRGB8    = 0x8051 // GL_RGB8
	GLBGRA8   = 0x93A1 // GL_BGRA8_EXT
	GLRGBA16F = 0x881A // GL_RGBA16F
)

// GLGetProcAddress resolves the address of a named OpenGL function. Windowing
// libraries provide one (e.g. glfw.GetProcAddress); wrap it to return an
// unsafe.Pointer.
type GLGetProcAddress func(name string) unsafe.Pointer

var (
	glProcMu       sync.Mutex
	glProcRegistry = map[int]GLGetProcAddress{}
	glProcNextID   int
)

//export goskiaGLGetProc
func goskiaGLGetProc(id C.int, name *C.char) unsafe.Pointer {
	glProcMu.Lock()
	loader := glProcRegistry[int(id)]
	glProcMu.Unlock()
	if loader == nil {
		return nil
	}
	return loader(C.GoString(name))
}

// GLInterface is a set of resolved OpenGL function pointers Skia uses to drive
// the GPU. Create one after making a GL context current.
type GLInterface struct{ ptr *C.gr_glinterface_t }

// NewNativeGLInterface assembles a GL interface from the current platform GL
// context using Skia's built-in loader. A GL context must already be current.
func NewNativeGLInterface() (*GLInterface, error) {
	p := C.gr_glinterface_create_native_interface()
	if p == nil {
		return nil, fmt.Errorf("skia: gpu: failed to create native GL interface (no current context?)")
	}
	return newGLInterface(p), nil
}

// NewGLInterface assembles a GL interface using the provided proc-address
// loader. A matching GL context must be current for the duration of this call.
func NewGLInterface(loader GLGetProcAddress) (*GLInterface, error) {
	if loader == nil {
		return nil, fmt.Errorf("skia: gpu: nil GL proc loader")
	}
	glProcMu.Lock()
	id := glProcNextID
	glProcNextID++
	glProcRegistry[id] = loader
	glProcMu.Unlock()
	defer func() {
		glProcMu.Lock()
		delete(glProcRegistry, id)
		glProcMu.Unlock()
	}()

	p := C.goskia_assemble_gl(C.int(id))
	if p == nil {
		return nil, fmt.Errorf("skia: gpu: failed to assemble GL interface")
	}
	return newGLInterface(p), nil
}

func newGLInterface(p *C.gr_glinterface_t) *GLInterface {
	gi := &GLInterface{ptr: p}
	runtime.SetFinalizer(gi, (*GLInterface).release)
	return gi
}

func (gi *GLInterface) release() {
	if gi != nil && gi.ptr != nil {
		C.gr_glinterface_unref(gi.ptr)
		gi.ptr = nil
		runtime.SetFinalizer(gi, nil)
	}
}

// Release frees this reference to the GL interface.
func (gi *GLInterface) Release() { gi.release() }

// DirectContext is a GPU context (a connection to the GPU). It is created from
// a backend interface and used to make GPU-backed surfaces.
type DirectContext struct{ ptr *C.gr_direct_context_t }

// NewGLContext creates a GPU context backed by OpenGL using iface (or the
// native interface when iface is nil).
func NewGLContext(iface *GLInterface) (*DirectContext, error) {
	var ip *C.gr_glinterface_t
	if iface != nil {
		ip = iface.ptr
	}
	p := C.gr_direct_context_make_gl(ip)
	runtime.KeepAlive(iface)
	if p == nil {
		return nil, fmt.Errorf("skia: gpu: failed to create GL direct context")
	}
	ctx := &DirectContext{ptr: p}
	runtime.SetFinalizer(ctx, (*DirectContext).release)
	return ctx, nil
}

func (ctx *DirectContext) recording() *C.gr_recording_context_t {
	return (*C.gr_recording_context_t)(unsafe.Pointer(ctx.ptr))
}

// Flush flushes pending GPU work to the backend (but does not submit it).
func (ctx *DirectContext) Flush() { C.gr_direct_context_flush(ctx.ptr); runtime.KeepAlive(ctx) }

// Submit submits flushed work to the GPU. If syncCPU is true it blocks until
// the GPU has finished.
func (ctx *DirectContext) Submit(syncCPU bool) bool {
	ok := bool(C.gr_direct_context_submit(ctx.ptr, C.bool(syncCPU)))
	runtime.KeepAlive(ctx)
	return ok
}

// FlushAndSubmit flushes and submits in one call.
func (ctx *DirectContext) FlushAndSubmit(syncCPU bool) {
	C.gr_direct_context_flush_and_submit(ctx.ptr, C.bool(syncCPU))
	runtime.KeepAlive(ctx)
}

// SetResourceCacheLimit caps the GPU resource cache at maxBytes. Skia evicts the
// least-recently-used resources (textures, glyph atlases, buffers) above this
// limit, bounding GPU-side memory.
func (ctx *DirectContext) SetResourceCacheLimit(maxBytes int) {
	C.gr_direct_context_set_resource_cache_limit(ctx.ptr, C.size_t(maxBytes))
	runtime.KeepAlive(ctx)
}

// ResourceCacheLimit returns the current GPU resource cache byte budget.
func (ctx *DirectContext) ResourceCacheLimit() int {
	n := C.gr_direct_context_get_resource_cache_limit(ctx.ptr)
	runtime.KeepAlive(ctx)
	return int(n)
}

// PurgeUnlockedResources frees GPU resources not currently in use. When
// scratchOnly is true only scratch/temporary resources are released.
func (ctx *DirectContext) PurgeUnlockedResources(scratchOnly bool) {
	C.gr_direct_context_purge_unlocked_resources(ctx.ptr, C.bool(scratchOnly))
	runtime.KeepAlive(ctx)
}

func (ctx *DirectContext) release() {
	if ctx != nil && ctx.ptr != nil {
		C.gr_recording_context_unref(ctx.recording())
		ctx.ptr = nil
		runtime.SetFinalizer(ctx, nil)
	}
}

// Release frees the GPU context.
func (ctx *DirectContext) Release() { ctx.release() }

// NewGPUSurface creates an offscreen GPU-backed surface (a render target
// texture) of the given size and format. sampleCount is the MSAA sample count
// (1 = none).
func NewGPUSurface(ctx *DirectContext, info ImageInfo, sampleCount int, origin SurfaceOrigin) (*Surface, error) {
	if ctx == nil {
		return nil, fmt.Errorf("skia: gpu: nil context")
	}
	ci := info.c()
	p := C.sk_surface_new_render_target(ctx.recording(), C.bool(true), &ci,
		C.int(sampleCount), C.gr_surfaceorigin_t(origin), nil, C.bool(false))
	runtime.KeepAlive(ctx)
	if p == nil {
		return nil, fmt.Errorf("skia: gpu: failed to create %dx%d render target", info.Width, info.Height)
	}
	s := &Surface{ptr: p, info: info}
	runtime.SetFinalizer(s, (*Surface).release)
	return s, nil
}

// NewGPUSurfaceFromFBO wraps an existing OpenGL framebuffer object (for example
// a window's default framebuffer, FBO 0) as a Skia surface. glFormat is a GL
// sized internal format such as GLRGBA8.
func NewGPUSurfaceFromFBO(ctx *DirectContext, fboID uint32, width, height, samples, stencilBits int, glFormat uint32, colorType ColorType, origin SurfaceOrigin) (*Surface, error) {
	if ctx == nil {
		return nil, fmt.Errorf("skia: gpu: nil context")
	}
	fb := C.gr_gl_framebufferinfo_t{fFBOID: C.uint(fboID), fFormat: C.uint(glFormat)}
	target := C.gr_backendrendertarget_new_gl(C.int(width), C.int(height), C.int(samples), C.int(stencilBits), &fb)
	if target == nil {
		return nil, fmt.Errorf("skia: gpu: failed to create backend render target")
	}
	defer C.gr_backendrendertarget_delete(target)

	p := C.sk_surface_new_backend_render_target(ctx.recording(), target,
		C.gr_surfaceorigin_t(origin), C.sk_colortype_t(colorType), nil, nil)
	runtime.KeepAlive(ctx)
	if p == nil {
		return nil, fmt.Errorf("skia: gpu: failed to wrap framebuffer %d", fboID)
	}
	s := &Surface{ptr: p, info: NewImageInfo(width, height, colorType, AlphaTypePremul)}
	runtime.SetFinalizer(s, (*Surface).release)
	return s, nil
}
