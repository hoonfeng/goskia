// Command hello renders a small composition with goskia and writes it to a PNG.
//
// Usage:
//
//	go run ./examples/hello [output.png]
package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/hoonfeng/goskia/skia"
)

func main() {
	out := "hello.png"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	skia.Init()

	const W, H = 512, 512
	surface, err := skia.NewRasterSurfaceN32Premul(W, H)
	if err != nil {
		log.Fatal(err)
	}
	defer surface.Release()

	c := surface.Canvas()
	c.Clear(skia.RGB(0xf5, 0xf5, 0xf5))

	// Three overlapping translucent circles — the classic Porter-Duff demo.
	paint := skia.NewPaint()
	defer paint.Release()
	paint.SetAntialias(true)
	paint.SetStyle(skia.PaintStyleFill)

	paint.SetColor(skia.RGB(0xe5, 0x39, 0x35).WithAlpha(0xb0))
	c.DrawCircle(200, 200, 110, paint)
	paint.SetColor(skia.RGBA(0x43, 0xa0, 0x47, 0xb0))
	c.DrawCircle(312, 200, 110, paint)
	paint.SetColor(skia.RGBA(0x1e, 0x88, 0xe5, 0xb0))
	c.DrawCircle(256, 300, 110, paint)

	// A filled five-pointed star via a path.
	star := buildStar(256, 256, 120, 50, 5)
	defer star.Release()
	paint.SetColor(skia.RGBA(0xff, 0xb3, 0x00, 0xcc))
	c.DrawPath(star, paint)

	// A stroked rounded-rect border around the whole canvas.
	border := skia.NewPaint()
	defer border.Release()
	border.SetAntialias(true)
	border.SetStyle(skia.PaintStyleStroke)
	border.SetStrokeWidth(6)
	border.SetColor(skia.RGB(0x37, 0x47, 0x4f))
	c.DrawRoundRect(skia.RectXYWH(8, 8, W-16, H-16), 24, 24, border)

	png, err := surface.EncodePNG()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(out, png, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (%d bytes, %dx%d)\n", out, len(png), W, H)
}

// buildStar returns a star path with the given number of points.
func buildStar(cx, cy, outer, inner float32, points int) *skia.Path {
	p := skia.NewPath()
	for i := 0; i < points*2; i++ {
		r := outer
		if i%2 == 1 {
			r = inner
		}
		a := math.Pi/2 + float64(i)*math.Pi/float64(points)
		x := cx + r*float32(math.Cos(a))
		y := cy - r*float32(math.Sin(a))
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	return p.Close()
}
