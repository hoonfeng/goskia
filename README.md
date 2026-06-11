# goskia

Idiomatic **Go bindings for [Skia](https://skia.org)**, Google's 2D graphics
library — built for **easy, fast cross-platform compilation**.

```go
surf, _ := skia.NewRasterSurfaceN32Premul(512, 512)
c := surf.Canvas()
c.Clear(skia.ColorWhite)

p := skia.NewPaint()
p.SetAntialias(true)
p.SetColor(skia.RGB(0xff, 0x6f, 0x00))
c.DrawCircle(256, 256, 120, p)

png, _ := surf.EncodePNG()
os.WriteFile("out.png", png, 0o644)
```

<p align="center"><img src="docs/hello.png" width="320" alt="example output"></p>

## Why this binding

Cgo + Skia is normally painful because Skia is a huge C++ project and
cross-compiling cgo needs a C/C++ toolchain per target. goskia sidesteps both:

- **No Skia build.** It binds the **C API of the prebuilt `libSkiaSharp`**
  shared library (published by the [SkiaSharp](https://github.com/mono/SkiaSharp)
  project on NuGet for every major platform). The C ABI is compiler-neutral, so
  it links cleanly even from MinGW on Windows.
- **One-command cross-compilation.** `goskia build -target os/arch` fetches the
  right native library and drives `go build` with [Zig](https://ziglang.org) as
  the C/C++ cross-compiler. Build a Linux ARM64 binary from your Windows laptop
  with a single command.

It is a thin, idiomatic layer: `Surface`, `Canvas`, `Paint`, `Path`, `Image`,
`Shader`, `Font`, gradients, PNG/JPEG/WebP export, and an optional GPU (OpenGL)
backend.

## Status

| Area | State |
|------|-------|
| Raster rendering → PNG/JPEG/WebP | ✅ working & tested |
| Paths, paints, gradients, transforms, clips | ✅ |
| Filters: blur, drop shadow, lighting, color matrix; path effects | ✅ |
| SkSL runtime effects (shaders, color filters, blenders) | ✅ |
| Text (typefaces, fonts, measuring, drawing) | ✅ |
| PDF / vector output | ✅ |
| GPU (OpenGL) context & surfaces | ✅ bound (needs a current GL context to run) |
| Windows (amd64, arm64) | ✅ amd64 verified end-to-end |
| Linux (amd64, arm64) | ✅ amd64 cross-compiled from Windows and run under WSL |
| macOS (Intel, Apple Silicon) | ✅ universal dylib fetched (build on macOS or with the SDK) |
| Android / iOS / WebAssembly | ⚙️ native libraries fetchable; toolchain notes below |

## Quick start

Requirements: **Go 1.23+** and a C toolchain for your host (any of MinGW-w64,
Clang, GCC, or Zig). For *cross*-compiling you need **Zig** (see below).

```sh
# 1. Fetch the native library for your machine (downloads from NuGet).
go run ./cmd/goskia fetch

# 2. Run an example.
go run ./examples/hello hello.png
```

`go run`/`go build` need `CGO_ENABLED=1`. It is on by default when a C compiler
is present; set it explicitly if needed.

> The native library must sit next to your binary at run time (or be installed
> system-wide). When you build with `goskia build` this is done for you; with a
> plain `go build`, copy `skia/lib/<os>_<arch>/libSkiaSharp.*` next to the
> executable (on Linux, use the SONAME name, e.g. `libSkiaSharp.so.119.0.0`).

## Cross-compilation

Install [Zig](https://ziglang.org/download/) (a single self-contained download).
Put it on `PATH`, set `$GOSKIA_ZIG`, or drop it in `tools/zig/`. Then:

```sh
# Linux ARM64 binary, from any host:
go run ./cmd/goskia build -target linux/arm64 -o bin/app ./examples/hello

# Windows ARM64:
go run ./cmd/goskia build -target windows/amd64 -o bin/app.exe ./examples/hello
```

`goskia build` will:

1. download & stage the target's `libSkiaSharp` (if not already present),
2. set `GOOS`/`GOARCH`/`CGO_ENABLED` and point `CC`/`CXX` at
   `zig cc -target <triple>`,
3. run `go build`, and
4. copy the runtime library next to your output binary.

List what's supported:

```sh
go run ./cmd/goskia targets
```

### Doing it by hand

`goskia build` is just convenience. The equivalent manual invocation:

```sh
go run ./cmd/goskia fetch -target linux/arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
  CC="zig cc -target aarch64-linux-gnu" \
  CXX="zig c++ -target aarch64-linux-gnu" \
  go build -o bin/app ./examples/hello
```

## How it works

```
your Go code
   │  (idiomatic API: Surface, Canvas, Paint, …)
   ▼
package skia  ── cgo ──▶  libSkiaSharp  (C API, prebuilt by SkiaSharp)
   ▲                          │
   └── vendored C headers ────┘   (skia/csrc/include/c/*.h, pinned)
```

- The C-API headers are vendored under `skia/csrc/include/c/` from the exact
  `mono/skia` revision that matches the pinned SkiaSharp version
  (see `internal/buildinfo`).
- `libSkiaSharp` binaries are downloaded from NuGet on demand (not committed).
- On Windows, an import library (`libSkiaSharp.dll.a`) is generated from the
  DLL's export table — in pure Go, so no extra tools are required (a `dlltool`,
  `llvm-dlltool`, or `zig dlltool` is used for the final archive step).

## API tour

```go
skia.Init() // optional; warms global caches

surf, _ := skia.NewRasterSurfaceN32Premul(640, 360)
defer surf.Release()
c := surf.Canvas()

// Gradient fill.
g := skia.NewLinearGradient(
    skia.Point{0, 0}, skia.Point{0, 360},
    []skia.Color{skia.RGB(0x20, 0x2a, 0x44), skia.RGB(0x4a, 0x2c, 0x6d)},
    nil, skia.TileModeClamp)
defer g.Release()
bg := skia.NewPaint(); defer bg.Release()
bg.SetShader(g)
c.DrawRect(skia.RectWH(640, 360), bg)

// Text.
font := skia.NewFont(skia.NewTypeface("", skia.FontStyleBold), 72)
defer font.Release()
ink := skia.NewPaint(); defer ink.Release()
ink.SetAntialias(true); ink.SetColor(skia.ColorWhite)
c.DrawText("goskia", 40, 150, font, ink)

// Export.
png, _ := surf.EncodePNG()
```

Memory: every wrapper has a `Release()` method and a finalizer safety net.
Release explicitly in hot paths; rely on the GC otherwise. A `Canvas` obtained
from a `Surface` is owned by the surface and must not be released.

See `examples/` for `hello` (shapes), `showcase` (text + gradients), `effects`
(filters & path effects), `sksl` (SkSL runtime shader + lighting), `pdf`
(vector PDF), and `smoke` (link test).

## Runtime effects (SkSL)

Compile GPU-style shaders at runtime and use them as a `Shader`, `ColorFilter`,
or `Blender`. Uniforms are set by name (no manual byte packing):

```go
eff, _ := skia.CompileRuntimeShader(`
    uniform float2 iResolution;
    half4 main(float2 p) {
        float2 uv = p / iResolution;
        return half4(half2(uv), 0.4, 1.0);
    }`)
defer eff.Release()
u := eff.NewUniforms()
u.SetVec2("iResolution", 512, 512)
shader := eff.MakeShader(u.Bytes(), nil /*children*/, nil /*localMatrix*/)
paint.SetShader(shader)
```

Lighting image filters (emboss / 3-D shading from an alpha height field) are
also available, e.g. `skia.NewPointLitSpecularImageFilter(...)` — see
`examples/sksl`.

## GPU (OpenGL)

goskia binds Skia's Ganesh GL backend. You supply a current GL context from your
windowing library (GLFW, SDL, …):

```go
iface, _ := skia.NewGLInterface(func(name string) unsafe.Pointer {
    return glfw.GetProcAddress(name) // or your loader
})
ctx, _ := skia.NewGLContext(iface)
defer ctx.Release()

// Render into the window's framebuffer (FBO 0):
surf, _ := skia.NewGPUSurfaceFromFBO(ctx, 0, w, h, samples, 8, skia.GLRGBA8,
    skia.ColorTypeRGBA8888, skia.SurfaceOriginBottomLeft)
// ... draw ...
ctx.FlushAndSubmit(false)
```

## Tools

| Command | Purpose |
|---------|---------|
| `go run ./cmd/goskia fetch [-target os/arch] [-all]` | download & stage native libraries |
| `go run ./cmd/goskia build -target os/arch -o OUT PKG` | cross-compile and stage the runtime lib |
| `go run ./cmd/goskia targets` | list supported targets |
| `go run ./cmd/goskia-fetch …` | the fetch step on its own |

Network downloads honor `HTTP(S)_PROXY`; pass `-proxy URL` to override.

## Platform notes

- **Linux** runtime needs `libfontconfig.so.1` present (used for system fonts).
- **macOS**: the dylib is a universal (x64 + arm64) binary. Cross-compiling to
  macOS from a non-Mac host needs the macOS SDK; building on a Mac is simplest.
- **Android/iOS/WebAssembly**: `goskia fetch -target android/arm64` (etc.)
  stages the native library, but linking needs the platform toolchain (NDK,
  Xcode, Emscripten) rather than Zig.

## Versions

Pinned in `internal/buildinfo`:

- SkiaSharp `3.119.4`
- `mono/skia` headers `7dbfc07`

To move to a newer SkiaSharp, update those constants, re-run
`go run ./cmd/goskia fetch -all`, and re-vendor the C headers from the matching
`mono/skia` commit.

## License

goskia is MIT licensed. It links the BSD-licensed Skia (via the MIT-licensed
SkiaSharp `libSkiaSharp`) and vendors SkiaSharp's C-API headers. See
`LICENSE` and `THIRD_PARTY_NOTICES.md`.
