package skia_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/hoonfeng/goskia/skia"
)

func TestColorComponents(t *testing.T) {
	c := skia.RGBA(0x12, 0x34, 0x56, 0x78)
	if c.R() != 0x12 || c.G() != 0x34 || c.B() != 0x56 || c.A() != 0x78 {
		t.Fatalf("unexpected components from %#x: r=%#x g=%#x b=%#x a=%#x", uint32(c), c.R(), c.G(), c.B(), c.A())
	}
	if got := c.WithAlpha(0xff).A(); got != 0xff {
		t.Fatalf("WithAlpha: got %#x", got)
	}
	if skia.RGB(1, 2, 3).A() != 0xff {
		t.Fatal("RGB should be opaque")
	}
}

func TestRectHelpers(t *testing.T) {
	r := skia.RectXYWH(10, 20, 30, 40)
	if r.Width() != 30 || r.Height() != 40 {
		t.Fatalf("RectXYWH size: %v", r)
	}
	if r.Right != 40 || r.Bottom != 60 {
		t.Fatalf("RectXYWH edges: %v", r)
	}
}

func TestPathBounds(t *testing.T) {
	p := skia.NewPath().MoveTo(0, 0).LineTo(100, 0).LineTo(100, 50).Close()
	defer p.Release()
	b := p.Bounds()
	if b.Left != 0 || b.Top != 0 || b.Right != 100 || b.Bottom != 50 {
		t.Fatalf("unexpected path bounds: %v", b)
	}
	if p.CountPoints() == 0 {
		t.Fatal("expected non-zero point count")
	}
}

// TestRenderDecode renders a known scene and decodes the PNG back, validating
// the full surface -> canvas -> snapshot -> encode pipeline end to end.
func TestRenderDecode(t *testing.T) {
	skia.Init()
	const W, H = 64, 48
	surf, err := skia.NewRasterSurfaceN32Premul(W, H)
	if err != nil {
		t.Fatal(err)
	}
	defer surf.Release()

	c := surf.Canvas()
	c.Clear(skia.ColorWhite)

	paint := skia.NewPaint()
	defer paint.Release()
	paint.SetColor(skia.RGB(0xff, 0x00, 0x00)) // opaque red
	c.DrawRect(skia.RectXYWH(0, 0, W, H), paint)

	data, err := surf.EncodePNG()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 8 || !bytes.Equal(data[:8], []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatal("output is not a PNG")
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if img.Bounds() != image.Rect(0, 0, W, H) {
		t.Fatalf("unexpected bounds: %v", img.Bounds())
	}
	r, g, b, a := img.At(W/2, H/2).RGBA()
	// 16-bit components; red opaque means r=0xffff, g=b=0.
	if r>>8 != 0xff || g != 0 || b != 0 || a>>8 != 0xff {
		t.Fatalf("center pixel not red: r=%#x g=%#x b=%#x a=%#x", r, g, b, a)
	}
}

func TestEncodeJPEG(t *testing.T) {
	skia.Init()
	surf, err := skia.NewRasterSurfaceN32Premul(32, 32)
	if err != nil {
		t.Fatal(err)
	}
	defer surf.Release()
	surf.Canvas().Clear(skia.ColorBlue)

	data, err := surf.EncodeToBytes(skia.EncodedFormatJPEG, 90)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 3 || data[0] != 0xff || data[1] != 0xd8 {
		t.Fatalf("not a JPEG (len=%d)", len(data))
	}
}
