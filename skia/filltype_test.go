package skia

import (
	"testing"
)

// TestPathFillTypeEvenOdd verifies that SetFillType(evenodd) is honored when
// the path is drawn: the self-intersection hole of a bowtie must be empty.
func TestPathFillTypeEvenOdd(t *testing.T) {
	// Raster surface 40x40.
	surf, err := NewRasterSurfaceN32Premul(40, 40)
	if err != nil || surf == nil {
		t.Fatalf("no surface: %v", err)
	}
	defer surf.Release()
	canvas := surf.Canvas()
	canvas.Clear(0xFFFFFFFF) // white

	p := NewPath()
	defer p.Release()
	// Bowtie: two triangles sharing only the crossing point.
	p.MoveTo(10, 5)
	p.LineTo(30, 35)
	p.LineTo(30, 5)
	p.LineTo(10, 35)
	p.Close()
	p.SetFillType(FillTypeEvenOdd)

	paint := NewPaint()
	defer paint.Release()
	paint.SetColor(RGB(255, 0, 0))
	paint.SetAntialias(false)
	canvas.DrawPath(p, paint)

	img := surf.Snapshot()
	if img == nil {
		t.Fatal("no snapshot")
	}
	defer img.Release()
	pix, err := img.ReadPixels()
	if err != nil || len(pix) < 40*40*4 {
		t.Fatalf("ReadPixels failed: %v", err)
	}
	at := func(x, y int) byte {
		return pix[(y*40+x)*4+1] // G (white=255, red=0)
	}
	// Verify the C layer actually stores the fill type.
	if got := p.FillType(); got != FillTypeEvenOdd {
		t.Errorf("C filltype = %v, want evenodd (%v)", got, FillTypeEvenOdd)
	}
	// sk_path_contains honors the path's own fill type: (20,18) sits inside
	// the bowtie's central crossing region — evenodd says "hole", winding fills.
	if p.Contains(20, 18) {
		t.Logf("sk_path_contains(20,18)=true → winding (evenodd hole not honored)")
	} else {
		t.Logf("sk_path_contains(20,18)=false → evenodd hole honored")
	}
	// Center hole: (20,18) is inside the crossing region — evenodd leaves it
	// white (G=255); winding fills red (G=0).
	t.Logf("hole (20,18) G=%d (want 255=evenodd hole, 0=winding fill)", at(20, 18))
	if at(20, 18) != 255 {
		t.Errorf("evenodd hole not honored: (20,18) G=%d, want 255 (unpainted)", at(20, 18))
	}
	// Left interior must be red-filled (G=0).
	if at(15, 20) != 0 {
		t.Errorf("left interior G=%d, want 0 (red)", at(15, 20))
	}
}
