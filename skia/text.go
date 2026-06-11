package skia

// #include "goskia.h"
import "C"

import (
	"runtime"
	"unsafe"
)

// FontStyle describes a typeface's weight, width, and slant.
type FontStyle struct {
	Weight int // 100..900; 400 = normal, 700 = bold
	Width  int // 1..9; 5 = normal
	Slant  FontSlant
}

// Common font styles.
var (
	FontStyleNormal     = FontStyle{400, 5, FontSlantUpright}
	FontStyleBold       = FontStyle{700, 5, FontSlantUpright}
	FontStyleItalic     = FontStyle{400, 5, FontSlantItalic}
	FontStyleBoldItalic = FontStyle{700, 5, FontSlantItalic}
)

// Typeface is a specific font face (family + style). Typefaces are immutable
// and may be shared between fonts and threads.
type Typeface struct{ ptr *C.sk_typeface_t }

func newTypeface(ptr *C.sk_typeface_t) *Typeface {
	if ptr == nil {
		return nil
	}
	t := &Typeface{ptr: ptr}
	runtime.SetFinalizer(t, (*Typeface).release)
	return t
}

// DefaultTypeface returns the platform's default typeface.
func DefaultTypeface() *Typeface { return newTypeface(C.sk_typeface_ref_default()) }

// NewTypeface looks up a typeface by family name and style. An empty family
// selects the default family. Returns nil if no match is found.
func NewTypeface(family string, style FontStyle) *Typeface {
	cs := C.sk_fontstyle_new(C.int(style.Weight), C.int(style.Width), C.sk_font_style_slant_t(style.Slant))
	defer C.sk_fontstyle_delete(cs)
	var cname *C.char
	if family != "" {
		cname = C.CString(family)
		defer C.free(unsafe.Pointer(cname))
	}
	return newTypeface(C.sk_typeface_create_from_name(cname, cs))
}

// NewTypefaceFromData creates a typeface from font file bytes (TTF/OTF). index
// selects a face within a font collection (.ttc); use 0 for a single face.
func NewTypefaceFromData(fontData []byte, index int) *Typeface {
	if len(fontData) == 0 {
		return nil
	}
	d := C.sk_data_new_with_copy(unsafe.Pointer(&fontData[0]), C.size_t(len(fontData)))
	defer C.sk_data_unref(d)
	return newTypeface(C.sk_typeface_create_from_data(d, C.int(index)))
}

func (t *Typeface) release() {
	if t != nil && t.ptr != nil {
		C.sk_typeface_unref(t.ptr)
		t.ptr = nil
		runtime.SetFinalizer(t, nil)
	}
}

// Release frees this reference to the typeface.
func (t *Typeface) Release() { t.release() }

// UnicharToGlyph returns the glyph id this typeface maps the given Unicode
// codepoint to, or 0 when the typeface has no glyph for it — a cheap
// character-coverage test that reuses the already-loaded typeface.
func (t *Typeface) UnicharToGlyph(r rune) uint16 {
	g := C.sk_typeface_unichar_to_glyph(t.ptr, C.int32_t(r))
	runtime.KeepAlive(t)
	return uint16(g)
}

// Font is a typeface combined with size and other rendering parameters.
type Font struct{ ptr *C.sk_font_t }

// NewFont creates a font from a typeface and size (in pixels). If tf is nil the
// platform default typeface is used.
func NewFont(tf *Typeface, size float32) *Font {
	var tp *C.sk_typeface_t
	if tf != nil {
		tp = tf.ptr
	}
	f := &Font{ptr: C.sk_font_new_with_values(tp, C.float(size), 1, 0)}
	runtime.SetFinalizer(f, (*Font).release)
	runtime.KeepAlive(tf)
	return f
}

// SetSize sets the text size in pixels.
func (f *Font) SetSize(size float32) { C.sk_font_set_size(f.ptr, C.float(size)); runtime.KeepAlive(f) }

// Size returns the text size in pixels.
func (f *Font) Size() float32 { s := float32(C.sk_font_get_size(f.ptr)); runtime.KeepAlive(f); return s }

// SetEdging sets glyph edge antialiasing.
func (f *Font) SetEdging(e FontEdging) { C.sk_font_set_edging(f.ptr, C.sk_font_edging_t(e)); runtime.KeepAlive(f) }

// SetHinting sets the glyph hinting level.
func (f *Font) SetHinting(h FontHinting) { C.sk_font_set_hinting(f.ptr, C.sk_font_hinting_t(h)); runtime.KeepAlive(f) }

// SetSubpixel enables sub-pixel glyph positioning.
func (f *Font) SetSubpixel(v bool) { C.sk_font_set_subpixel(f.ptr, C.bool(v)); runtime.KeepAlive(f) }

// MeasureText returns the advance width and the tight bounding box of text.
func (f *Font) MeasureText(text string, paint *Paint) (width float32, bounds Rect) {
	if text == "" {
		return 0, Rect{}
	}
	b := []byte(text)
	var cb C.sk_rect_t
	w := C.sk_font_measure_text(f.ptr, unsafe.Pointer(&b[0]), C.size_t(len(b)),
		C.sk_text_encoding_t(C.UTF8_SK_TEXT_ENCODING), &cb, paintPtr(paint))
	runtime.KeepAlive(f)
	runtime.KeepAlive(b)
	runtime.KeepAlive(paint)
	return float32(w), rectFromC(cb)
}

// FontMetrics holds a font's vertical metrics in pixels at the font's current
// size. Ascent/Top are negative (above the baseline); Descent/Bottom positive
// (below). These are the typeface's real, designed metrics — use them instead of
// guessing ascent/descent from the size.
type FontMetrics struct {
	Top       float32 // 基线到最高字形顶（负，含重音/修饰区）
	Ascent    float32 // 基线到推荐顶（负）
	Descent   float32 // 基线到推荐底（正）
	Bottom    float32 // 基线到最低字形底（正）
	Leading   float32 // 行间额外间距
	CapHeight float32 // 大写字母高度
	XHeight   float32 // 小写 x 高度
}

// Metrics returns the font's real vertical metrics (from Skia sk_font_get_metrics)
// and the recommended line spacing (descent - ascent + leading), both in pixels.
func (f *Font) Metrics() (m FontMetrics, lineSpacing float32) {
	var cm C.sk_fontmetrics_t
	ls := C.sk_font_get_metrics(f.ptr, &cm)
	runtime.KeepAlive(f)
	return FontMetrics{
		Top:       float32(cm.fTop),
		Ascent:    float32(cm.fAscent),
		Descent:   float32(cm.fDescent),
		Bottom:    float32(cm.fBottom),
		Leading:   float32(cm.fLeading),
		CapHeight: float32(cm.fCapHeight),
		XHeight:   float32(cm.fXHeight),
	}, float32(ls)
}

func (f *Font) release() {
	if f != nil && f.ptr != nil {
		C.sk_font_delete(f.ptr)
		f.ptr = nil
		runtime.SetFinalizer(f, nil)
	}
}

// Release frees the font.
func (f *Font) Release() { f.release() }

// --- Canvas integration -----------------------------------------------------

// DrawText draws a single run of UTF-8 text with its baseline starting at
// (x, y), using font for the glyphs and paint for color/style.
func (c *Canvas) DrawText(text string, x, y float32, font *Font, paint *Paint) {
	if text == "" || font == nil {
		return
	}
	b := []byte(text)
	C.sk_canvas_draw_simple_text(c.ptr, unsafe.Pointer(&b[0]), C.size_t(len(b)),
		C.sk_text_encoding_t(C.UTF8_SK_TEXT_ENCODING), C.float(x), C.float(y), font.ptr, paintPtr(paint))
	runtime.KeepAlive(c)
	runtime.KeepAlive(font)
	runtime.KeepAlive(paint)
	runtime.KeepAlive(b)
}
