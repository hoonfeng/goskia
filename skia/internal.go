package skia

// #include "goskia.h"
import "C"

// paintPtr returns the underlying paint pointer, or nil for a nil *Paint.
func paintPtr(p *Paint) *C.sk_paint_t {
	if p == nil {
		return nil
	}
	return p.ptr
}
