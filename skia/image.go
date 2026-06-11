package skia

// #include "goskia.h"
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// Image is an immutable, reference-counted rectangle of pixels (Skia's
// SkImage). It may be raster- or GPU-backed.
type Image struct{ ptr *C.sk_image_t }

func newImage(ptr *C.sk_image_t) *Image {
	if ptr == nil {
		return nil
	}
	img := &Image{ptr: ptr}
	runtime.SetFinalizer(img, (*Image).release)
	return img
}

// DecodeImage decodes encoded image bytes (PNG, JPEG, WEBP, GIF, ...) into an
// immutable raster-backed image.
func DecodeImage(encoded []byte) (*Image, error) {
	if len(encoded) == 0 {
		return nil, fmt.Errorf("skia: DecodeImage: empty input")
	}
	d := C.sk_data_new_with_copy(unsafe.Pointer(&encoded[0]), C.size_t(len(encoded)))
	defer C.sk_data_unref(d)
	img := C.sk_image_new_from_encoded(d)
	if img == nil {
		return nil, fmt.Errorf("skia: DecodeImage: unsupported or corrupt image data")
	}
	return newImage(img), nil
}

// NewImageFromPixels creates a raster image by copying the given pixel buffer.
func NewImageFromPixels(info ImageInfo, pixels []byte, rowBytes int) (*Image, error) {
	if rowBytes == 0 {
		rowBytes = info.MinRowBytes()
	}
	if len(pixels) < rowBytes*info.Height {
		return nil, fmt.Errorf("skia: NewImageFromPixels: buffer too small (%d < %d)", len(pixels), rowBytes*info.Height)
	}
	ci := info.c()
	img := C.sk_image_new_raster_copy(&ci, unsafe.Pointer(&pixels[0]), C.size_t(rowBytes))
	if img == nil {
		return nil, fmt.Errorf("skia: NewImageFromPixels: failed")
	}
	return newImage(img), nil
}

// Width returns the image width in pixels.
func (img *Image) Width() int {
	w := int(C.sk_image_get_width(img.ptr))
	runtime.KeepAlive(img)
	return w
}

// Height returns the image height in pixels.
func (img *Image) Height() int {
	h := int(C.sk_image_get_height(img.ptr))
	runtime.KeepAlive(img)
	return h
}

// ColorType returns the image's pixel color type.
func (img *Image) ColorType() ColorType {
	ct := ColorType(C.sk_image_get_color_type(img.ptr))
	runtime.KeepAlive(img)
	return ct
}

// AlphaType returns the image's alpha type.
func (img *Image) AlphaType() AlphaType {
	at := AlphaType(C.sk_image_get_alpha_type(img.ptr))
	runtime.KeepAlive(img)
	return at
}

// Encode encodes the image. For PNG, quality is ignored; for JPEG and WEBP it
// is the 0..100 quality (a value <= 0 is treated as 100).
func (img *Image) Encode(format EncodedFormat, quality int) ([]byte, error) {
	pm := C.sk_pixmap_new()
	defer C.sk_pixmap_destructor(pm)

	if bool(C.sk_image_peek_pixels(img.ptr, pm)) {
		out, err := encodePixmap(pm, format, quality)
		runtime.KeepAlive(img)
		return out, err
	}

	// Fallback for non-raster (e.g. GPU/lazy) images: read pixels into a
	// C-owned buffer, then encode from a pixmap wrapping it. We use C memory
	// (not a Go slice) because the pixmap retains the pointer across the call.
	w := int(C.sk_image_get_width(img.ptr))
	h := int(C.sk_image_get_height(img.ptr))
	info := NewImageInfo(w, h, ColorTypeRGBA8888, AlphaTypePremul)
	rowBytes := info.MinRowBytes()
	size := rowBytes * h
	buf := C.malloc(C.size_t(size))
	if buf == nil {
		return nil, fmt.Errorf("skia: Encode: out of memory allocating %d bytes", size)
	}
	defer C.free(buf)
	ci := info.c()
	if !bool(C.sk_image_read_pixels(img.ptr, &ci, buf, C.size_t(rowBytes), 0, 0, C.DISALLOW_SK_IMAGE_CACHING_HINT)) {
		runtime.KeepAlive(img)
		return nil, fmt.Errorf("skia: Encode: failed to read pixels from image")
	}
	pm2 := C.sk_pixmap_new_with_params(&ci, buf, C.size_t(rowBytes))
	defer C.sk_pixmap_destructor(pm2)
	out, err := encodePixmap(pm2, format, quality)
	runtime.KeepAlive(img)
	return out, err
}

func (img *Image) release() {
	if img != nil && img.ptr != nil {
		C.sk_image_unref(img.ptr)
		img.ptr = nil
		runtime.SetFinalizer(img, nil)
	}
}

// Release frees this reference to the image. It is safe to call multiple times.
func (img *Image) Release() { img.release() }
