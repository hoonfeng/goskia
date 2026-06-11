package skia

// #include "goskia.h"
import "C"

// ColorType describes how pixels are laid out in memory.
type ColorType int32

const (
	ColorTypeUnknown     ColorType = C.UNKNOWN_SK_COLORTYPE
	ColorTypeAlpha8      ColorType = C.ALPHA_8_SK_COLORTYPE
	ColorTypeRGB565      ColorType = C.RGB_565_SK_COLORTYPE
	ColorTypeARGB4444    ColorType = C.ARGB_4444_SK_COLORTYPE
	ColorTypeRGBA8888    ColorType = C.RGBA_8888_SK_COLORTYPE
	ColorTypeRGB888x     ColorType = C.RGB_888X_SK_COLORTYPE
	ColorTypeBGRA8888    ColorType = C.BGRA_8888_SK_COLORTYPE
	ColorTypeRGBA1010102 ColorType = C.RGBA_1010102_SK_COLORTYPE
	ColorTypeBGRA1010102 ColorType = C.BGRA_1010102_SK_COLORTYPE
	ColorTypeGray8       ColorType = C.GRAY_8_SK_COLORTYPE
	ColorTypeRGBAF16     ColorType = C.RGBA_F16_SK_COLORTYPE
	ColorTypeRGBAF32     ColorType = C.RGBA_F32_SK_COLORTYPE
)

// BytesPerPixel returns the number of bytes used by a single pixel of this type.
func (ct ColorType) BytesPerPixel() int {
	switch ct {
	case ColorTypeAlpha8, ColorTypeGray8:
		return 1
	case ColorTypeRGB565, ColorTypeARGB4444:
		return 2
	case ColorTypeRGBA8888, ColorTypeBGRA8888, ColorTypeRGB888x,
		ColorTypeRGBA1010102, ColorTypeBGRA1010102:
		return 4
	case ColorTypeRGBAF16:
		return 8
	case ColorTypeRGBAF32:
		return 16
	default:
		return 0
	}
}

// AlphaType describes how to interpret the alpha component of a pixel.
type AlphaType int32

const (
	AlphaTypeUnknown  AlphaType = C.UNKNOWN_SK_ALPHATYPE
	AlphaTypeOpaque   AlphaType = C.OPAQUE_SK_ALPHATYPE
	AlphaTypePremul   AlphaType = C.PREMUL_SK_ALPHATYPE
	AlphaTypeUnpremul AlphaType = C.UNPREMUL_SK_ALPHATYPE
)

// BlendMode is a Porter-Duff or separable blend mode.
type BlendMode int32

const (
	BlendModeClear      BlendMode = C.CLEAR_SK_BLENDMODE
	BlendModeSrc        BlendMode = C.SRC_SK_BLENDMODE
	BlendModeDst        BlendMode = C.DST_SK_BLENDMODE
	BlendModeSrcOver    BlendMode = C.SRCOVER_SK_BLENDMODE
	BlendModeDstOver    BlendMode = C.DSTOVER_SK_BLENDMODE
	BlendModeSrcIn      BlendMode = C.SRCIN_SK_BLENDMODE
	BlendModeDstIn      BlendMode = C.DSTIN_SK_BLENDMODE
	BlendModeSrcOut     BlendMode = C.SRCOUT_SK_BLENDMODE
	BlendModeDstOut     BlendMode = C.DSTOUT_SK_BLENDMODE
	BlendModeSrcATop    BlendMode = C.SRCATOP_SK_BLENDMODE
	BlendModeDstATop    BlendMode = C.DSTATOP_SK_BLENDMODE
	BlendModeXor        BlendMode = C.XOR_SK_BLENDMODE
	BlendModePlus       BlendMode = C.PLUS_SK_BLENDMODE
	BlendModeModulate   BlendMode = C.MODULATE_SK_BLENDMODE
	BlendModeScreen     BlendMode = C.SCREEN_SK_BLENDMODE
	BlendModeOverlay    BlendMode = C.OVERLAY_SK_BLENDMODE
	BlendModeDarken     BlendMode = C.DARKEN_SK_BLENDMODE
	BlendModeLighten    BlendMode = C.LIGHTEN_SK_BLENDMODE
	BlendModeColorDodge BlendMode = C.COLORDODGE_SK_BLENDMODE
	BlendModeColorBurn  BlendMode = C.COLORBURN_SK_BLENDMODE
	BlendModeHardLight  BlendMode = C.HARDLIGHT_SK_BLENDMODE
	BlendModeSoftLight  BlendMode = C.SOFTLIGHT_SK_BLENDMODE
	BlendModeDifference BlendMode = C.DIFFERENCE_SK_BLENDMODE
	BlendModeExclusion  BlendMode = C.EXCLUSION_SK_BLENDMODE
	BlendModeMultiply   BlendMode = C.MULTIPLY_SK_BLENDMODE
	BlendModeHue        BlendMode = C.HUE_SK_BLENDMODE
	BlendModeSaturation BlendMode = C.SATURATION_SK_BLENDMODE
	BlendModeColor      BlendMode = C.COLOR_SK_BLENDMODE
	BlendModeLuminosity BlendMode = C.LUMINOSITY_SK_BLENDMODE
)

// PaintStyle selects whether geometry is filled, stroked, or both.
type PaintStyle int32

const (
	PaintStyleFill          PaintStyle = C.FILL_SK_PAINT_STYLE
	PaintStyleStroke        PaintStyle = C.STROKE_SK_PAINT_STYLE
	PaintStyleStrokeAndFill PaintStyle = C.STROKE_AND_FILL_SK_PAINT_STYLE
)

// StrokeCap is the geometry drawn at the beginning and end of strokes.
type StrokeCap int32

const (
	StrokeCapButt   StrokeCap = C.BUTT_SK_STROKE_CAP
	StrokeCapRound  StrokeCap = C.ROUND_SK_STROKE_CAP
	StrokeCapSquare StrokeCap = C.SQUARE_SK_STROKE_CAP
)

// StrokeJoin is the geometry drawn at corners of strokes.
type StrokeJoin int32

const (
	StrokeJoinMiter StrokeJoin = C.MITER_SK_STROKE_JOIN
	StrokeJoinRound StrokeJoin = C.ROUND_SK_STROKE_JOIN
	StrokeJoinBevel StrokeJoin = C.BEVEL_SK_STROKE_JOIN
)

// FillType controls how a path's interior is computed.
type FillType int32

const (
	FillTypeWinding        FillType = C.WINDING_SK_PATH_FILLTYPE
	FillTypeEvenOdd        FillType = C.EVENODD_SK_PATH_FILLTYPE
	FillTypeInverseWinding FillType = C.INVERSE_WINDING_SK_PATH_FILLTYPE
	FillTypeInverseEvenOdd FillType = C.INVERSE_EVENODD_SK_PATH_FILLTYPE
)

// PathDirection is the winding direction used when adding shapes to a path.
type PathDirection int32

const (
	PathDirectionCW  PathDirection = C.CW_SK_PATH_DIRECTION
	PathDirectionCCW PathDirection = C.CCW_SK_PATH_DIRECTION
)

// ArcSize selects the smaller or larger arc for sk arc-to operations.
type ArcSize int32

const (
	ArcSizeSmall ArcSize = C.SMALL_SK_PATH_ARC_SIZE
	ArcSizeLarge ArcSize = C.LARGE_SK_PATH_ARC_SIZE
)

// PathAddMode controls how one path is appended to another.
type PathAddMode int32

const (
	PathAddModeAppend PathAddMode = C.APPEND_SK_PATH_ADD_MODE
	PathAddModeExtend PathAddMode = C.EXTEND_SK_PATH_ADD_MODE
)

// TileMode describes how a shader fills the space outside its primary bounds.
type TileMode int32

const (
	TileModeClamp  TileMode = C.CLAMP_SK_SHADER_TILEMODE
	TileModeRepeat TileMode = C.REPEAT_SK_SHADER_TILEMODE
	TileModeMirror TileMode = C.MIRROR_SK_SHADER_TILEMODE
	TileModeDecal  TileMode = C.DECAL_SK_SHADER_TILEMODE
)

// FilterMode selects nearest or linear sampling.
type FilterMode int32

const (
	FilterModeNearest FilterMode = C.NEAREST_SK_FILTER_MODE
	FilterModeLinear  FilterMode = C.LINEAR_SK_FILTER_MODE
)

// MipmapMode selects how mipmap levels are sampled.
type MipmapMode int32

const (
	MipmapModeNone    MipmapMode = C.NONE_SK_MIPMAP_MODE
	MipmapModeNearest MipmapMode = C.NEAREST_SK_MIPMAP_MODE
	MipmapModeLinear  MipmapMode = C.LINEAR_SK_MIPMAP_MODE
)

// PointMode selects how sk_canvas DrawPoints interprets its points.
type PointMode int32

const (
	PointModePoints  PointMode = C.POINTS_SK_POINT_MODE
	PointModeLines   PointMode = C.LINES_SK_POINT_MODE
	PointModePolygon PointMode = C.POLYGON_SK_POINT_MODE
)

// ClipOp selects how a clip region is combined with the current clip.
type ClipOp int32

const (
	ClipOpDifference ClipOp = C.DIFFERENCE_SK_CLIPOP
	ClipOpIntersect  ClipOp = C.INTERSECT_SK_CLIPOP
)

// TextEncoding describes how text bytes map to glyphs.
type TextEncoding int32

const (
	TextEncodingUTF8    TextEncoding = C.UTF8_SK_TEXT_ENCODING
	TextEncodingUTF16   TextEncoding = C.UTF16_SK_TEXT_ENCODING
	TextEncodingUTF32   TextEncoding = C.UTF32_SK_TEXT_ENCODING
	TextEncodingGlyphID TextEncoding = C.GLYPH_ID_SK_TEXT_ENCODING
)

// FontHinting controls glyph outline adjustment to the pixel grid.
type FontHinting int32

const (
	FontHintingNone   FontHinting = C.NONE_SK_FONT_HINTING
	FontHintingSlight FontHinting = C.SLIGHT_SK_FONT_HINTING
	FontHintingNormal FontHinting = C.NORMAL_SK_FONT_HINTING
	FontHintingFull   FontHinting = C.FULL_SK_FONT_HINTING
)

// FontEdging controls antialiasing of glyph edges.
type FontEdging int32

const (
	FontEdgingAlias            FontEdging = C.ALIAS_SK_FONT_EDGING
	FontEdgingAntialias        FontEdging = C.ANTIALIAS_SK_FONT_EDGING
	FontEdgingSubpixelAntalias FontEdging = C.SUBPIXEL_ANTIALIAS_SK_FONT_EDGING
)

// FontSlant is the slant of a font style.
type FontSlant int32

const (
	FontSlantUpright FontSlant = C.UPRIGHT_SK_FONT_STYLE_SLANT
	FontSlantItalic  FontSlant = C.ITALIC_SK_FONT_STYLE_SLANT
	FontSlantOblique FontSlant = C.OBLIQUE_SK_FONT_STYLE_SLANT
)

// PixelGeometry describes the sub-pixel layout of a surface, for LCD text.
type PixelGeometry int32

const (
	PixelGeometryUnknown PixelGeometry = C.UNKNOWN_SK_PIXELGEOMETRY
	PixelGeometryRGBH    PixelGeometry = C.RGB_H_SK_PIXELGEOMETRY
	PixelGeometryBGRH    PixelGeometry = C.BGR_H_SK_PIXELGEOMETRY
	PixelGeometryRGBV    PixelGeometry = C.RGB_V_SK_PIXELGEOMETRY
	PixelGeometryBGRV    PixelGeometry = C.BGR_V_SK_PIXELGEOMETRY
)

// BlurStyle selects how a blur mask filter is applied.
type BlurStyle int32

const (
	BlurStyleNormal BlurStyle = C.NORMAL_SK_BLUR_STYLE
	BlurStyleSolid  BlurStyle = C.SOLID_SK_BLUR_STYLE
	BlurStyleOuter  BlurStyle = C.OUTER_SK_BLUR_STYLE
	BlurStyleInner  BlurStyle = C.INNER_SK_BLUR_STYLE
)

// EncodedFormat identifies an encoded image container format.
type EncodedFormat int32

const (
	EncodedFormatBMP    EncodedFormat = C.BMP_SK_ENCODED_FORMAT
	EncodedFormatGIF    EncodedFormat = C.GIF_SK_ENCODED_FORMAT
	EncodedFormatICO    EncodedFormat = C.ICO_SK_ENCODED_FORMAT
	EncodedFormatJPEG   EncodedFormat = C.JPEG_SK_ENCODED_FORMAT
	EncodedFormatPNG    EncodedFormat = C.PNG_SK_ENCODED_FORMAT
	EncodedFormatWBMP   EncodedFormat = C.WBMP_SK_ENCODED_FORMAT
	EncodedFormatWEBP   EncodedFormat = C.WEBP_SK_ENCODED_FORMAT
	EncodedFormatDNG    EncodedFormat = C.DNG_SK_ENCODED_FORMAT
	EncodedFormatHEIF   EncodedFormat = C.HEIF_SK_ENCODED_FORMAT
	EncodedFormatAVIF   EncodedFormat = C.AVIF_SK_ENCODED_FORMAT
	EncodedFormatJPEGXL EncodedFormat = C.JPEGXL_SK_ENCODED_FORMAT
)

// PathOp is a boolean operation between two paths.
type PathOp int32

const (
	PathOpDifference        PathOp = C.DIFFERENCE_SK_PATHOP
	PathOpIntersect         PathOp = C.INTERSECT_SK_PATHOP
	PathOpUnion             PathOp = C.UNION_SK_PATHOP
	PathOpXor               PathOp = C.XOR_SK_PATHOP
	PathOpReverseDifference PathOp = C.REVERSE_DIFFERENCE_SK_PATHOP
)

// TrimMode selects which portion of a path a trim path effect keeps.
type TrimMode int32

const (
	TrimModeNormal   TrimMode = C.NORMAL_SK_PATH_EFFECT_TRIM_MODE
	TrimModeInverted TrimMode = C.INVERTED_SK_PATH_EFFECT_TRIM_MODE
)

// SurfaceOrigin is the origin convention for GPU-backed surfaces.
type SurfaceOrigin int32

const (
	SurfaceOriginTopLeft    SurfaceOrigin = C.TOP_LEFT_GR_SURFACE_ORIGIN
	SurfaceOriginBottomLeft SurfaceOrigin = C.BOTTOM_LEFT_GR_SURFACE_ORIGIN
)
