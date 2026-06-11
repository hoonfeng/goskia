// Package nativefetch downloads and stages the prebuilt libSkiaSharp native
// libraries from NuGet, and provides the GOOS/GOARCH -> Zig target mapping used
// by the cross-compile orchestrator. It has no cgo dependency.
package nativefetch

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/hoonfeng/goskia/internal/buildinfo"
)

// Target describes how to obtain libSkiaSharp for one GOOS/GOARCH.
type Target struct {
	GOOS, GOARCH string
	Nupkg        string   // NuGet package id (lower-case)
	LibName      string   // output file name
	RIDMatches   []string // zip-entry suffixes tried in order
	ZigTriple    string   // zig `-target` triple ("" = not turnkey via zig)
}

// Targets is the table of supported native targets.
var Targets = []Target{
	{"windows", "amd64", "skiasharp.nativeassets.win32", "libSkiaSharp.dll", []string{"win-x64/native/libSkiaSharp.dll"}, "x86_64-windows-gnu"},
	{"windows", "arm64", "skiasharp.nativeassets.win32", "libSkiaSharp.dll", []string{"win-arm64/native/libSkiaSharp.dll"}, "aarch64-windows-gnu"},
	{"windows", "386", "skiasharp.nativeassets.win32", "libSkiaSharp.dll", []string{"win-x86/native/libSkiaSharp.dll"}, "x86-windows-gnu"},

	{"linux", "amd64", "skiasharp.nativeassets.linux", "libSkiaSharp.so", []string{"linux-x64/native/libSkiaSharp.so"}, "x86_64-linux-gnu"},
	{"linux", "arm64", "skiasharp.nativeassets.linux", "libSkiaSharp.so", []string{"linux-arm64/native/libSkiaSharp.so"}, "aarch64-linux-gnu"},
	{"linux", "arm", "skiasharp.nativeassets.linux", "libSkiaSharp.so", []string{"linux-arm/native/libSkiaSharp.so"}, "arm-linux-gnueabihf"},

	{"darwin", "amd64", "skiasharp.nativeassets.macos", "libSkiaSharp.dylib", []string{"osx-x64/native/libSkiaSharp.dylib", "osx/native/libSkiaSharp.dylib"}, "x86_64-macos-none"},
	{"darwin", "arm64", "skiasharp.nativeassets.macos", "libSkiaSharp.dylib", []string{"osx-arm64/native/libSkiaSharp.dylib", "osx/native/libSkiaSharp.dylib"}, "aarch64-macos-none"},

	// Android ships .so files inside an AAR (a zip) under jni/<abi>/.
	{"android", "arm64", "skiasharp.nativeassets.android", "libSkiaSharp.so", []string{"arm64-v8a/libSkiaSharp.so"}, ""},
	{"android", "arm", "skiasharp.nativeassets.android", "libSkiaSharp.so", []string{"armeabi-v7a/libSkiaSharp.so"}, ""},
	{"android", "amd64", "skiasharp.nativeassets.android", "libSkiaSharp.so", []string{"x86_64/libSkiaSharp.so"}, ""},
	{"android", "386", "skiasharp.nativeassets.android", "libSkiaSharp.so", []string{"x86/libSkiaSharp.so"}, ""},
}

// DesktopTargets is the turnkey cross-compile set selected by -all.
var DesktopTargets = []string{
	"windows/amd64", "windows/arm64",
	"linux/amd64", "linux/arm64",
	"darwin/amd64", "darwin/arm64",
}

// Find returns the Target for a GOOS/GOARCH.
func Find(goos, goarch string) (Target, bool) {
	for _, t := range Targets {
		if t.GOOS == goos && t.GOARCH == goarch {
			return t, true
		}
	}
	return Target{}, false
}

// Options configures a fetch run.
type Options struct {
	Version  string // SkiaSharp NuGet version ("" = pinned default)
	OutDir   string // lib root ("" = <module>/skia/lib)
	CacheDir string // nupkg cache ("" = <module>/.cache)
	Proxy    string // proxy URL override ("" = use env)
	Force    bool
	Logf     func(format string, args ...any)
}

func (o *Options) defaults() error {
	if o.Version == "" {
		o.Version = buildinfo.SkiaSharpVersion
	}
	root, err := ModuleRoot()
	if err != nil {
		return err
	}
	if o.OutDir == "" {
		o.OutDir = filepath.Join(root, "skia", "lib")
	}
	if o.CacheDir == "" {
		o.CacheDir = filepath.Join(root, ".cache")
	}
	if o.Logf == nil {
		o.Logf = func(string, ...any) {}
	}
	return nil
}

// Fetch stages the native libraries for the given "os/arch" selectors.
func Fetch(selectors []string, opts Options) error {
	if err := opts.defaults(); err != nil {
		return err
	}
	client := newClient(opts.Proxy)
	var firstErr error
	for _, sel := range selectors {
		goos, goarch, ok := splitTarget(sel)
		if !ok {
			err := fmt.Errorf("invalid target %q (want os/arch)", sel)
			opts.Logf("FAIL %s: %v", sel, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		t, ok := Find(goos, goarch)
		if !ok {
			err := fmt.Errorf("no known SkiaSharp native package for %s/%s", goos, goarch)
			opts.Logf("skip %s/%s: %v", goos, goarch, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if _, err := fetchOne(client, t, opts); err != nil {
			opts.Logf("FAIL %s/%s: %v", goos, goarch, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// EnsureLib makes sure the native library for goos/goarch is staged and returns
// the directory containing it.
func EnsureLib(goos, goarch string, opts Options) (string, error) {
	if err := opts.defaults(); err != nil {
		return "", err
	}
	t, ok := Find(goos, goarch)
	if !ok {
		return "", fmt.Errorf("no known SkiaSharp native package for %s/%s", goos, goarch)
	}
	return fetchOne(newClient(opts.Proxy), t, opts)
}

func fetchOne(client *http.Client, t Target, opts Options) (string, error) {
	destDir := filepath.Join(opts.OutDir, t.GOOS+"_"+t.GOARCH)
	destLib := filepath.Join(destDir, t.LibName)
	if !opts.Force {
		if _, err := os.Stat(destLib); err == nil {
			opts.Logf("ok   %s/%s: already present", t.GOOS, t.GOARCH)
			return destDir, maybeImportLib(t, destDir, destLib, opts)
		}
	}

	nupkg, err := downloadNupkg(client, t.Nupkg, opts.Version, opts.CacheDir, opts.Logf)
	if err != nil {
		return "", err
	}
	zr, err := zip.OpenReader(nupkg)
	if err != nil {
		return "", fmt.Errorf("open nupkg: %w", err)
	}
	defer zr.Close()

	var entry *zip.File
	for _, want := range t.RIDMatches {
		for _, f := range zr.File {
			if strings.HasSuffix(strings.ToLower(filepath.ToSlash(f.Name)), strings.ToLower(want)) {
				entry = f
				break
			}
		}
		if entry != nil {
			break
		}
	}
	if entry == nil {
		return "", fmt.Errorf("library not found in %s %s (looked for %v)", t.Nupkg, opts.Version, t.RIDMatches)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	if err := extractZipFile(entry, destLib); err != nil {
		return "", err
	}
	opts.Logf("ok   %s/%s: %s", t.GOOS, t.GOARCH, entry.Name)
	return destDir, maybeImportLib(t, destDir, destLib, opts)
}

func maybeImportLib(t Target, destDir, destLib string, opts Options) error {
	if t.GOOS != "windows" {
		return nil
	}
	dllA := filepath.Join(destDir, "libSkiaSharp.dll.a") // MinGW gcc convention
	plainA := filepath.Join(destDir, "libSkiaSharp.a")   // clang/lld (Zig) convention
	if !opts.Force {
		_, e1 := os.Stat(dllA)
		_, e2 := os.Stat(plainA)
		if e1 == nil && e2 == nil {
			return nil
		}
	}
	if err := generateImportLib(destLib, destDir, t.GOARCH); err != nil {
		return fmt.Errorf("import lib: %w", err)
	}
	// Different linkers look for different import-library names for -lSkiaSharp:
	// MinGW gcc wants libSkiaSharp.dll.a, clang/lld (Zig) wants libSkiaSharp.a.
	// Provide both from the same generated archive.
	if err := copyFileNF(dllA, plainA); err != nil {
		return fmt.Errorf("import lib alias: %w", err)
	}
	opts.Logf("     %s/%s: generated libSkiaSharp.dll.a (+ .a alias)", t.GOOS, t.GOARCH)
	return nil
}

func copyFileNF(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

func downloadNupkg(client *http.Client, id, version, cacheDir string, logf func(string, ...any)) (string, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(cacheDir, fmt.Sprintf("%s.%s.nupkg", id, version))
	if fi, err := os.Stat(path); err == nil && fi.Size() > 0 {
		return path, nil
	}
	u := fmt.Sprintf("https://api.nuget.org/v3-flatcontainer/%s/%s/%s.%s.nupkg", id, version, id, version)
	logf("     downloading %s ...", u)
	resp, err := client.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", u, resp.Status)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return path, os.Rename(tmp, path)
}

func extractZipFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	tmp := dest + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}

// --- Windows import library -------------------------------------------------

func generateImportLib(dllPath, destDir, goarch string) error {
	names, err := dllExports(dllPath)
	if err != nil {
		return fmt.Errorf("read exports: %w", err)
	}
	if len(names) == 0 {
		return fmt.Errorf("no exported symbols found")
	}
	var def bytes.Buffer
	fmt.Fprintf(&def, "LIBRARY %s\nEXPORTS\n", filepath.Base(dllPath))
	for _, n := range names {
		fmt.Fprintf(&def, "%s\n", n)
	}
	defPath := filepath.Join(destDir, "libSkiaSharp.def")
	if err := os.WriteFile(defPath, def.Bytes(), 0o644); err != nil {
		return err
	}
	implib := filepath.Join(destDir, "libSkiaSharp.dll.a")
	cands := dlltoolCandidates(goarch, defPath, implib, filepath.Base(dllPath))
	if len(cands) == 0 {
		return fmt.Errorf("no dlltool found (install MinGW binutils, LLVM, or Zig)")
	}
	var errs []string
	for _, c := range cands {
		if out, err := exec.Command(c.tool, c.args...).CombinedOutput(); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v: %s", filepath.Base(c.tool), err, strings.TrimSpace(string(out))))
		}
	}
	return fmt.Errorf("all dlltool candidates failed:\n      %s", strings.Join(errs, "\n      "))
}

type dlltoolCand struct {
	tool string
	args []string
}

// dlltoolCandidates returns dlltool invocations to try in order. binutils
// dlltool handles x86 well but not arm64; llvm-dlltool and `zig dlltool` handle
// every architecture, so they are included as fallbacks.
func dlltoolCandidates(goarch, def, out, dllName string) []dlltoolCand {
	machine := map[string]string{"amd64": "i386:x86-64", "386": "i386", "arm64": "arm64", "arm": "arm"}[goarch]
	common := []string{"-m", machine, "-d", def, "-l", out, "-D", dllName}
	var c []dlltoolCand
	if env := os.Getenv("GOSKIA_DLLTOOL"); env != "" {
		c = append(c, dlltoolCand{env, common})
	}
	for _, name := range []string{"x86_64-w64-mingw32-dlltool", "aarch64-w64-mingw32-dlltool", "dlltool", "llvm-dlltool"} {
		if p, err := exec.LookPath(name); err == nil {
			c = append(c, dlltoolCand{p, common})
		}
	}
	if z := findZig(); z != "" {
		c = append(c, dlltoolCand{z, append([]string{"dlltool"}, common...)})
	}
	return c
}

// findZig locates a zig executable usable as a fallback dlltool/cross-compiler.
func findZig() string {
	if e := os.Getenv("GOSKIA_ZIG"); e != "" {
		return e
	}
	if p, err := exec.LookPath("zig"); err == nil {
		return p
	}
	if root, err := ModuleRoot(); err == nil {
		name := "zig"
		if runtime.GOOS == "windows" {
			name = "zig.exe"
		}
		cand := filepath.Join(root, "tools", "zig", name)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return ""
}

// dllExports parses a PE file and returns its exported symbol names. It uses no
// external tools so the fetch works on any host OS.
func dllExports(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 0x40 || data[0] != 'M' || data[1] != 'Z' {
		return nil, fmt.Errorf("not a PE file")
	}
	peOff := int64(binary.LittleEndian.Uint32(data[0x3c:]))
	if peOff+24 > int64(len(data)) || string(data[peOff:peOff+4]) != "PE\x00\x00" {
		return nil, fmt.Errorf("bad PE signature")
	}
	numSections := int(binary.LittleEndian.Uint16(data[peOff+6:]))
	optSize := int64(binary.LittleEndian.Uint16(data[peOff+20:]))
	optOff := peOff + 24
	magic := binary.LittleEndian.Uint16(data[optOff:])

	var ddOff int64
	switch magic {
	case 0x20b: // PE32+
		ddOff = optOff + 112
	case 0x10b: // PE32
		ddOff = optOff + 96
	default:
		return nil, fmt.Errorf("unknown optional header magic 0x%x", magic)
	}
	exportRVA := binary.LittleEndian.Uint32(data[ddOff:])
	if exportRVA == 0 {
		return nil, nil
	}

	type section struct{ va, size, raw uint32 }
	secStart := optOff + optSize
	var sections []section
	for i := 0; i < numSections; i++ {
		off := secStart + int64(i)*40
		if off+40 > int64(len(data)) {
			break
		}
		vs := binary.LittleEndian.Uint32(data[off+8:])
		va := binary.LittleEndian.Uint32(data[off+12:])
		raw := binary.LittleEndian.Uint32(data[off+20:])
		sections = append(sections, section{va, vs, raw})
	}
	rvaToOff := func(rva uint32) (int64, bool) {
		for _, s := range sections {
			size := s.size
			if size == 0 {
				size = 1
			}
			if rva >= s.va && rva < s.va+size {
				return int64(s.raw + (rva - s.va)), true
			}
		}
		return 0, false
	}

	expOff, ok := rvaToOff(exportRVA)
	if !ok {
		return nil, fmt.Errorf("export directory RVA not mapped")
	}
	numNames := binary.LittleEndian.Uint32(data[expOff+24:])
	namesRVA := binary.LittleEndian.Uint32(data[expOff+32:])
	namesOff, ok := rvaToOff(namesRVA)
	if !ok {
		return nil, fmt.Errorf("name table RVA not mapped")
	}
	names := make([]string, 0, numNames)
	for i := uint32(0); i < numNames; i++ {
		nameRVA := binary.LittleEndian.Uint32(data[namesOff+int64(i)*4:])
		nameOff, ok := rvaToOff(nameRVA)
		if !ok {
			continue
		}
		end := nameOff
		for end < int64(len(data)) && data[end] != 0 {
			end++
		}
		names = append(names, string(data[nameOff:end]))
	}
	sort.Strings(names)
	return names, nil
}

// --- helpers ----------------------------------------------------------------

func newClient(proxyOverride string) *http.Client {
	tr := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if proxyOverride != "" {
		if u, err := url.Parse(proxyOverride); err == nil {
			tr.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Transport: tr, Timeout: 10 * time.Minute}
}

func splitTarget(s string) (goos, goarch string, ok bool) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// SOName reads the DT_SONAME of an ELF shared object (e.g.
// "libSkiaSharp.so.119.0.0"). It returns "" if the file is not an ELF with a
// soname. Only little-endian ELF (the targets we support) is handled.
func SOName(path string) (string, error) {
	d, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(d) < 64 || string(d[:4]) != "\x7fELF" {
		return "", nil
	}
	is64 := d[4] == 2

	var shoff int64
	var shentsize, shnum int
	if is64 {
		shoff = int64(binary.LittleEndian.Uint64(d[0x28:]))
		shentsize = int(binary.LittleEndian.Uint16(d[0x3a:]))
		shnum = int(binary.LittleEndian.Uint16(d[0x3c:]))
	} else {
		shoff = int64(binary.LittleEndian.Uint32(d[0x20:]))
		shentsize = int(binary.LittleEndian.Uint16(d[0x2e:]))
		shnum = int(binary.LittleEndian.Uint16(d[0x30:]))
	}

	type sec struct {
		typ, link    uint32
		offset, size int64
	}
	secs := make([]sec, 0, shnum)
	for i := 0; i < shnum; i++ {
		o := shoff + int64(i*shentsize)
		if o+int64(shentsize) > int64(len(d)) {
			break
		}
		var s sec
		s.typ = binary.LittleEndian.Uint32(d[o+4:])
		if is64 {
			s.offset = int64(binary.LittleEndian.Uint64(d[o+24:]))
			s.size = int64(binary.LittleEndian.Uint64(d[o+32:]))
			s.link = binary.LittleEndian.Uint32(d[o+40:])
		} else {
			s.offset = int64(binary.LittleEndian.Uint32(d[o+16:]))
			s.size = int64(binary.LittleEndian.Uint32(d[o+20:]))
			s.link = binary.LittleEndian.Uint32(d[o+24:])
		}
		secs = append(secs, s)
	}

	const shtDynamic = 6
	const dtNull, dtSoname = 0, 14
	for _, s := range secs {
		if s.typ != shtDynamic {
			continue
		}
		if int(s.link) >= len(secs) {
			continue
		}
		strOff := secs[s.link].offset
		entSize := int64(8)
		if is64 {
			entSize = 16
		}
		for off := s.offset; off+entSize <= s.offset+s.size; off += entSize {
			var tag, val int64
			if is64 {
				tag = int64(binary.LittleEndian.Uint64(d[off:]))
				val = int64(binary.LittleEndian.Uint64(d[off+8:]))
			} else {
				tag = int64(int32(binary.LittleEndian.Uint32(d[off:])))
				val = int64(binary.LittleEndian.Uint32(d[off+4:]))
			}
			if tag == dtNull {
				break
			}
			if tag == dtSoname {
				p := strOff + val
				end := p
				for end < int64(len(d)) && d[end] != 0 {
					end++
				}
				return string(d[p:end]), nil
			}
		}
	}
	return "", nil
}

// ModuleRoot walks up from the working directory to find the go.mod directory.
func ModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from working directory upward")
		}
		dir = parent
	}
}
