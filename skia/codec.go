package skia

// #include "goskia.h"
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// CodecFrame 是动画里的一帧：解码后的位图 + 该帧的停留时长。
type CodecFrame struct {
	// Image 是解码后的帧（RGBA8888 预乘，与 ImageInfoN32Premul 同一格式）。
	Image *Image
	// DurationMS 是该帧的停留时长（毫秒）；0 表示容器没给（兜底时长由调用方决定）。
	DurationMS int
}

// Codec 是编码图像的多帧解码器（Skia 的 SkCodec）：GIF / WebP / APNG 等动画容器。
//
// 与 DecodeImage 的差别：DecodeImage 只解**第一帧**，动画的其余帧要按索引解码，
// 且帧之间可能有依赖（增量帧只描述与前一帧的差异）——DecodeFrame 用 fPriorFrame
// 告诉 SkCodec 前一帧是谁。引擎侧（wb-ui）的动图帧推进靠它拿到全部帧与延时。
type Codec struct{ ptr *C.sk_codec_t }

func newCodec(ptr *C.sk_codec_t) *Codec {
	if ptr == nil {
		return nil
	}
	c := &Codec{ptr: ptr}
	runtime.SetFinalizer(c, (*Codec).release)
	return c
}

// NewCodec 用编码字节（GIF/WebP/PNG/JPEG/APNG…）建多帧解码器。静态图也能建，
// 此时 FrameCount() == 1。
func NewCodec(encoded []byte) (*Codec, error) {
	if len(encoded) == 0 {
		return nil, fmt.Errorf("skia: NewCodec: empty input")
	}
	d := C.sk_data_new_with_copy(unsafe.Pointer(&encoded[0]), C.size_t(len(encoded)))
	if d == nil {
		return nil, fmt.Errorf("skia: NewCodec: sk_data_new_with_copy failed")
	}
	defer C.sk_data_unref(d)
	c := C.sk_codec_new_from_data(d)
	if c == nil {
		return nil, fmt.Errorf("skia: NewCodec: unsupported or corrupt image data")
	}
	return newCodec(c), nil
}

// FrameCount 返回**动画帧数**。★ Skia 语义：不提供帧计数的容器（含静态 PNG/JPEG）
// 返回 0——它表示「没有动画帧计数」，不等于「没有图像」。要「总能拿到至少一帧」的
// 语义请用 DecodeFrames()（把 0 当 1 处理）。
func (c *Codec) FrameCount() int {
	if c == nil || c.ptr == nil {
		return 0
	}
	n := int(C.sk_codec_get_frame_count(c.ptr))
	if n < 0 {
		return 0
	}
	return n
}

// RepetitionCount 返回循环次数（Skia 语义：-1 = 无限循环，0 = 不循环，n > 0 = 再播 n 次）。
func (c *Codec) RepetitionCount() int {
	if c == nil || c.ptr == nil {
		return 0
	}
	return int(C.sk_codec_get_repetition_count(c.ptr))
}

// Dimensions 返回容器里的图像尺寸（解码前即可得）。
func (c *Codec) Dimensions() (int, int) {
	if c == nil || c.ptr == nil {
		return 0, 0
	}
	var info C.sk_imageinfo_t
	C.sk_codec_get_info(c.ptr, &info)
	return int(info.width), int(info.height)
}

// FrameDurationMS 返回第 i 帧的停留时长（毫秒）；越界或容器不给时返回 0。
func (c *Codec) FrameDurationMS(i int) int {
	if c == nil || c.ptr == nil {
		return 0
	}
	var fi C.sk_codec_frameinfo_t
	if !C.sk_codec_get_frame_info_for_index(c.ptr, C.int(i), &fi) {
		return 0
	}
	if d := int(fi.fDuration); d > 0 {
		return d
	}
	return 0
}

// DecodeFrame 把第 i 帧解成 RGBA8888 预乘位图。
//
// fPriorFrame 传 i-1（首帧传 -1）：GIF/APNG 的帧可能是「增量」的（只描述与前一帧
// 的差异），SkCodec 用它决定要不要先解前置帧。
func (c *Codec) DecodeFrame(i int) (*Image, error) {
	if c == nil || c.ptr == nil {
		return nil, fmt.Errorf("skia: DecodeFrame: nil codec")
	}
	if n := c.FrameCount(); n > 0 && (i < 0 || i >= n) {
		return nil, fmt.Errorf("skia: DecodeFrame: frame %d out of range (0..%d)", i, n-1)
	}
	w, h := c.Dimensions()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("skia: DecodeFrame: bad dimensions %dx%d", w, h)
	}
	info := ImageInfoN32Premul(w, h)
	ci := info.c()
	rowBytes := info.MinRowBytes()
	pixels := make([]byte, info.ByteSize())
	opts := C.sk_codec_options_t{
		fZeroInitialized: C.NO_SK_CODEC_ZERO_INITIALIZED,
		fSubset:          nil,
		fFrameIndex:      C.int(i),
		fPriorFrame:      C.int(i - 1),
	}
	res := C.sk_codec_get_pixels(c.ptr, &ci, unsafe.Pointer(&pixels[0]), C.size_t(rowBytes), &opts)
	if res != C.SUCCESS_SK_CODEC_RESULT {
		return nil, fmt.Errorf("skia: DecodeFrame(%d): sk_codec_get_pixels = %d", i, int(res))
	}
	img := C.sk_image_new_raster_copy(&ci, unsafe.Pointer(&pixels[0]), C.size_t(rowBytes))
	if img == nil {
		return nil, fmt.Errorf("skia: DecodeFrame(%d): sk_image_new_raster_copy failed", i)
	}
	return newImage(img), nil
}

// DecodeFrames 解码全部帧（含每帧时长）。静态图返回单帧。
func (c *Codec) DecodeFrames() ([]CodecFrame, error) {
	n := c.FrameCount()
	if n <= 0 {
		// 静态容器没有帧计数（Skia 返回 0）：按单帧处理——上层不必区分「动画」与
		// 「静态图」，两者走同一条解码路径。
		n = 1
	}
	out := make([]CodecFrame, 0, n)
	for i := 0; i < n; i++ {
		img, err := c.DecodeFrame(i)
		if err != nil {
			return nil, err
		}
		out = append(out, CodecFrame{Image: img, DurationMS: c.FrameDurationMS(i)})
	}
	return out, nil
}

func (c *Codec) release() {
	if c != nil && c.ptr != nil {
		C.sk_codec_destroy(c.ptr)
		c.ptr = nil
		runtime.SetFinalizer(c, nil)
	}
}

// Release frees this reference to the underlying codec. Safe to call multiple times.
func (c *Codec) Release() { c.release() }
