package skia

import "testing"

// TestCanvasConcatCombo：scale+translate 组合矩阵（canvas 2D transform 的
// 真实形态——0.25 缩放 + (200,50) 平移）。拆分测试掩盖行列错乱。
func TestCanvasConcatCombo(t *testing.T) {
	surf, err := NewRasterSurfaceN32Premul(300, 200)
	if err != nil {
		t.Fatalf("surface: %v", err)
	}
	defer surf.Release()
	c := surf.Canvas()
	paint := NewPaint()
	paint.SetColor(RGBA(255, 0, 0, 255)) // red
	defer paint.Release()

	c.SetMatrix(Matrix{
		ScaleX: 0.25, SkewX: 0, TransX: 200,
		SkewY: 0, ScaleY: 0.25, TransY: 50,
	})
	c.DrawRect(RectXYWH(0, 0, 64, 64), paint)
	// 屏幕 (200,50)-(216,66) 内应有红
	pix := readPix(t, surf, 210, 60)
	if pix[0] < 200 || pix[3] < 200 {
		t.Fatalf("combo SetMatrix (210,60)=%v, want red", pix)
	}
	out := readPix(t, surf, 280, 100)
	if out[3] > 20 {
		t.Fatalf("combo SetMatrix outside (280,100)=%v, want transparent", out)
	}
	// Concat 同样组合
	c.ResetMatrix()
	c.Concat(Matrix{
		ScaleX: 0.25, SkewX: 0, TransX: 200,
		SkewY: 0, ScaleY: 0.25, TransY: 50,
	})
	c.DrawRect(RectXYWH(0, 0, 64, 64), paint)
	pix2 := readPix(t, surf, 210, 60)
	if pix2[0] < 200 || pix2[3] < 200 {
		t.Fatalf("combo Concat (210,60)=%v, want red", pix2)
	}
}
func TestCanvasConcatMatrix(t *testing.T) {
	surf, err := NewRasterSurfaceN32Premul(200, 200)
	if err != nil {
		t.Fatalf("surface: %v", err)
	}
	defer surf.Release()
	c := surf.Canvas()
	paint := NewPaint()
	paint.SetColor(RGBA(0, 0, 255, 255)) // blue
	defer paint.Release()

	// SetMatrix 2x scale
	c.SetMatrix(Matrix{
		ScaleX: 2, SkewX: 0, TransX: 0,
		SkewY: 0, ScaleY: 2, TransY: 0,
	})
	c.DrawRect(RectXYWH(0, 0, 10, 10), paint)
	pix := readPix(t, surf, 15, 15)
	if pix[3] < 200 || pix[2] < 200 {
		t.Fatalf("SetMatrix scaled pixel (15,15)=%v, want blue", pix)
	}
	// ResetMatrix 后（未加变换）DrawRect(100,100,10,10) → (105,105) 应为蓝
	// （若 ResetMatrix 静默失败，120,120 平移残留 → (105,105) 透明）。
	c.ResetMatrix()
	c.DrawRect(RectXYWH(100, 100, 10, 10), paint)
	pixR := readPix(t, surf, 105, 105)
	if pixR[3] < 200 || pixR[2] < 200 {
		t.Fatalf("after ResetMatrix pixel (105,105)=%v, want blue (ResetMatrix broken?)", pixR)
	}
	// Concat translate
	c.ResetMatrix()
	c.Concat(Matrix{
		ScaleX: 1, SkewX: 0, TransX: 100,
		SkewY: 0, ScaleY: 1, TransY: 100,
	})
	c.DrawRect(RectXYWH(0, 0, 10, 10), paint)
	pix3 := readPix(t, surf, 105, 105)
	if pix3[3] < 200 {
		t.Fatalf("Concat translated pixel (105,105)=%v, want non-transparent", pix3)
	}
}

func readPix(t *testing.T, surf *Surface, x, y int) []byte {
	t.Helper()
	img := surf.Snapshot()
	defer img.Release()
	buf, err := img.ReadPixels()
	if err != nil {
		t.Fatalf("ReadPixels: %v", err)
	}
	w := img.Width()
	i := (y*w + x) * 4
	if i+3 >= len(buf) {
		t.Fatalf("index out of range: x=%d y=%d w=%d len=%d", x, y, w, len(buf))
	}
	return buf[i : i+4]
}
