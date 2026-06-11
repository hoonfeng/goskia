# goskia

Google [Skia](https://skia.org) 2D 图形库的**地道 Go 语言绑定** — 专为**简单、快速的跨平台编译**而构建。

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

<p align="center"><img src="docs/hello.png" width="320" alt="示例输出"></p>

## 为什么选择这个绑定

通常情况下，Cgo + Skia 的组合很痛苦，因为 Skia 是一个庞大的 C++ 项目，而交叉编译 cgo 需要每个目标平台都有一套 C/C++ 工具链。goskia 避开了这两个问题：

- **无需构建 Skia。** 它绑定的是预构建的 `libSkiaSharp` 共享库的 **C API**（由 [SkiaSharp](https://github.com/mono/SkiaSharp) 项目在 NuGet 上为所有主流平台发布）。C ABI 与编译器无关，即使在 Windows 上通过 MinGW 也能干净地链接。
- **一键交叉编译。** `goskia build -target os/arch` 会自动获取对应的原生库，并使用 [Zig](https://ziglang.org) 作为 C/C++ 交叉编译器驱动 `go build`。在 Windows 笔记本上一条命令就能构建 Linux ARM64 二进制文件。

这是一个轻量且地道的封装层：`Surface`、`Canvas`、`Paint`、`Path`、`Image`、`Shader`、`Font`、渐变、PNG/JPEG/WebP 导出，以及可选的 GPU（OpenGL）后端。

## 状态

| 领域 | 状态 |
|------|------|
| 光栅渲染 → PNG/JPEG/WebP | ✅ 可用且经过测试 |
| 路径、画笔、渐变、变换、裁剪 | ✅ |
| 滤镜：模糊、阴影、光照、色彩矩阵；路径特效 | ✅ |
| SkSL 运行时特效（着色器、颜色滤镜、混合器） | ✅ |
| 文本（字体、字型、测量、绘制） | ✅ |
| PDF / 矢量输出 | ✅ |
| GPU（OpenGL）上下文与表面 | ✅ 已绑定（运行时需要有效的 GL 上下文） |
| Windows（amd64、arm64） | ✅ amd64 端到端已验证 |
| Linux（amd64、arm64） | ✅ amd64 从 Windows 交叉编译并在 WSL 下运行 |
| macOS（Intel、Apple Silicon） | ✅ 获取通用 dylib（在 Mac 上构建或使用 SDK） |
| Android / iOS / WebAssembly | ⚙️ 可获取原生库；工具链说明见下文 |

## 快速开始

前置条件：**Go 1.23+** 和一个适用于你主机的 C 工具链（MinGW-w64、Clang、GCC 或 Zig 均可）。对于**交叉**编译，你需要 **Zig**（见下文）。

```sh
# 1. 获取当前机器的原生库（从 NuGet 下载）。
go run ./cmd/goskia fetch

# 2. 运行示例。
go run ./examples/hello hello.png
```

`go run`/`go build` 需要设置 `CGO_ENABLED=1`。当存在 C 编译器时该选项默认开启；如有需要可手动设置。

> 原生库在运行时必须放在你的二进制文件旁边（或安装到系统目录）。使用 `goskia build` 构建时会自动完成此操作；使用普通的 `go build` 时，请将 `skia/lib/<os>_<arch>/libSkiaSharp.*` 复制到可执行文件旁边（在 Linux 上，使用 SONAME 名称，如 `libSkiaSharp.so.119.0.0`）。

## 交叉编译

安装 [Zig](https://ziglang.org/download/)（一个独立的单个下载文件）。将其加入 `PATH`，设置 `$GOSKIA_ZIG` 环境变量，或将其放入 `tools/zig/` 目录。然后：

```sh
# 从任意主机构建 Linux ARM64 二进制文件：
go run ./cmd/goskia build -target linux/arm64 -o bin/app ./examples/hello

# Windows ARM64：
go run ./cmd/goskia build -target windows/amd64 -o bin/app.exe ./examples/hello
```

`goskia build` 会自动：

1. 下载并准备目标平台的 `libSkiaSharp`（如果尚未存在），
2. 设置 `GOOS`/`GOARCH`/`CGO_ENABLED` 并将 `CC`/`CXX` 指向 `zig cc -target <triple>`，
3. 运行 `go build`，
4. 将运行时库复制到输出二进制文件旁边。

查看支持的目标平台：

```sh
go run ./cmd/goskia targets
```

### 手动操作

`goskia build` 只是一个便捷工具。等效的手动调用方式：

```sh
go run ./cmd/goskia fetch -target linux/arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
  CC="zig cc -target aarch64-linux-gnu" \
  CXX="zig c++ -target aarch64-linux-gnu" \
  go build -o bin/app ./examples/hello
```

## 工作原理

```
你的 Go 代码
   │  (地道 API: Surface, Canvas, Paint, …)
   ▼
package skia  ── cgo ──▶  libSkiaSharp  (C API, 由 SkiaSharp 预构建)
   ▲                          │
   └── 附带的 C 头文件 ────────┘   (skia/csrc/include/c/*.h, 固定版本)
```

- C API 头文件从与固定 SkiaSharp 版本匹配的 `mono/skia` 修订版引入到 `skia/csrc/include/c/` 目录下（参见 `internal/buildinfo`）。
- `libSkiaSharp` 二进制文件按需从 NuGet 下载（不提交到代码仓库）。
- 在 Windows 上，从 DLL 的导出表生成导入库（`libSkiaSharp.dll.a`）—— 纯 Go 实现，无需额外工具（最终归档步骤使用 `dlltool`、`llvm-dlltool` 或 `zig dlltool`）。

## API 概览

```go
skia.Init() // 可选；预热全局缓存

surf, _ := skia.NewRasterSurfaceN32Premul(640, 360)
defer surf.Release()
c := surf.Canvas()

// 渐变填充。
g := skia.NewLinearGradient(
    skia.Point{0, 0}, skia.Point{0, 360},
    []skia.Color{skia.RGB(0x20, 0x2a, 0x44), skia.RGB(0x4a, 0x2c, 0x6d)},
    nil, skia.TileModeClamp)
defer g.Release()
bg := skia.NewPaint(); defer bg.Release()
bg.SetShader(g)
c.DrawRect(skia.RectWH(640, 360), bg)

// 文本。
font := skia.NewFont(skia.NewTypeface("", skia.FontStyleBold), 72)
defer font.Release()
ink := skia.NewPaint(); defer ink.Release()
ink.SetAntialias(true); ink.SetColor(skia.ColorWhite)
c.DrawText("goskia", 40, 150, font, ink)

// 导出。
png, _ := surf.EncodePNG()
```

内存管理：每个封装对象都有 `Release()` 方法和终结器安全网。在热点路径中显式调用 Release；其他情况可依赖 GC。从 `Surface` 获取的 `Canvas` 由 surface 拥有，不应手动释放。

参见 `examples/` 目录下的 `hello`（图形）、`showcase`（文本+渐变）、`effects`（滤镜和路径特效）、`sksl`（SkSL 运行时着色器+光照）、`pdf`（矢量 PDF）和 `smoke`（链接测试）示例。

## 运行时特效（SkSL）

在运行时编译 GPU 风格的着色器，并将其用作 `Shader`、`ColorFilter` 或 `Blender`。统一变量按名称设置（无需手动字节对齐）：

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

光照图像滤镜（浮雕 / 基于 Alpha 高度场的 3D 着色）也可用，例如 `skia.NewPointLitSpecularImageFilter(...)` —— 参见 `examples/sksl`。

## GPU（OpenGL）

goskia 绑定 Skia 的 Ganesh GL 后端。你需要从窗口库（GLFW、SDL 等）提供一个当前的 GL 上下文：

```go
iface, _ := skia.NewGLInterface(func(name string) unsafe.Pointer {
    return glfw.GetProcAddress(name) // 或你自己的加载函数
})
ctx, _ := skia.NewGLContext(iface)
defer ctx.Release()

// 渲染到窗口的帧缓冲区（FBO 0）：
surf, _ := skia.NewGPUSurfaceFromFBO(ctx, 0, w, h, samples, 8, skia.GLRGBA8,
    skia.ColorTypeRGBA8888, skia.SurfaceOriginBottomLeft)
// ... 绘制 ...
ctx.FlushAndSubmit(false)
```

## 工具

| 命令 | 用途 |
|------|------|
| `go run ./cmd/goskia fetch [-target os/arch] [-all]` | 下载并准备原生库 |
| `go run ./cmd/goskia build -target os/arch -o OUT PKG` | 交叉编译并准备运行时库 |
| `go run ./cmd/goskia targets` | 列出支持的目标平台 |
| `go run ./cmd/goskia-fetch …` | 独立的获取步骤 |

网络下载支持 `HTTP(S)_PROXY` 代理；传递 `-proxy URL` 可覆盖设置。

## 平台说明

- **Linux** 运行时需要安装 `libfontconfig.so.1`（用于系统字体）。
- **macOS**：dylib 是通用二进制（x64 + arm64）。从非 Mac 主机交叉编译到 macOS 需要 macOS SDK；在 Mac 上构建最为简单。
- **Android/iOS/WebAssembly**：`goskia fetch -target android/arm64`（等）可以获取原生库，但链接需要平台工具链（NDK、Xcode、Emscripten）而非 Zig。

## 版本

固定版本记录在 `internal/buildinfo` 中：

- SkiaSharp `3.119.4`
- `mono/skia` 头文件 `7dbfc07`

要升级到更新的 SkiaSharp，请更新这些常量，重新运行 `go run ./cmd/goskia fetch -all`，并从对应的 `mono/skia` 提交中重新引入 C 头文件。

## 许可证

goskia 使用 MIT 许可证。它链接了 BSD 许可证的 Skia（通过 MIT 许可证的 SkiaSharp `libSkiaSharp`），并附带了 SkiaSharp 的 C API 头文件。参见 `LICENSE` 和 `THIRD_PARTY_NOTICES.md`。
