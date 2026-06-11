package skia

// #include "goskia.h"
import "C"

// SamplingOptions controls how images are sampled when drawn at a different
// scale or with a transform.
type SamplingOptions struct {
	Filter FilterMode
	Mipmap MipmapMode

	// UseCubic, when true, selects bicubic resampling using CubicB and CubicC
	// (the Mitchell-Netravali B and C parameters), ignoring Filter and Mipmap.
	UseCubic       bool
	CubicB, CubicC float32

	// MaxAniso, when > 1, enables anisotropic filtering.
	MaxAniso int
}

// Common sampling presets.
var (
	// SamplingNearest does nearest-neighbour sampling (sharp, pixelated).
	SamplingNearest = SamplingOptions{Filter: FilterModeNearest, Mipmap: MipmapModeNone}
	// SamplingLinear does bilinear sampling (smooth).
	SamplingLinear = SamplingOptions{Filter: FilterModeLinear, Mipmap: MipmapModeNone}
	// SamplingLinearMipmap does trilinear sampling for minification.
	SamplingLinearMipmap = SamplingOptions{Filter: FilterModeLinear, Mipmap: MipmapModeLinear}
	// SamplingMitchell does high-quality bicubic resampling.
	SamplingMitchell = SamplingOptions{UseCubic: true, CubicB: 1.0 / 3.0, CubicC: 1.0 / 3.0}
)

func (s SamplingOptions) c() C.sk_sampling_options_t {
	return C.sk_sampling_options_t{
		fMaxAniso: C.int(s.MaxAniso),
		fUseCubic: C.bool(s.UseCubic),
		fCubic:    C.sk_cubic_resampler_t{fB: C.float(s.CubicB), fC: C.float(s.CubicC)},
		fFilter:   C.sk_filter_mode_t(s.Filter),
		fMipmap:   C.sk_mipmap_mode_t(s.Mipmap),
	}
}
