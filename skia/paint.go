package skia

// #include "goskia.h"
import "C"

import "runtime"

// Paint holds the style and color used to draw geometry, text, and images.
type Paint struct{ ptr *C.sk_paint_t }

// NewPaint returns a new Paint with default values (black fill, no antialias).
func NewPaint() *Paint {
	p := &Paint{ptr: C.sk_paint_new()}
	runtime.SetFinalizer(p, (*Paint).release)
	return p
}

// Clone returns an independent copy of the paint.
func (p *Paint) Clone() *Paint {
	c := &Paint{ptr: C.sk_paint_clone(p.ptr)}
	runtime.SetFinalizer(c, (*Paint).release)
	runtime.KeepAlive(p)
	return c
}

// Reset restores the paint to its default values.
func (p *Paint) Reset() { C.sk_paint_reset(p.ptr) }

// SetAntialias enables or disables antialiased edges.
func (p *Paint) SetAntialias(aa bool) { C.sk_paint_set_antialias(p.ptr, C.bool(aa)) }

// IsAntialias reports whether antialiasing is enabled.
func (p *Paint) IsAntialias() bool { return bool(C.sk_paint_is_antialias(p.ptr)) }

// SetDither enables or disables dithering.
func (p *Paint) SetDither(d bool) { C.sk_paint_set_dither(p.ptr, C.bool(d)) }

// IsDither reports whether dithering is enabled.
func (p *Paint) IsDither() bool { return bool(C.sk_paint_is_dither(p.ptr)) }

// SetColor sets the paint color (0xAARRGGBB).
func (p *Paint) SetColor(c Color) { C.sk_paint_set_color(p.ptr, c.c()) }

// Color returns the paint color.
func (p *Paint) Color() Color { return Color(C.sk_paint_get_color(p.ptr)) }

// SetColor4f sets the paint color from float components (no color space).
func (p *Paint) SetColor4f(c Color4f) {
	cc := c.c()
	C.sk_paint_set_color4f(p.ptr, &cc, nil)
}

// SetStyle selects fill, stroke, or both.
func (p *Paint) SetStyle(s PaintStyle) { C.sk_paint_set_style(p.ptr, C.sk_paint_style_t(s)) }

// Style returns the paint style.
func (p *Paint) Style() PaintStyle { return PaintStyle(C.sk_paint_get_style(p.ptr)) }

// SetStrokeWidth sets the stroke width in local coordinates. Zero means a
// hairline (one device pixel).
func (p *Paint) SetStrokeWidth(w float32) { C.sk_paint_set_stroke_width(p.ptr, C.float(w)) }

// StrokeWidth returns the stroke width.
func (p *Paint) StrokeWidth() float32 { return float32(C.sk_paint_get_stroke_width(p.ptr)) }

// SetStrokeMiter sets the stroke miter limit.
func (p *Paint) SetStrokeMiter(m float32) { C.sk_paint_set_stroke_miter(p.ptr, C.float(m)) }

// StrokeMiter returns the stroke miter limit.
func (p *Paint) StrokeMiter() float32 { return float32(C.sk_paint_get_stroke_miter(p.ptr)) }

// SetStrokeCap sets the cap drawn at the ends of open strokes.
func (p *Paint) SetStrokeCap(c StrokeCap) { C.sk_paint_set_stroke_cap(p.ptr, C.sk_stroke_cap_t(c)) }

// StrokeCap returns the stroke cap.
func (p *Paint) StrokeCap() StrokeCap { return StrokeCap(C.sk_paint_get_stroke_cap(p.ptr)) }

// SetStrokeJoin sets the join drawn at stroke corners.
func (p *Paint) SetStrokeJoin(j StrokeJoin) { C.sk_paint_set_stroke_join(p.ptr, C.sk_stroke_join_t(j)) }

// StrokeJoin returns the stroke join.
func (p *Paint) StrokeJoin() StrokeJoin { return StrokeJoin(C.sk_paint_get_stroke_join(p.ptr)) }

// SetBlendMode sets the blend mode used to composite drawing onto the canvas.
func (p *Paint) SetBlendMode(m BlendMode) { C.sk_paint_set_blendmode(p.ptr, C.sk_blendmode_t(m)) }

// BlendMode returns the blend mode.
func (p *Paint) BlendMode() BlendMode { return BlendMode(C.sk_paint_get_blendmode(p.ptr)) }

func (p *Paint) release() {
	if p != nil && p.ptr != nil {
		C.sk_paint_delete(p.ptr)
		p.ptr = nil
		runtime.SetFinalizer(p, nil)
	}
}

// Release frees the paint. It is safe to call multiple times.
func (p *Paint) Release() { p.release() }
