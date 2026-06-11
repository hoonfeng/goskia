// Package buildinfo holds pinned upstream revisions shared by the skia package
// and the command-line tools. It deliberately has no cgo dependency so the
// tools can import it without linking the native library.
package buildinfo

const (
	// SkiaSharpVersion is the NuGet package version that supplies the prebuilt
	// libSkiaSharp native libraries.
	SkiaSharpVersion = "3.119.4"

	// SkiaForkCommit is the mono/skia commit the vendored C-API headers came
	// from (the skia submodule pointer of SkiaSharp v3.119.4).
	SkiaForkCommit = "7dbfc07dd33181f84e0958afb7ee805c6c769f0b"
)
