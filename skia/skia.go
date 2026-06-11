// Package skia provides idiomatic Go bindings for Skia, Google's 2D graphics
// library, via the prebuilt libSkiaSharp C API.
//
// See the repository README for the cross-compilation workflow.
package skia

// #include "goskia.h"
import "C"

// Init initializes Skia's process-global state (font/resource caches, etc.).
//
// It is optional but recommended to call once at program startup. It is
// idempotent and safe to call more than once.
func Init() { C.sk_graphics_init() }

// SetFontCacheLimit caps the process-global font (glyph) cache at maxBytes and
// returns the previous limit. Bounds CPU-side glyph cache memory.
func SetFontCacheLimit(maxBytes int) int {
	return int(C.sk_graphics_set_font_cache_limit(C.size_t(maxBytes)))
}

// FontCacheUsed returns the process-global font cache bytes currently in use.
func FontCacheUsed() int { return int(C.sk_graphics_get_font_cache_used()) }

// PurgeAllCaches frees all process-global Skia caches (font + resource).
func PurgeAllCaches() { C.sk_graphics_purge_all_caches() }
