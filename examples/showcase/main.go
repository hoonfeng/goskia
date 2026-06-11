// Command showcase renders text and gradients to demonstrate more of the API.
//
//	go run ./examples/showcase [output.png]
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hoonfeng/goskia/skia"
)

func main() {
	out := "showcase.png"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	skia.Init()

	const W, H = 640, 360
	surf, err := skia.NewRasterSurfaceN32Premul(W, H)
	if err != nil {
		log.Fatal(err)
	}
	defer surf.Release()
	c := surf.Canvas()

	// Vertical linear-gradient background.
	bg := skia.NewPaint()
	defer bg.Release()
	grad := skia.NewLinearGradient(
		skia.Point{X: 0, Y: 0}, skia.Point{X: 0, Y: H},
		[]skia.Color{skia.RGB(0x20, 0x2a, 0x44), skia.RGB(0x4a, 0x2c, 0x6d)},
		nil, skia.TileModeClamp,
	)
	defer grad.Release()
	bg.SetShader(grad)
	c.DrawRect(skia.RectWH(W, H), bg)

	// A radial-gradient orb.
	orb := skia.NewPaint()
	defer orb.Release()
	orb.SetAntialias(true)
	orbShader := skia.NewRadialGradient(
		skia.Point{X: 110, Y: 250}, 70,
		[]skia.Color{skia.RGB(0xff, 0xd1, 0x80), skia.RGBA(0xff, 0x6f, 0x00, 0x00)},
		nil, skia.TileModeClamp,
	)
	defer orbShader.Release()
	orb.SetShader(orbShader)
	c.DrawCircle(110, 250, 70, orb)

	// Title text, centered.
	tf := skia.NewTypeface("", skia.FontStyleBold)
	defer tf.Release()
	title := skia.NewFont(tf, 72)
	defer title.Release()
	title.SetEdging(skia.FontEdgingAntialias)

	text := skia.NewPaint()
	defer text.Release()
	text.SetAntialias(true)
	text.SetColor(skia.ColorWhite)

	w, _ := title.MeasureText("goskia", text)
	c.DrawText("goskia", (W-w)/2, 150, title, text)

	// Subtitle.
	sub := skia.NewFont(nil, 22)
	defer sub.Release()
	sub.SetEdging(skia.FontEdgingAntialias)
	text.SetColor(skia.RGB(0xc7, 0xd0, 0xe8))
	sw, _ := sub.MeasureText("Skia bindings for Go — cross-compiled", text)
	c.DrawText("Skia bindings for Go — cross-compiled", (W-sw)/2, 195, sub, text)

	png, err := surf.EncodePNG()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(out, png, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (%d bytes)\n", out, len(png))
}
