// Command goskia is the all-in-one helper for building goskia programs across
// platforms. It fetches the right native library and drives `go build` with a
// cross C/C++ toolchain (Zig) so cross-compilation is a single command.
//
// Usage:
//
//	goskia targets                              list supported targets
//	goskia fetch [-target os/arch ...] [-all]   stage native libraries
//	goskia build -target os/arch [-o out] PKG   cross-compile a package
//
// Examples:
//
//	go run ./cmd/goskia build -target linux/arm64 -o bin/app ./examples/hello
//	go run ./cmd/goskia build -target windows/amd64 -o bin/app.exe ./examples/hello
//
// For cross targets a Zig toolchain is required. goskia looks for zig in
// $GOSKIA_ZIG, then on PATH, then in <module>/tools/zig/.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hoonfeng/goskia/internal/nativefetch"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "targets":
		cmdTargets()
	case "fetch":
		err = cmdFetch(os.Args[2:])
	case "build":
		err = cmdBuild(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `goskia - cross-platform build helper for Skia bindings

Commands:
  targets                              list supported targets
  fetch [-target os/arch ...] [-all]   download & stage native libraries
  build -target os/arch [-o out] PKG   cross-compile (fetches lib, uses Zig)

Run "goskia <command> -h" for command flags.
`)
}

func cmdTargets() {
	host := runtime.GOOS + "/" + runtime.GOARCH
	fmt.Println("supported targets (* = turnkey cross-compile via Zig):")
	for _, t := range nativefetch.Targets {
		mark := " "
		if t.ZigTriple != "" {
			mark = "*"
		}
		sel := t.GOOS + "/" + t.GOARCH
		hostMark := ""
		if sel == host {
			hostMark = "  (host)"
		}
		fmt.Printf("  %s %-16s %s%s\n", mark, sel, t.ZigTriple, hostMark)
	}
}

func cmdFetch(args []string) error {
	opts, selectors, err := parseFetchArgs(args)
	if err != nil {
		return err
	}
	opts.Logf = func(f string, a ...any) { fmt.Printf(f+"\n", a...) }
	return nativefetch.Fetch(selectors, opts)
}

func parseFetchArgs(args []string) (nativefetch.Options, []string, error) {
	var selectors []string
	var opts nativefetch.Options
	all := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-all", "--all":
			all = true
		case "-target", "--target":
			i++
			if i >= len(args) {
				return opts, nil, fmt.Errorf("-target needs a value")
			}
			selectors = append(selectors, args[i])
		case "-proxy":
			i++
			opts.Proxy = args[i]
		case "-force":
			opts.Force = true
		default:
			return opts, nil, fmt.Errorf("unknown fetch flag %q", args[i])
		}
	}
	if all {
		selectors = append(selectors, nativefetch.DesktopTargets...)
	}
	if len(selectors) == 0 {
		selectors = []string{runtime.GOOS + "/" + runtime.GOARCH}
	}
	return opts, selectors, nil
}

func cmdBuild(args []string) error {
	// Parse our flags; everything after is forwarded to `go build`.
	var target, output string
	copyLib := true
	var rest []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-target":
			i++
			if i >= len(args) {
				return fmt.Errorf("-target needs a value")
			}
			target = args[i]
		case args[i] == "-o":
			i++
			if i >= len(args) {
				return fmt.Errorf("-o needs a value")
			}
			output = args[i]
		case args[i] == "-no-copy-lib":
			copyLib = false
		default:
			rest = append(rest, args[i])
		}
	}
	if target == "" {
		target = runtime.GOOS + "/" + runtime.GOARCH
	}
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid -target %q (want os/arch)", target)
	}
	goos, goarch := parts[0], parts[1]

	t, ok := nativefetch.Find(goos, goarch)
	if !ok {
		return fmt.Errorf("unsupported target %s/%s (see `goskia targets`)", goos, goarch)
	}

	// Ensure the native library is staged.
	libDir, err := nativefetch.EnsureLib(goos, goarch, nativefetch.Options{
		Logf: func(f string, a ...any) { fmt.Printf(f+"\n", a...) },
	})
	if err != nil {
		return err
	}

	env := append(os.Environ(), "CGO_ENABLED=1", "GOOS="+goos, "GOARCH="+goarch)

	// macOS uses an @loader_path rpath (so the dylib is found next to the
	// binary). cgo's flag-security policy rejects the '@' character in #cgo
	// LDFLAGS, so it must be explicitly allowed.
	if goos == "darwin" {
		env = append(env, "CGO_LDFLAGS_ALLOW=-Wl,-rpath,@loader_path")
	}

	host := runtime.GOOS + "/" + runtime.GOARCH
	cross := target != host
	if cross || os.Getenv("GOSKIA_FORCE_ZIG") != "" {
		if t.ZigTriple == "" {
			return fmt.Errorf("cross-compiling to %s/%s is not turnkey via Zig; build on that platform or set CC/CXX yourself", goos, goarch)
		}
		zig, err := findZig()
		if err != nil {
			return err
		}
		env = append(env,
			"CC="+zig+" cc -target "+t.ZigTriple,
			"CXX="+zig+" c++ -target "+t.ZigTriple,
		)
		fmt.Printf("using zig: %s (-target %s)\n", zig, t.ZigTriple)
	}

	goArgs := []string{"build"}
	if output != "" {
		goArgs = append(goArgs, "-o", output)
	}
	goArgs = append(goArgs, rest...)

	fmt.Printf("building %s/%s: go %s\n", goos, goarch, strings.Join(goArgs, " "))
	cmd := exec.Command("go", goArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	if copyLib && output != "" {
		if err := stageRuntimeLib(libDir, t.LibName, output); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not stage runtime library: %v\n", err)
		} else {
			fmt.Printf("staged %s next to %s\n", t.LibName, output)
		}
	}
	return nil
}

// stageRuntimeLib copies the shared library next to the output binary so it can
// be found by the OS loader at runtime. For ELF shared objects the file is
// named by its SONAME (e.g. libSkiaSharp.so.119.0.0), which is what the dynamic
// loader actually requests.
func stageRuntimeLib(libDir, libName, output string) error {
	src := filepath.Join(libDir, libName)
	dstDir := filepath.Dir(output)
	if dstDir == "" {
		dstDir = "."
	}
	dstName := libName
	if strings.Contains(libName, ".so") {
		if so, err := nativefetch.SOName(src); err == nil && so != "" {
			dstName = so
		}
	}
	return copyFile(src, filepath.Join(dstDir, dstName))
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}

// findZig locates a zig executable.
func findZig() (string, error) {
	if env := os.Getenv("GOSKIA_ZIG"); env != "" {
		return env, nil
	}
	if p, err := exec.LookPath("zig"); err == nil {
		return p, nil
	}
	root, err := nativefetch.ModuleRoot()
	if err == nil {
		name := "zig"
		if runtime.GOOS == "windows" {
			name = "zig.exe"
		}
		cand := filepath.Join(root, "tools", "zig", name)
		if _, err := os.Stat(cand); err == nil {
			return cand, nil
		}
	}
	return "", fmt.Errorf("zig not found: install it, set $GOSKIA_ZIG, or place it in <module>/tools/zig/ (https://ziglang.org/download/)")
}
