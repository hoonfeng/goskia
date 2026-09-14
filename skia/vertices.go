package skia

// #include "goskia.h"
import "C"

import (
	"runtime"
	"unsafe"
)

// VertexMode 是顶点网格的绘制模式（Skia SkVertices::VertexMode 对应
// SkiaSharp SKVertexMode：Triangles / TriangleStrip / TriangleFan）。
type VertexMode int

const (
	TrianglesVertexMode    VertexMode = 0 // TRIANGLES_SK_VERTICES_VERTEX_MODE
	TriangleStripVertexMode VertexMode = 1
	TriangleFanVertexMode   VertexMode = 2
)

// Vertices 表示用于 drawVertices 的顶点网格（Skia SkVertices）。
// 用途：Live2D 等「纹理三角形网格」渲染——位置数组 + 纹理坐标数组 +
// 三角形索引，一次 Skia 调用批量光栅化（替代逐三角形 clip+drawImage，
// 大幅降低 CGO/绘制调用次数）。
type Vertices struct {
	ptr *C.sk_vertices_t
}

func newVertices(ptr *C.sk_vertices_t) *Vertices {
	if ptr == nil {
		return nil
	}
	v := &Vertices{ptr: ptr}
	runtime.SetFinalizer(v, (*Vertices).release)
	return v
}

func (v *Vertices) release() {
	if v != nil && v.ptr != nil {
		C.sk_vertices_unref(v.ptr)
		v.ptr = nil
	}
}

// Release 释放底层 SkVertices（通常由 finalizer 自动处理）。
func (v *Vertices) Release() { v.release() }

// NewVerticesCopyFlat 从扁平数组构造顶点网格（SkVertices::MakeCopy，
// 内部深拷贝数据，调用后可释放入参切片）。
//
//   - positions/texs：交错布局 [x0,y0,x1,y1,...]（sk_point_t 内存布局
//     与两个连续 float32 相同 → 零拷贝直传 C）。
//   - colors：nil 表示无顶点色（顶点颜色与纹理 shader 混合，Live2D 不用）；
//     非 nil 时为 [a,r,g,b] 或 [r,g,b,a] 打包（按 sk_color_t 布局）。
//   - indices：三角形索引；Skia C-API 用 uint16（顶点数 ≤ 65535，
//     Live2D 模型超出时调用方需自行分块）。
func NewVerticesCopyFlat(mode VertexMode, positions []float32, texs []float32, colors []float32, indices []uint16) *Vertices {
	var posPtr *C.sk_point_t
	if len(positions) > 0 {
		posPtr = (*C.sk_point_t)(unsafe.Pointer(&positions[0]))
	}
	var texPtr *C.sk_point_t
	if len(texs) > 0 {
		texPtr = (*C.sk_point_t)(unsafe.Pointer(&texs[0]))
	}
	var colPtr *C.sk_color_t
	if len(colors) > 0 {
		colPtr = (*C.sk_color_t)(unsafe.Pointer(&colors[0]))
	}
	var idxPtr *C.uint16_t
	if len(indices) > 0 {
		idxPtr = (*C.uint16_t)(unsafe.Pointer(&indices[0]))
	}
	// 注意：sk_vertices_make_copy 的 colors 数组长度隐含 = vertexCount。
	ptr := C.sk_vertices_make_copy(C.sk_vertices_vertex_mode_t(mode),
		C.int(len(positions)/2), posPtr, texPtr, colPtr,
		C.int(len(indices)), idxPtr)
	return newVertices(ptr)
}
