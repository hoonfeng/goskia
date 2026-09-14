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

// c44 expands the 3x3 matrix to the 4x4 form passed to the canvas
// set/concat entry points. libSkiaSharp's AsM44 interprets the 16 floats
// as SkM44's column-major constructor arguments, so the values are laid out
// by COLUMN (m<col><row> in the C struct memory order):
//
//	col0: ScaleX, SkewY, 0, Persp0
//	col1: SkewX, ScaleY, 0, Persp1
//	col2: 0, 0, 1, 0
//	col3: TransX, TransY, 0, Persp2
func (m Matrix) c44() C.sk_matrix44_t {
	m33 := C.float(m.Persp2)
	// ★ 字面量 Matrix（直接写字段而非 MatrixIdentity 构造）的 Persp 全为
	// 零值 → 4×4 的 m33=0 → 矩阵奇异（w 分量恒 0 → 变换结果 NaN → 绘制
	// 直接消失。canvas 2D transform/setTransform 曾因此整段失效）。仿射
	// 矩阵的 Persp2 默认应为 1；仅当调用方显式给 Persp 分量时按原值。
	if m.Persp2 == 0 && m.Persp0 == 0 && m.Persp1 == 0 {
		m33 = 1
	}
	return C.sk_matrix44_t{
		m00: C.float(m.ScaleX), m01: C.float(m.SkewY), m02: 0, m03: C.float(m.Persp0),
		m10: C.float(m.SkewX), m11: C.float(m.ScaleY), m12: 0, m13: C.float(m.Persp1),
		m20: 0, m21: 0, m22: 1, m23: 0,
		m30: C.float(m.TransX), m31: C.float(m.TransY), m32: 0, m33: m33,
	}
}

// matrixFromC44 从 sk_canvas_get_matrix 的输出构造 Matrix。
// 与 c44 对称：libSkiaSharp 按列主序读写（SkM44 内部列主序存储），
// 16 个 float 的物理顺序为 col<col><row>：
//
//	col0: ScaleX, SkewY, 0, Persp0
//	col1: SkewX, ScaleY, 0, Persp1
//	col2: 0, 0, 1, 0
//	col3: TransX, TransY, 0, Persp2
func matrixFromC44(c C.sk_matrix44_t) Matrix {
	return Matrix{
		ScaleX: float32(c.m00), SkewY: float32(c.m01), Persp0: float32(c.m03),
		SkewX: float32(c.m10), ScaleY: float32(c.m11), Persp1: float32(c.m13),
		TransX: float32(c.m30), TransY: float32(c.m31), Persp2: float32(c.m33),
	}
}
