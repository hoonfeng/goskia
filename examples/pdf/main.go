// Command pdf renders a one-page vector PDF with goskia.
//
//	go run ./examples/pdf [output.pdf]
package main

import (
	"log"
	"os"

	"github.com/hoonfeng/goskia/skia"
)

func main() {
	out := "out.pdf"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	skia.Init()

	doc := skia.NewPDFDocument()
	if doc == nil {
		log.Fatal("failed to create PDF document")
	}

	// A4 page in PostScript points (1/72 inch).
	c := doc.BeginPage(595, 842)
	c.Clear(skia.ColorWhite)

	fill := skia.NewPaint()
	fill.SetAntialias(true)
	fill.SetColor(skia.RGB(0x1e, 0x88, 0xe5))
	c.DrawCircle(297, 320, 130, fill)
	fill.Release()

	font := skia.NewFont(skia.NewTypeface("", skia.FontStyleBold), 42)
	font.SetEdging(skia.FontEdgingAntialias)
	ink := skia.NewPaint()
	ink.SetAntialias(true)
	ink.SetColor(skia.ColorBlack)
	w, _ := font.MeasureText("goskia PDF", ink)
	c.DrawText("goskia PDF", (595-w)/2, 560, font, ink)
	font.Release()
	ink.Release()

	doc.EndPage()

	data, err := doc.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%d bytes)", out, len(data))
}
