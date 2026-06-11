package skia

// #include "goskia.h"
import "C"

// ImageInfo describes the dimensions and pixel format of an image or surface.
//
// ColorSpace is intentionally omitted from this struct for now; the native
// color space pointer is left null, which Skia treats as an unmanaged
// (legacy/sRGB-unaware) color space. That is correct for most offscreen
// rendering and PNG export.
type ImageInfo struct {
	Width, Height int
	ColorType     ColorType
	AlphaType     AlphaType
}

// NewImageInfo returns an ImageInfo with the given size and pixel format.
func NewImageInfo(width, height int, ct ColorType, at AlphaType) ImageInfo {
	return ImageInfo{Width: width, Height: height, ColorType: ct, AlphaType: at}
}

// ImageInfoN32Premul returns an ImageInfo using a portable 32-bit premultiplied
// RGBA format, suitable for most rendering and image export.
func ImageInfoN32Premul(width, height int) ImageInfo {
	return NewImageInfo(width, height, ColorTypeRGBA8888, AlphaTypePremul)
}

// BytesPerPixel returns the size in bytes of a single pixel.
func (ii ImageInfo) BytesPerPixel() int { return ii.ColorType.BytesPerPixel() }

// MinRowBytes returns the minimum number of bytes for one row of pixels.
func (ii ImageInfo) MinRowBytes() int { return ii.Width * ii.BytesPerPixel() }

// ByteSize returns the minimum number of bytes for the full pixel buffer.
func (ii ImageInfo) ByteSize() int { return ii.MinRowBytes() * ii.Height }

func (ii ImageInfo) c() C.sk_imageinfo_t {
	return C.sk_imageinfo_t{
		colorspace: nil,
		width:      C.int32_t(ii.Width),
		height:     C.int32_t(ii.Height),
		colorType:  C.sk_colortype_t(ii.ColorType),
		alphaType:  C.sk_alphatype_t(ii.AlphaType),
	}
}

func imageInfoFromC(c C.sk_imageinfo_t) ImageInfo {
	return ImageInfo{
		Width:     int(c.width),
		Height:    int(c.height),
		ColorType: ColorType(c.colorType),
		AlphaType: AlphaType(c.alphaType),
	}
}
