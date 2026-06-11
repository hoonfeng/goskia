# 在 macOS 上编译 goskia

本手册介绍如何为 macOS 构建使用 goskia 的程序,涵盖**在 Mac 上原生编译**、**通用二进制(Intel + Apple Silicon)**、**运行时库定位**、**代码签名**,以及**从非 Mac 主机交叉编译**。

## 0. 背景:macOS 原生库长什么样

goskia 在 macOS 上链接 SkiaSharp 的预编译 `libSkiaSharp.dylib`。它有两个关键特性(本仓库已实测确认):

- **通用二进制(universal / fat)**:同一个 `.dylib` 同时包含 `x86_64` 和 `arm64` 两个切片,Intel 与 Apple Silicon 共用一份文件。
- **`install_name = @rpath/libSkiaSharp.dylib`**:链接进你的程序后,加载器会按 `@rpath` 去查找它。goskia 在链接时设置了 rpath `@loader_path`,因此只要 dylib 和可执行文件**放在同一目录**即可被找到。

它依赖的都是 macOS 系统自带的框架(CoreFoundation、CoreGraphics、CoreText、Cocoa、Metal、MetalKit、Foundation、libc++、libSystem 等),**无需额外安装**任何库(不像 Linux 需要 fontconfig)。

## 1. 在 Mac 上原生编译(推荐)

前置:

- 安装 Xcode Command Line Tools(提供 clang + macOS SDK + 系统框架):
  ```sh
  xcode-select --install
  ```
- Go 1.23+(`CGO_ENABLED` 默认开启,因为有 clang)。

一条命令完成(自动拉取原生库 + 链接 + 把 dylib 拷到产物旁):

```sh
# Apple Silicon:
go run ./cmd/goskia build -target darwin/arm64 -o bin/app ./examples/hello
# Intel:
go run ./cmd/goskia build -target darwin/amd64 -o bin/app ./examples/hello

./bin/app out.png
```

> 也可以直接用原生工具链。注意 cgo 的安全策略会拒绝 rpath 里的 `@` 字符,所以手动
> `go build` 时需要显式放行:
>
> ```sh
> go run ./cmd/goskia fetch -target darwin/arm64
> export CGO_LDFLAGS_ALLOW='-Wl,-rpath,@loader_path'
> CGO_ENABLED=1 go build -o bin/app ./examples/hello
> cp skia/lib/darwin_arm64/libSkiaSharp.dylib bin/    # 放到可执行文件旁
> ```
>
> `goskia build` 会自动设置 `CGO_LDFLAGS_ALLOW` 并完成拷贝,省去这两步。

## 2. 通用二进制(Intel + Apple Silicon 合一)

因为原生 dylib 本身是通用的,你可以分别编出两个架构再用 `lipo` 合并成一个通用可执行文件:

```sh
go run ./cmd/goskia build -target darwin/amd64 -o bin/app-amd64 ./examples/hello
go run ./cmd/goskia build -target darwin/arm64 -o bin/app-arm64 ./examples/hello
lipo -create -output bin/app bin/app-amd64 bin/app-arm64
lipo -archs bin/app           # 应输出: x86_64 arm64
cp skia/lib/darwin_arm64/libSkiaSharp.dylib bin/   # dylib 已是通用的,放一份即可
```

在同一台 Mac 上跨架构编译没有问题:Go 的交叉目标 + clang 的 `-arch` 都受支持(zig 也可)。

## 3. 运行时:dylib 如何被找到

加载器解析顺序的关键点:

- 你的程序记录了对 `@rpath/libSkiaSharp.dylib` 的依赖;
- goskia 链接时加入了 rpath `@loader_path`(即"可执行文件所在目录");
- 因此**把 `libSkiaSharp.dylib` 放在可执行文件同目录**即可。`goskia build` 会自动拷贝。

调试期临时指定路径:

```sh
DYLD_LIBRARY_PATH="$PWD/skia/lib/darwin_arm64" ./bin/app
```

若你移动/改名了库,或需要不同布局,可用 `install_name_tool` 调整:

```sh
otool -L bin/app                              # 查看实际依赖与路径
install_name_tool -add_rpath @loader_path bin/app
# 修改记录的依赖路径:
install_name_tool -change @rpath/libSkiaSharp.dylib @loader_path/libSkiaSharp.dylib bin/app
```

## 4. 代码签名与 Gatekeeper(Apple Silicon 必读)

Apple Silicon 上**所有可执行代码都必须有签名**(至少是 ad-hoc 签名),否则会被系统直接杀掉(`Killed: 9`)。

- NuGet 上的 `libSkiaSharp.dylib` 通常已带签名。但只要你用 `install_name_tool` 改过它,签名就会失效,**必须重新签名**:
  ```sh
  codesign --force --sign - bin/libSkiaSharp.dylib   # ad-hoc 重签
  codesign --force --sign - bin/app
  ```
- 从网络下载的文件可能带"隔离"属性,运行前清除:
  ```sh
  xattr -dr com.apple.quarantine bin/
  ```
- **对外分发**时需用 Developer ID 证书签名并做公证(notarization):
  ```sh
  codesign --force --options runtime --sign "Developer ID Application: 你的名字 (TEAMID)" bin/libSkiaSharp.dylib bin/app
  xcrun notarytool submit app.zip --keychain-profile "AC_PASSWORD" --wait
  xcrun stapler staple bin/app
  ```

## 5. 从非 Mac(Windows / Linux)交叉编译到 macOS

跨平台编译到 macOS 是最棘手的一档,因为链接 Go 的 darwin 运行时需要 macOS SDK(`libSystem`)。按可靠程度排序:

### 方案 A(推荐):在 macOS 上构建 / 用 macOS CI

最省心。见第 6 节的 GitHub Actions 示例(`macos-14` = Apple Silicon,`macos-13` = Intel)。

### 方案 B:Zig(不推荐,实测不可行)

`goskia build` 交叉到 darwin 时会尝试 `zig cc -target {x86_64,aarch64}-macos-none`。

> **实测结论(本仓库,Zig 0.16,从 Windows 交叉)**:能通过编译,但**链接阶段失败**
> —— Go 的 darwin 运行时会链接 `-lresolv` 和系统框架(如 `CoreFoundation`),而 Zig
> 自带的 macOS libc 桩**不包含** `libresolv`/这些框架:
> `error: unable to find dynamic system library 'resolv'`。
>
> 因此**仅靠 Zig 无法从非 Mac 交叉编译到 macOS**。除非你给 Zig 指定一个完整的 macOS
> SDK 作为 sysroot(`--sysroot`,高级用法),否则请用方案 A 或方案 C。

### 方案 C:osxcross(在 Linux 上准备 macOS SDK + clang)

1. 按 [osxcross](https://github.com/tpoechtrager/osxcross) 文档准备 SDK 与工具链(需自备合法的 macOS SDK)。
2. 用 osxcross 的 clang 作为 cgo 编译器:
   ```sh
   go run ./cmd/goskia fetch -target darwin/arm64
   GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
     CC=oa64-clang CXX=oa64-clang++ \
     go build -o bin/app ./examples/hello
   cp skia/lib/darwin_arm64/libSkiaSharp.dylib bin/
   ```
   (Intel 用 `o64-clang` / `o64-clang++`,`GOARCH=amd64`。)

## 6. GitHub Actions(macOS CI)示例

```yaml
name: build-macos
on: [push]
jobs:
  build:
    strategy:
      matrix:
        include:
          - { runner: macos-14, arch: arm64 }   # Apple Silicon
          - { runner: macos-13, arch: amd64 }   # Intel
    runs-on: ${{ matrix.runner }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - name: Build
        run: |
          go run ./cmd/goskia build -target darwin/${{ matrix.arch }} -o bin/app ./examples/hello
          codesign --force --sign - bin/libSkiaSharp.dylib bin/app
      - uses: actions/upload-artifact@v4
        with:
          name: app-darwin-${{ matrix.arch }}
          path: bin/
```

## 7. 常见错误排查

| 现象 | 原因与修复 |
|------|-----------|
| `dyld: Library not loaded: @rpath/libSkiaSharp.dylib` | dylib 不在可执行文件旁。用 `goskia build`(自动拷贝),或手动 `cp libSkiaSharp.dylib` 到产物目录;调试期可用 `DYLD_LIBRARY_PATH`。 |
| `Killed: 9`(Apple Silicon) | 代码未签名或被隔离。`codesign --force --sign - <dylib> <bin>` + `xattr -dr com.apple.quarantine`。 |
| `code signature in ... is not valid` | 改过 dylib(如 `install_name_tool`)导致签名失效,需 ad-hoc 重签。 |
| `building for 'macOS-arm64' but attempting to link with 'x86_64'` | 架构不匹配。确认 `GOARCH` 与目标一致;dylib 是通用的所以不是它的问题。 |
| 交叉编译时报缺少 `libSystem` / framework | Zig 桩不可用。改用 macOS / CI(方案 A)或 osxcross(方案 C)。 |

---

更多通用说明见仓库根目录的 [README](../README.md)。
