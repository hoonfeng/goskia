// Command effects demonstrates mask/color/image filters and path effects.
//
//	go run ./examples/effects [output.png]
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hoonfeng/goskia/skia"
)

func main() {
	out := "effects.png"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	skia.Init()

	const W, H = 512, 512
	surf, err := skia.NewRasterSurfaceN32Premul(W, H)
	if err != nil {
		log.Fatal(err)
	}
	defer surf.Release()
	c := surf.Canvas()
	c.Clear(skia.RGB(0xfa, 0xfa, 0xfa))

	label := skia.NewFont(nil, 18)
	defer label.Release()
	label.SetEdging(skia.FontEdgingAntialias)
	labelPaint := skia.NewPaint()
	defer labelPaint.Release()
	labelPaint.SetAntialias(true)
	labelPaint.SetColor(skia.RGB(0x37, 0x47, 0x4f))
	text := func(s string, x, y float32) { c.DrawText(s, x, y, label, labelPaint) }

	// 1) Drop-shadow image filter.
	card := skia.NewPaint()
	card.SetAntialias(true)
	card.SetColor(skia.RGB(0x42, 0x85, 0xf4))
	shadow := skia.NewDropShadowImageFilter(0, 6, 6, 6, skia.ColorBlack.WithAlpha(0x99), nil)
	card.SetImageFilter(shadow)
	c.DrawRoundRect(skia.RectXYWH(50, 60, 170, 100), 18, 18, card)
	shadow.Release()
	card.Release()
	text("drop shadow", 60, 200)

	// 2) Blur mask filter (soft glow).
	glow := skia.NewPaint()
	glow.SetAntialias(true)
	glow.SetColor(skia.RGB(0xff, 0x6f, 0x00))
	blur := skia.NewBlurMaskFilter(skia.BlurStyleNormal, 14)
	glow.SetMaskFilter(blur)
	c.DrawCircle(380, 110, 44, glow)
	blur.Release()
	glow.Release()
	text("blur glow", 330, 200)

	// 3) Dashed stroke path effect.
	dash := skia.NewPaint()
	dash.SetAntialias(true)
	dash.SetStyle(skia.PaintStyleStroke)
	dash.SetStrokeWidth(5)
	dash.SetColor(skia.RGB(0x2e, 0x7d, 0x32))
	de := skia.NewDashPathEffect([]float32{18, 12}, 0)
	dash.SetPathEffect(de)
	c.DrawRoundRect(skia.RectXYWH(50, 300, 170, 100), 18, 18, dash)
	de.Release()
	dash.Release()
	text("dashed stroke", 60, 440)

	// 4) Grayscale color filter over a gradient.
	grad := skia.NewLinearGradient(
		skia.Point{X: 300, Y: 300}, skia.Point{X: 470, Y: 400},
		[]skia.Color{skia.RGB(0xe5, 0x39, 0x35), skia.RGB(0x8e, 0x24, 0xaa), skia.RGB(0x1e, 0x88, 0xe5)},
		nil, skia.TileModeClamp,
	)
	gp := skia.NewPaint()
	gp.SetAntialias(true)
	gp.SetShader(grad)
	gray := skia.GrayscaleColorFilter()
	gp.SetColorFilter(gray)
	c.DrawRoundRect(skia.RectXYWH(300, 300, 170, 100), 18, 18, gp)
	grad.Release()
	gray.Release()
	gp.Release()
	text("grayscale filter", 300, 440)

	png, err := surf.EncodePNG()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(out, png, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (%d bytes)\n", out, len(png))
}
