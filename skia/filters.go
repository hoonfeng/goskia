package skia

// #include "goskia.h"
import "C"

import "runtime"

// ----------------------------------------------------------------------------
// MaskFilter — operates on the coverage (alpha) mask before it is colored.
// ----------------------------------------------------------------------------

// MaskFilter transforms the alpha mask of a draw (e.g. blurring it for a glow
// or soft shadow). Install one with Paint.SetMaskFilter.
type MaskFilter struct{ ptr *C.sk_maskfilter_t }

func newMaskFilter(p *C.sk_maskfilter_t) *MaskFilter {
	if p == nil {
		return nil
	}
	m := &MaskFilter{ptr: p}
	runtime.SetFinalizer(m, (*MaskFilter).release)
	return m
}

// NewBlurMaskFilter returns a blur mask filter. sigma is the Gaussian standard
// deviation in pixels; a rough rule is sigma ≈ radius/2.
func NewBlurMaskFilter(style BlurStyle, sigma float32) *MaskFilter {
	return newMaskFilter(C.sk_maskfilter_new_blur(C.sk_blurstyle_t(style), C.float(sigma)))
}

func (m *MaskFilter) release() {
	if m != nil && m.ptr != nil {
		C.sk_maskfilter_unref(m.ptr)
		m.ptr = nil
		runtime.SetFinalizer(m, nil)
	}
}

// Release frees this reference to the mask filter.
func (m *MaskFilter) Release() { m.release() }

// SetMaskFilter installs mf on the paint, or clears it when mf is nil.
func (p *Paint) SetMaskFilter(mf *MaskFilter) {
	p.setMaskFilter(mf)
}

func (p *Paint) setMaskFilter(mf *MaskFilter) {
	if mf == nil {
		C.sk_paint_set_maskfilter(p.ptr, nil)
	} else {
		C.sk_paint_set_maskfilter(p.ptr, mf.ptr)
	}
	runtime.KeepAlive(p)
	runtime.KeepAlive(mf)
}

// ----------------------------------------------------------------------------
// ColorFilter — remaps colors per pixel.
// ----------------------------------------------------------------------------

// ColorFilter transforms each color a paint produces (tinting, grayscale,
// matrix transforms, ...). Install one with Paint.SetColorFilter.
type ColorFilter struct{ ptr *C.sk_colorfilter_t }

func newColorFilter(p *C.sk_colorfilter_t) *ColorFilter {
	if p == nil {
		return nil
	}
	cf := &ColorFilter{ptr: p}
	runtime.SetFinalizer(cf, (*ColorFilter).release)
	return cf
}

// NewBlendColorFilter blends the given color into the source using mode.
func NewBlendColorFilter(c Color, mode BlendMode) *ColorFilter {
	return newColorFilter(C.sk_colorfilter_new_mode(c.c(), C.sk_blendmode_t(mode)))
}

// NewLightingColorFilter multiplies by mul and adds add (both as colors).
func NewLightingColorFilter(mul, add Color) *ColorFilter {
	return newColorFilter(C.sk_colorfilter_new_lighting(mul.c(), add.c()))
}

// NewColorMatrixFilter applies a 4x5 row-major color matrix (20 values).
func NewColorMatrixFilter(matrix [20]float32) *ColorFilter {
	var arr [20]C.float
	for i, v := range matrix {
		arr[i] = C.float(v)
	}
	return newColorFilter(C.sk_colorfilter_new_color_matrix(&arr[0]))
}

// NewComposeColorFilter returns outer(inner(color)).
func NewComposeColorFilter(outer, inner *ColorFilter) *ColorFilter {
	cf := newColorFilter(C.sk_colorfilter_new_compose(cfPtr(outer), cfPtr(inner)))
	runtime.KeepAlive(outer)
	runtime.KeepAlive(inner)
	return cf
}

// GrayscaleColorFilter converts colors to luminance-weighted grayscale.
func GrayscaleColorFilter() *ColorFilter {
	const r, g, b = 0.2126, 0.7152, 0.0722
	return NewColorMatrixFilter([20]float32{
		r, g, b, 0, 0,
		r, g, b, 0, 0,
		r, g, b, 0, 0,
		0, 0, 0, 1, 0,
	})
}

func cfPtr(cf *ColorFilter) *C.sk_colorfilter_t {
	if cf == nil {
		return nil
	}
	return cf.ptr
}

func (cf *ColorFilter) release() {
	if cf != nil && cf.ptr != nil {
		C.sk_colorfilter_unref(cf.ptr)
		cf.ptr = nil
		runtime.SetFinalizer(cf, nil)
	}
}

// Release frees this reference to the color filter.
func (cf *ColorFilter) Release() { cf.release() }

// SetColorFilter installs cf on the paint, or clears it when cf is nil.
func (p *Paint) SetColorFilter(cf *ColorFilter) {
	C.sk_paint_set_colorfilter(p.ptr, cfPtr(cf))
	runtime.KeepAlive(p)
	runtime.KeepAlive(cf)
}

// ----------------------------------------------------------------------------
// ImageFilter — operates on the rasterized result (blur, drop shadow, ...).
// ----------------------------------------------------------------------------

// ImageFilter processes the pixels a draw produces. Filters can be chained via
// their input parameter (nil means the draw itself). Install one with
// Paint.SetImageFilter.
type ImageFilter struct{ ptr *C.sk_imagefilter_t }

func newImageFilter(p *C.sk_imagefilter_t) *ImageFilter {
	if p == nil {
		return nil
	}
	f := &ImageFilter{ptr: p}
	runtime.SetFinalizer(f, (*ImageFilter).release)
	return f
}

func ifPtr(f *ImageFilter) *C.sk_imagefilter_t {
	if f == nil {
		return nil
	}
	return f.ptr
}

// NewBlurImageFilter blurs its input by the given Gaussian sigmas.
func NewBlurImageFilter(sigmaX, sigmaY float32, tile TileMode, input *ImageFilter) *ImageFilter {
	f := newImageFilter(C.sk_imagefilter_new_blur(C.float(sigmaX), C.float(sigmaY), C.sk_shader_tilemode_t(tile), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewDropShadowImageFilter draws input plus a drop shadow offset by (dx, dy),
// blurred by the given sigmas and tinted color.
func NewDropShadowImageFilter(dx, dy, sigmaX, sigmaY float32, color Color, input *ImageFilter) *ImageFilter {
	f := newImageFilter(C.sk_imagefilter_new_drop_shadow(C.float(dx), C.float(dy), C.float(sigmaX), C.float(sigmaY), color.c(), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewDropShadowOnlyImageFilter produces only the shadow (not the input).
func NewDropShadowOnlyImageFilter(dx, dy, sigmaX, sigmaY float32, color Color, input *ImageFilter) *ImageFilter {
	f := newImageFilter(C.sk_imagefilter_new_drop_shadow_only(C.float(dx), C.float(dy), C.float(sigmaX), C.float(sigmaY), color.c(), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewOffsetImageFilter translates its input by (dx, dy).
func NewOffsetImageFilter(dx, dy float32, input *ImageFilter) *ImageFilter {
	f := newImageFilter(C.sk_imagefilter_new_offset(C.float(dx), C.float(dy), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewColorFilterImageFilter applies a ColorFilter as an image filter.
func NewColorFilterImageFilter(cf *ColorFilter, input *ImageFilter) *ImageFilter {
	f := newImageFilter(C.sk_imagefilter_new_color_filter(cfPtr(cf), ifPtr(input), nil))
	runtime.KeepAlive(cf)
	runtime.KeepAlive(input)
	return f
}

// NewComposeImageFilter returns outer(inner(...)).
func NewComposeImageFilter(outer, inner *ImageFilter) *ImageFilter {
	f := newImageFilter(C.sk_imagefilter_new_compose(ifPtr(outer), ifPtr(inner)))
	runtime.KeepAlive(outer)
	runtime.KeepAlive(inner)
	return f
}

// Lighting image filters treat the alpha of their input as a height field and
// shade it with a light, producing embossed / 3-D effects. surfaceScale scales
// the bump height; kd/ks are the diffuse/specular reflectance constants.

// NewDistantLitDiffuseImageFilter shades input with a distant (directional)
// light using the diffuse model.
func NewDistantLitDiffuseImageFilter(direction Point3, light Color, surfaceScale, kd float32, input *ImageFilter) *ImageFilter {
	d := direction.c()
	f := newImageFilter(C.sk_imagefilter_new_distant_lit_diffuse(&d, light.c(), C.float(surfaceScale), C.float(kd), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewPointLitDiffuseImageFilter shades input with a point light (diffuse).
func NewPointLitDiffuseImageFilter(location Point3, light Color, surfaceScale, kd float32, input *ImageFilter) *ImageFilter {
	l := location.c()
	f := newImageFilter(C.sk_imagefilter_new_point_lit_diffuse(&l, light.c(), C.float(surfaceScale), C.float(kd), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewSpotLitDiffuseImageFilter shades input with a spot light (diffuse).
// cutoffAngle is in degrees.
func NewSpotLitDiffuseImageFilter(location, target Point3, specularExponent, cutoffAngle float32, light Color, surfaceScale, kd float32, input *ImageFilter) *ImageFilter {
	l, t := location.c(), target.c()
	f := newImageFilter(C.sk_imagefilter_new_spot_lit_diffuse(&l, &t, C.float(specularExponent), C.float(cutoffAngle), light.c(), C.float(surfaceScale), C.float(kd), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewDistantLitSpecularImageFilter shades input with a distant light (specular).
func NewDistantLitSpecularImageFilter(direction Point3, light Color, surfaceScale, ks, shininess float32, input *ImageFilter) *ImageFilter {
	d := direction.c()
	f := newImageFilter(C.sk_imagefilter_new_distant_lit_specular(&d, light.c(), C.float(surfaceScale), C.float(ks), C.float(shininess), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewPointLitSpecularImageFilter shades input with a point light (specular).
func NewPointLitSpecularImageFilter(location Point3, light Color, surfaceScale, ks, shininess float32, input *ImageFilter) *ImageFilter {
	l := location.c()
	f := newImageFilter(C.sk_imagefilter_new_point_lit_specular(&l, light.c(), C.float(surfaceScale), C.float(ks), C.float(shininess), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

// NewSpotLitSpecularImageFilter shades input with a spot light (specular).
// cutoffAngle is in degrees.
func NewSpotLitSpecularImageFilter(location, target Point3, specularExponent, cutoffAngle float32, light Color, surfaceScale, ks, shininess float32, input *ImageFilter) *ImageFilter {
	l, t := location.c(), target.c()
	f := newImageFilter(C.sk_imagefilter_new_spot_lit_specular(&l, &t, C.float(specularExponent), C.float(cutoffAngle), light.c(), C.float(surfaceScale), C.float(ks), C.float(shininess), ifPtr(input), nil))
	runtime.KeepAlive(input)
	return f
}

func (f *ImageFilter) release() {
	if f != nil && f.ptr != nil {
		C.sk_imagefilter_unref(f.ptr)
		f.ptr = nil
		runtime.SetFinalizer(f, nil)
	}
}

// Release frees this reference to the image filter.
func (f *ImageFilter) Release() { f.release() }

// SetImageFilter installs f on the paint, or clears it when f is nil.
func (p *Paint) SetImageFilter(f *ImageFilter) {
	C.sk_paint_set_imagefilter(p.ptr, ifPtr(f))
	runtime.KeepAlive(p)
	runtime.KeepAlive(f)
}

// ----------------------------------------------------------------------------
// PathEffect — alters geometry of stroked/filled paths.
// ----------------------------------------------------------------------------

// PathEffect modifies a path's geometry before it is filled or stroked
// (dashing, rounding corners, ...). Install one with Paint.SetPathEffect.
type PathEffect struct{ ptr *C.sk_path_effect_t }

func newPathEffect(p *C.sk_path_effect_t) *PathEffect {
	if p == nil {
		return nil
	}
	e := &PathEffect{ptr: p}
	runtime.SetFinalizer(e, (*PathEffect).release)
	return e
}

func pePtr(e *PathEffect) *C.sk_path_effect_t {
	if e == nil {
		return nil
	}
	return e.ptr
}

// NewDashPathEffect dashes the stroke using on/off interval lengths (an even
// count, >= 2). phase offsets the start of the pattern.
func NewDashPathEffect(intervals []float32, phase float32) *PathEffect {
	if len(intervals) < 2 || len(intervals)%2 != 0 {
		return nil
	}
	arr := make([]C.float, len(intervals))
	for i, v := range intervals {
		arr[i] = C.float(v)
	}
	return newPathEffect(C.sk_path_effect_create_dash(&arr[0], C.int(len(intervals)), C.float(phase)))
}

// NewCornerPathEffect rounds sharp corners with the given radius.
func NewCornerPathEffect(radius float32) *PathEffect {
	return newPathEffect(C.sk_path_effect_create_corner(C.float(radius)))
}

// NewDiscretePathEffect roughens a path into random segments (a "sketchy" look).
func NewDiscretePathEffect(segLength, deviation float32) *PathEffect {
	return newPathEffect(C.sk_path_effect_create_discrete(C.float(segLength), C.float(deviation), 0))
}

// NewTrimPathEffect keeps the portion of the path in [start, stop] (each 0..1).
func NewTrimPathEffect(start, stop float32, mode TrimMode) *PathEffect {
	return newPathEffect(C.sk_path_effect_create_trim(C.float(start), C.float(stop), C.sk_path_effect_trim_mode_t(mode)))
}

// NewComposePathEffect returns outer(inner(path)).
func NewComposePathEffect(outer, inner *PathEffect) *PathEffect {
	e := newPathEffect(C.sk_path_effect_create_compose(pePtr(outer), pePtr(inner)))
	runtime.KeepAlive(outer)
	runtime.KeepAlive(inner)
	return e
}

func (e *PathEffect) release() {
	if e != nil && e.ptr != nil {
		C.sk_path_effect_unref(e.ptr)
		e.ptr = nil
		runtime.SetFinalizer(e, nil)
	}
}

// Release frees this reference to the path effect.
func (e *PathEffect) Release() { e.release() }

// SetPathEffect installs e on the paint, or clears it when e is nil.
func (p *Paint) SetPathEffect(e *PathEffect) {
	C.sk_paint_set_path_effect(p.ptr, pePtr(e))
	runtime.KeepAlive(p)
	runtime.KeepAlive(e)
}
