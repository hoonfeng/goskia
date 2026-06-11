package skia

// #include "goskia.h"
import "C"

// Matrix is a 3x3 affine/perspective transform, matching Skia's SkMatrix.
// Most fields default to the identity transform's zero values except the
// diagonal; prefer the constructors below.
type Matrix struct {
	ScaleX, SkewX, TransX float32
	SkewY, ScaleY, TransY float32
	Persp0, Persp1, Persp2 float32
}

// MatrixIdentity returns the identity matrix.
func MatrixIdentity() Matrix {
	return Matrix{ScaleX: 1, ScaleY: 1, Persp2: 1}
}

// MatrixTranslate returns a translation matrix.
func MatrixTranslate(dx, dy float32) Matrix {
	m := MatrixIdentity()
	m.TransX, m.TransY = dx, dy
	return m
}

// MatrixScale returns a scaling matrix.
func MatrixScale(sx, sy float32) Matrix {
	m := MatrixIdentity()
	m.ScaleX, m.ScaleY = sx, sy
	return m
}

func (m Matrix) c() C.sk_matrix_t {
	return C.sk_matrix_t{
		scaleX: C.float(m.ScaleX), skewX: C.float(m.SkewX), transX: C.float(m.TransX),
		skewY: C.float(m.SkewY), scaleY: C.float(m.ScaleY), transY: C.float(m.TransY),
		persp0: C.float(m.Persp0), persp1: C.float(m.Persp1), persp2: C.float(m.Persp2),
	}
}

func matrixFromC(c C.sk_matrix_t) Matrix {
	return Matrix{
		ScaleX: float32(c.scaleX), SkewX: float32(c.skewX), TransX: float32(c.transX),
		SkewY: float32(c.skewY), ScaleY: float32(c.scaleY), TransY: float32(c.transY),
		Persp0: float32(c.persp0), Persp1: float32(c.persp1), Persp2: float32(c.persp2),
	}
}

// c44 expands the 3x3 matrix to the row-major 4x4 form expected by some canvas
// entry points (set/get/concat matrix), following Skia's SkM44(SkMatrix).
func (m Matrix) c44() C.sk_matrix44_t {
	return C.sk_matrix44_t{
		m00: C.float(m.ScaleX), m01: C.float(m.SkewX), m02: 0, m03: C.float(m.TransX),
		m10: C.float(m.SkewY), m11: C.float(m.ScaleY), m12: 0, m13: C.float(m.TransY),
		m20: 0, m21: 0, m22: 1, m23: 0,
		m30: C.float(m.Persp0), m31: C.float(m.Persp1), m32: 0, m33: C.float(m.Persp2),
	}
}

func matrixFromC44(c C.sk_matrix44_t) Matrix {
	return Matrix{
		ScaleX: float32(c.m00), SkewX: float32(c.m01), TransX: float32(c.m03),
		SkewY: float32(c.m10), ScaleY: float32(c.m11), TransY: float32(c.m13),
		Persp0: float32(c.m30), Persp1: float32(c.m31), Persp2: float32(c.m33),
	}
}
