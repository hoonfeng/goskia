package skia

// This file centralizes all #cgo build directives for the package. Every other
// .go file in this package only needs `// #include "goskia.h"` in its preamble;
// the CFLAGS/LDFLAGS below are merged by cgo across the whole package.
//
// Linking model
// -------------
// goskia links dynamically against the prebuilt libSkiaSharp shared library
// (fetched per-target by `go run ./cmd/goskia-fetch` into skia/lib/<os>_<arch>/).
// At runtime the OS loader must find the library:
//
//   - Windows: libSkiaSharp.dll must sit next to the .exe (or on %PATH%).
//   - Linux:   the rpath $ORIGIN lets the loader find libSkiaSharp.so next to
//              the binary; otherwise set LD_LIBRARY_PATH.
//   - macOS:   the rpath @loader_path finds libSkiaSharp.dylib next to the
//              binary. cgo rejects the '@' in this flag unless allowed, so a
//              plain `go build` for darwin needs:
//                  export CGO_LDFLAGS_ALLOW='-Wl,-rpath,@loader_path'
//              (the `goskia build` tool sets this for you).
//
// The `goskia` build tool copies the right library next to your output binary
// and sets any required CGO_LDFLAGS_ALLOW.

/*
#cgo CFLAGS: -I${SRCDIR}/csrc

#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/lib/windows_amd64 -lSkiaSharp
#cgo windows,arm64 LDFLAGS: -L${SRCDIR}/lib/windows_arm64 -lSkiaSharp

#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib/linux_amd64 -lSkiaSharp -Wl,-rpath,$ORIGIN
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/linux_arm64 -lSkiaSharp -Wl,-rpath,$ORIGIN

#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib/darwin_amd64 -lSkiaSharp -Wl,-rpath,@loader_path
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib/darwin_arm64 -lSkiaSharp -Wl,-rpath,@loader_path

#include "goskia.h"
*/
import "C"
