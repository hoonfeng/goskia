package skia

// #include "goskia.h"
import "C"

import (
	"fmt"
	"runtime"
)

// Surface is the destination for drawing. A raster surface is backed by a
// pixel buffer in memory; a GPU surface (see the gpu APIs) is backed by a
// texture or render target.
type Surface struct {
	ptr    *C.sk_surface_t
	info   ImageInfo
	canvas *Canvas
}

// NewRasterSurface creates a CPU surface backed by a freshly allocated,
// zero-initialized pixel buffer described by info.
func NewRasterSurface(info ImageInfo) (*Surface, error) {
	if info.Width <= 0 || info.Height <= 0 {
		return nil, fmt.Errorf("skia: invalid surface size %dx%d", info.Width, info.Height)
	}
	ci := info.c()
	ptr := C.sk_surface_new_raster(&ci, 0, nil)
	if ptr == nil {
		return nil, fmt.Errorf("skia: failed to create %dx%d raster surface (colorType=%d alphaType=%d)",
			info.Width, info.Height, info.ColorType, info.AlphaType)
	}
	s := &Surface{ptr: ptr, info: info}
	runtime.SetFinalizer(s, (*Surface).release)
	return s, nil
}

// NewRasterSurfaceN32Premul is a convenience for a width x height surface using
// a portable premultiplied RGBA pixel format.
func NewRasterSurfaceN32Premul(width, height int) (*Surface, error) {
	return NewRasterSurface(ImageInfoN32Premul(width, height))
}

// Width returns the surface width in pixels.
func (s *Surface) Width() int { return s.info.Width }

// Height returns the surface height in pixels.
func (s *Surface) Height() int { return s.info.Height }

// Canvas returns the canvas that draws into this surface. The returned canvas
// is owned by the surface and must not be released; it stays valid for the
// lifetime of the surface.
func (s *Surface) Canvas() *Canvas {
	if s.canvas == nil {
		s.canvas = &Canvas{ptr: C.sk_surface_get_canvas(s.ptr), owner: s}
	}
	return s.canvas
}

// Snapshot returns an immutable image of the surface's current contents.
func (s *Surface) Snapshot() *Image {
	img := newImage(C.sk_surface_new_image_snapshot(s.ptr))
	runtime.KeepAlive(s)
	return img
}

// EncodeToBytes snapshots the surface and encodes it. For PNG, quality is
// ignored; for JPEG/WEBP it is the 0..100 quality (<=0 means 100).
func (s *Surface) EncodeToBytes(format EncodedFormat, quality int) ([]byte, error) {
	img := s.Snapshot()
	if img == nil {
		return nil, fmt.Errorf("skia: surface snapshot failed")
	}
	defer img.Release()
	return img.Encode(format, quality)
}

// EncodePNG snapshots the surface and encodes it as PNG.
func (s *Surface) EncodePNG() ([]byte, error) { return s.EncodeToBytes(EncodedFormatPNG, 100) }

func (s *Surface) release() {
	if s != nil && s.ptr != nil {
		C.sk_surface_unref(s.ptr)
		s.ptr = nil
		s.canvas = nil
		runtime.SetFinalizer(s, nil)
	}
}

// Release frees this reference to the surface. It is safe to call multiple
// times. Any canvas obtained from this surface becomes invalid afterwards.
func (s *Surface) Release() { s.release() }
