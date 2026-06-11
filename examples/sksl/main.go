// Command sksl demonstrates SkSL runtime shaders and lighting image filters.
//
//	go run ./examples/sksl [output.png]
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hoonfeng/goskia/skia"
)

// A procedural shader: a radial "ripple" colored by angle, with a vignette.
const ripple = `
uniform float2 iResolution;

half4 main(float2 fragCoord) {
    float2 uv = fragCoord / iResolution;
    float2 c  = uv - 0.5;
    float  r  = length(c);
    float  a  = atan(c.y, c.x);
    float  w  = 0.5 + 0.5 * sin(a * 8.0 + r * 46.0);
    half3  col = mix(half3(0.05, 0.10, 0.30), half3(1.0, 0.55, 0.15), half(w));
    col *= half(smoothstep(0.52, 0.08, r)); // vignette
    return half4(col, 1.0);
}`

func main() {
	out := "sksl.png"
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
	c.Clear(skia.ColorBlack)

	// 1) SkSL runtime shader fills the whole canvas.
	eff, err := skia.CompileRuntimeShader(ripple)
	if err != nil {
		log.Fatal(err)
	}
	defer eff.Release()

	u := eff.NewUniforms()
	if err := u.SetVec2("iResolution", W, H); err != nil {
		log.Fatal(err)
	}
	shader := eff.MakeShader(u.Bytes(), nil, nil)
	if shader == nil {
		log.Fatal("failed to instantiate SkSL shader")
	}
	defer shader.Release()

	sp := skia.NewPaint()
	defer sp.Release()
	sp.SetShader(shader)
	c.DrawRect(skia.RectWH(W, H), sp)

	// 2) A lit "sphere": light the alpha of a blurred disc (point light, specular).
	lit := skia.NewPaint()
	defer lit.Release()
	lit.SetAntialias(true)
	lit.SetColor(skia.ColorWhite)
	blur := skia.NewBlurImageFilter(12, 12, skia.TileModeDecal, nil)
	light := skia.NewPointLitSpecularImageFilter(
		skia.Point3{X: 110, Y: 360, Z: 90}, // light position above the disc
		skia.ColorWhite, 1.6 /*surfaceScale*/, 0.9 /*ks*/, 16 /*shininess*/, blur)
	lit.SetImageFilter(light)
	c.DrawCircle(150, 390, 80, lit)
	blur.Release()
	light.Release()

	png, err := surf.EncodePNG()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(out, png, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (%d bytes)\n", out, len(png))
}
