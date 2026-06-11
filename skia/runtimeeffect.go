package skia

// #include "goskia.h"
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"unsafe"
)

// RuntimeEffect is a compiled SkSL program. Compile it once, then instantiate
// it many times (with different uniforms) as a Shader, ColorFilter, or Blender.
type RuntimeEffect struct{ ptr *C.sk_runtimeeffect_t }

func compileEffect(sksl string, kind int) (*RuntimeEffect, error) {
	csrc := C.CString(sksl)
	defer C.free(unsafe.Pointer(csrc))
	src := C.sk_string_new_with_copy(csrc, C.size_t(len(sksl)))
	defer C.sk_string_destructor(src)
	errStr := C.sk_string_new_empty()
	defer C.sk_string_destructor(errStr)

	var eff *C.sk_runtimeeffect_t
	switch kind {
	case 0:
		eff = C.sk_runtimeeffect_make_for_shader(src, errStr)
	case 1:
		eff = C.sk_runtimeeffect_make_for_color_filter(src, errStr)
	case 2:
		eff = C.sk_runtimeeffect_make_for_blender(src, errStr)
	}
	if eff == nil {
		msg := C.GoString(C.sk_string_get_c_str(errStr))
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("skia: SkSL compile failed: %s", msg)
	}
	e := &RuntimeEffect{ptr: eff}
	runtime.SetFinalizer(e, (*RuntimeEffect).release)
	return e, nil
}

// CompileRuntimeShader compiles an SkSL shader: `half4 main(float2 fragCoord)`.
func CompileRuntimeShader(sksl string) (*RuntimeEffect, error) { return compileEffect(sksl, 0) }

// CompileRuntimeColorFilter compiles an SkSL color filter:
// `half4 main(half4 inColor)`.
func CompileRuntimeColorFilter(sksl string) (*RuntimeEffect, error) { return compileEffect(sksl, 1) }

// CompileRuntimeBlender compiles an SkSL blender: `half4 main(half4 src, half4 dst)`.
func CompileRuntimeBlender(sksl string) (*RuntimeEffect, error) { return compileEffect(sksl, 2) }

// UniformByteSize returns the size in bytes of the effect's uniform block.
func (e *RuntimeEffect) UniformByteSize() int {
	n := int(C.sk_runtimeeffect_get_uniform_byte_size(e.ptr))
	runtime.KeepAlive(e)
	return n
}

// MakeShader instantiates the effect as a Shader. uniforms is the packed
// uniform block (build it with NewUniforms); children supplies any child
// shaders declared in the SkSL; localMatrix may be nil.
func (e *RuntimeEffect) MakeShader(uniforms []byte, children []*Shader, localMatrix *Matrix) *Shader {
	data := uniformData(uniforms)
	defer C.sk_data_unref(data)
	cp, arr := flattenChildren(children)
	var cm *C.sk_matrix_t
	if localMatrix != nil {
		m := localMatrix.c()
		cm = &m
	}
	sh := C.sk_runtimeeffect_make_shader(e.ptr, data, cp, C.size_t(len(children)), cm)
	runtime.KeepAlive(e)
	runtime.KeepAlive(arr)
	runtime.KeepAlive(children)
	return newShader(sh)
}

// MakeColorFilter instantiates the effect as a ColorFilter.
func (e *RuntimeEffect) MakeColorFilter(uniforms []byte, children []*Shader) *ColorFilter {
	data := uniformData(uniforms)
	defer C.sk_data_unref(data)
	cp, arr := flattenChildren(children)
	cf := C.sk_runtimeeffect_make_color_filter(e.ptr, data, cp, C.size_t(len(children)))
	runtime.KeepAlive(e)
	runtime.KeepAlive(arr)
	runtime.KeepAlive(children)
	return newColorFilter(cf)
}

func uniformData(uniforms []byte) *C.sk_data_t {
	if len(uniforms) == 0 {
		return C.sk_data_new_empty()
	}
	return C.sk_data_new_with_copy(unsafe.Pointer(&uniforms[0]), C.size_t(len(uniforms)))
}

func flattenChildren(children []*Shader) (**C.sk_flattenable_t, []*C.sk_flattenable_t) {
	if len(children) == 0 {
		return nil, nil
	}
	arr := make([]*C.sk_flattenable_t, len(children))
	for i, c := range children {
		if c != nil {
			arr[i] = (*C.sk_flattenable_t)(unsafe.Pointer(c.ptr))
		}
	}
	return &arr[0], arr
}

func (e *RuntimeEffect) release() {
	if e != nil && e.ptr != nil {
		C.sk_runtimeeffect_unref(e.ptr)
		e.ptr = nil
		runtime.SetFinalizer(e, nil)
	}
}

// Release frees the compiled effect.
func (e *RuntimeEffect) Release() { e.release() }

// Uniforms builds the packed uniform block for a RuntimeEffect by uniform name,
// so callers don't compute byte offsets by hand.
type Uniforms struct {
	e   *RuntimeEffect
	buf []byte
}

// NewUniforms returns a zeroed uniform block sized for the effect.
func (e *RuntimeEffect) NewUniforms() *Uniforms {
	return &Uniforms{e: e, buf: make([]byte, e.UniformByteSize())}
}

// Bytes returns the packed block to pass to MakeShader / MakeColorFilter.
func (u *Uniforms) Bytes() []byte { return u.buf }

func (u *Uniforms) offset(name string) (int, bool) {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	var cu C.sk_runtimeeffect_uniform_t
	C.sk_runtimeeffect_get_uniform_from_name(u.e.ptr, cn, C.size_t(len(name)), &cu)
	runtime.KeepAlive(u.e)
	if cu.fName == nil {
		return 0, false
	}
	if C.GoStringN(cu.fName, C.int(cu.fNameLength)) != name {
		return 0, false
	}
	return int(cu.fOffset), true
}

func (u *Uniforms) putFloats(name string, vals ...float32) error {
	off, ok := u.offset(name)
	if !ok {
		return fmt.Errorf("skia: uniform %q not found", name)
	}
	if off+4*len(vals) > len(u.buf) {
		return fmt.Errorf("skia: uniform %q write out of range", name)
	}
	for i, v := range vals {
		binary.LittleEndian.PutUint32(u.buf[off+4*i:], math.Float32bits(v))
	}
	return nil
}

// SetFloat sets a `float` uniform.
func (u *Uniforms) SetFloat(name string, v float32) error { return u.putFloats(name, v) }

// SetVec2 sets a `float2` uniform.
func (u *Uniforms) SetVec2(name string, x, y float32) error { return u.putFloats(name, x, y) }

// SetVec3 sets a `float3` uniform.
func (u *Uniforms) SetVec3(name string, x, y, z float32) error { return u.putFloats(name, x, y, z) }

// SetVec4 sets a `float4` uniform.
func (u *Uniforms) SetVec4(name string, x, y, z, w float32) error { return u.putFloats(name, x, y, z, w) }

// SetInt sets an `int` uniform.
func (u *Uniforms) SetInt(name string, v int32) error {
	off, ok := u.offset(name)
	if !ok {
		return fmt.Errorf("skia: uniform %q not found", name)
	}
	if off+4 > len(u.buf) {
		return fmt.Errorf("skia: uniform %q write out of range", name)
	}
	binary.LittleEndian.PutUint32(u.buf[off:], uint32(v))
	return nil
}
