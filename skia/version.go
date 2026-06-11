package skia

import "github.com/hoonfeng/goskia/internal/buildinfo"

// Pinned upstream revisions. See internal/buildinfo for the canonical values.
//
// The prebuilt native libSkiaSharp binaries (fetched by cmd/goskia-fetch) and
// the vendored C-API headers under csrc/include/c MUST come from these exact
// revisions, or the ABI will not match and calls will crash.
const (
	// SkiaSharpVersion is the NuGet package version that supplies the
	// libSkiaSharp.{dll,so,dylib} native libraries.
	SkiaSharpVersion = buildinfo.SkiaSharpVersion

	// SkiaForkCommit is the mono/skia commit that the vendored C-API headers
	// (csrc/include/c/*.h) were taken from.
	SkiaForkCommit = buildinfo.SkiaForkCommit
)
