package skia

// #include "goskia.h"
import "C"

// Color is a 32-bit color in 0xAARRGGBB order (non-premultiplied), matching
// Skia's SkColor.
type Color uint32

// RGBA builds a Color from 8-bit components.
func RGBA(r, g, b, a uint8) Color {
	return Color(uint32(a)<<24 | uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

// RGB builds an opaque Color from 8-bit components.
func RGB(r, g, b uint8) Color { return RGBA(r, g, b, 0xff) }

// A returns the alpha component.
func (c Color) A() uint8 { return uint8(c >> 24) }

// R returns the red component.
func (c Color) R() uint8 { return uint8(c >> 16) }

// G returns the green component.
func (c Color) G() uint8 { return uint8(c >> 8) }

// B returns the blue component.
func (c Color) B() uint8 { return uint8(c) }

// WithAlpha returns the color with its alpha replaced.
func (c Color) WithAlpha(a uint8) Color { return RGBA(c.R(), c.G(), c.B(), a) }

func (c Color) c() C.sk_color_t { return C.sk_color_t(c) }

// Common opaque colors.
const (
	ColorTransparent Color = 0x00000000
	ColorBlack       Color = 0xff000000
	ColorWhite       Color = 0xffffffff
	ColorRed         Color = 0xffff0000
	ColorGreen       Color = 0xff00ff00
	ColorBlue        Color = 0xff0000ff
	ColorYellow      Color = 0xffffff00
	ColorCyan        Color = 0xff00ffff
	ColorMagenta     Color = 0xffff00ff
	ColorGray        Color = 0xff808080
)

// Color4f is a color with 32-bit float components in RGBA order, unclamped and
// linear-friendly. It maps to Skia's SkColor4f.
type Color4f struct{ R, G, B, A float32 }

func (c Color4f) c() C.sk_color4f_t {
	return C.sk_color4f_t{fR: C.float(c.R), fG: C.float(c.G), fB: C.float(c.B), fA: C.float(c.A)}
}

func color4fFromC(c C.sk_color4f_t) Color4f {
	return Color4f{float32(c.fR), float32(c.fG), float32(c.fB), float32(c.fA)}
}
