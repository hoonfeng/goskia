package skia

// #include "goskia.h"
import "C"

import "runtime"

// Shader produces the colors used to fill geometry (gradients, images, noise,
// solid colors). Install one on a Paint with SetShader.
type Shader struct{ ptr *C.sk_shader_t }

func newShader(ptr *C.sk_shader_t) *Shader {
	if ptr == nil {
		return nil
	}
	s := &Shader{ptr: ptr}
	runtime.SetFinalizer(s, (*Shader).release)
	return s
}

// NewColorShader returns a shader that paints a single solid color.
func NewColorShader(c Color) *Shader {
	return newShader(C.sk_shader_new_color(c.c()))
}

// NewLinearGradient returns a linear gradient from start to end.
//
// colors must have at least two entries. positions, if non-nil, must have the
// same length as colors and contain increasing values in [0,1]; pass nil for
// an even distribution.
func NewLinearGradient(start, end Point, colors []Color, positions []float32, tile TileMode) *Shader {
	ccolors, cpos, n, ok := gradientArgs(colors, positions)
	if !ok {
		return nil
	}
	pts := [2]C.sk_point_t{start.c(), end.c()}
	return newShader(C.sk_shader_new_linear_gradient(&pts[0], &ccolors[0], cpos, n, C.sk_shader_tilemode_t(tile), nil))
}

// NewRadialGradient returns a radial gradient centered at center.
func NewRadialGradient(center Point, radius float32, colors []Color, positions []float32, tile TileMode) *Shader {
	ccolors, cpos, n, ok := gradientArgs(colors, positions)
	if !ok {
		return nil
	}
	cc := center.c()
	return newShader(C.sk_shader_new_radial_gradient(&cc, C.float(radius), &ccolors[0], cpos, n, C.sk_shader_tilemode_t(tile), nil))
}

// NewSweepGradient returns a sweep (angular) gradient around center, spanning
// startAngle..endAngle degrees.
func NewSweepGradient(center Point, colors []Color, positions []float32, startAngle, endAngle float32, tile TileMode) *Shader {
	ccolors, cpos, n, ok := gradientArgs(colors, positions)
	if !ok {
		return nil
	}
	cc := center.c()
	return newShader(C.sk_shader_new_sweep_gradient(&cc, &ccolors[0], cpos, n, C.sk_shader_tilemode_t(tile), C.float(startAngle), C.float(endAngle), nil))
}

// gradientArgs validates and converts the color/position slices. It returns the
// C color slice, a pointer to the C position array (or nil), the count, and ok.
func gradientArgs(colors []Color, positions []float32) ([]C.sk_color_t, *C.float, C.int, bool) {
	if len(colors) < 2 {
		return nil, nil, 0, false
	}
	if positions != nil && len(positions) != len(colors) {
		return nil, nil, 0, false
	}
	ccolors := make([]C.sk_color_t, len(colors))
	for i, c := range colors {
		ccolors[i] = c.c()
	}
	var cpos *C.float
	if positions != nil {
		cp := make([]C.float, len(positions))
		for i, p := range positions {
			cp[i] = C.float(p)
		}
		cpos = &cp[0]
	}
	return ccolors, cpos, C.int(len(colors)), true
}

func (s *Shader) release() {
	if s != nil && s.ptr != nil {
		C.sk_shader_unref(s.ptr)
		s.ptr = nil
		runtime.SetFinalizer(s, nil)
	}
}

// Release frees this reference to the shader. It is safe to call multiple times.
func (s *Shader) Release() { s.release() }

// --- Paint integration ------------------------------------------------------

// SetShader installs sh as the paint's shader, or clears it when sh is nil. The
// paint takes its own reference, so the Shader may be released afterwards.
func (p *Paint) SetShader(sh *Shader) {
	if sh == nil {
		C.sk_paint_set_shader(p.ptr, nil)
	} else {
		C.sk_paint_set_shader(p.ptr, sh.ptr)
	}
	runtime.KeepAlive(p)
	runtime.KeepAlive(sh)
}
