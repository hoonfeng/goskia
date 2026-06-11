package skia

// #include "goskia.h"
import "C"

import (
	"runtime"
	"unsafe"
)

// Data is an immutable, reference-counted byte buffer (Skia's SkData). It is
// used to pass encoded image bytes, file contents, etc. across the boundary.
type Data struct{ ptr *C.sk_data_t }

func newData(ptr *C.sk_data_t) *Data {
	if ptr == nil {
		return nil
	}
	d := &Data{ptr: ptr}
	runtime.SetFinalizer(d, (*Data).release)
	return d
}

// NewData returns a Data holding a copy of b.
func NewData(b []byte) *Data {
	if len(b) == 0 {
		return newData(C.sk_data_new_empty())
	}
	return newData(C.sk_data_new_with_copy(unsafe.Pointer(&b[0]), C.size_t(len(b))))
}

// Size returns the number of bytes held.
func (d *Data) Size() int { return int(C.sk_data_get_size(d.ptr)) }

// Bytes returns a Go copy of the data.
func (d *Data) Bytes() []byte {
	n := C.sk_data_get_size(d.ptr)
	if n == 0 {
		return nil
	}
	out := C.GoBytes(C.sk_data_get_data(d.ptr), C.int(n))
	runtime.KeepAlive(d)
	return out
}

func (d *Data) release() {
	if d != nil && d.ptr != nil {
		C.sk_data_unref(d.ptr)
		d.ptr = nil
		runtime.SetFinalizer(d, nil)
	}
}

// Release frees this reference to the underlying buffer. It is safe to call
// multiple times.
func (d *Data) Release() { d.release() }
