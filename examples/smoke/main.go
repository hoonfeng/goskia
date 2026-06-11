// Command smoke is a minimal link/load test for the goskia native library.
package main

import (
	"fmt"

	"github.com/hoonfeng/goskia/skia"
)

func main() {
	skia.Init()
	fmt.Println("goskia smoke test")
	fmt.Println("  SkiaSharp version:", skia.SkiaSharpVersion)
	fmt.Println("  skia fork commit: ", skia.SkiaForkCommit)
	fmt.Println("  native library linked and loaded OK")
}
