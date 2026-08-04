package skia

// #include "goskia.h"
import "C"

import "runtime"

// RRect describes a rounded rectangle: a rectangle with optionally
// non-uniform corner radii. Skia's standard rounded-rect primitive.
//
// RRect is the native Skia type used by ClipRRect and DrawRRect — the GPU
// backend has a dedicated optimized clip path for RRects (skgpu::ClipStack
// treats them analytically), which is far more reliable than path-based
// clipping.
type RRect struct{ ptr *C.sk_rrect_t }

// NewRRect returns a new, empty rounded rectangle (type Empty).
func NewRRect() *RRect {
	r := &RRect{ptr: C.sk_rrect_new()}
	runtime.SetFinalizer(r, (*RRect).release)
	return r
}

// release frees the underlying Skia object.
func (r *RRect) release() {
	if r.ptr != nil {
		C.sk_rrect_delete(r.ptr)
		r.ptr = nil
	}
}

// Release explicitly frees the underlying Skia object. The receiver must not
// be used afterwards.
func (r *RRect) Release() { r.release() }

// SetRectXY sets the rounded rectangle to the given bounds with uniform
// corner radii (rx, ry). Radii are clamped to the half-width/height, matching
// CSS border-radius clamping semantics.
func (r *RRect) SetRectXY(rect Rect, rx, ry float32) *RRect {
	cr := rect.c()
	C.sk_rrect_set_rect_xy(r.ptr, &cr, C.float(rx), C.float(ry))
	return r
}
