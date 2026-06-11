// Command goskia-fetch downloads the prebuilt libSkiaSharp native libraries
// from NuGet and stages them under skia/lib/<os>_<arch>/.
//
// Usage:
//
//	go run ./cmd/goskia-fetch                 # host os/arch
//	go run ./cmd/goskia-fetch -target linux/arm64
//	go run ./cmd/goskia-fetch -all            # all desktop targets
//
// Network access uses the standard HTTP(S)_PROXY env vars; -proxy overrides.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/hoonfeng/goskia/internal/nativefetch"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	var (
		targetFlags multiFlag
		all         = flag.Bool("all", false, "fetch all desktop targets (win/linux/mac, amd64+arm64)")
		version     = flag.String("version", "", "SkiaSharp NuGet version (default: pinned)")
		outDir      = flag.String("out", "", "output lib root (default: <module>/skia/lib)")
		cacheDir    = flag.String("cache", "", "nupkg cache dir (default: <module>/.cache)")
		proxy       = flag.String("proxy", "", "HTTP(S) proxy URL (overrides env)")
		force       = flag.Bool("force", false, "re-download and overwrite even if present")
	)
	flag.Var(&targetFlags, "target", "target as os/arch (repeatable); defaults to host")
	flag.Parse()

	var selectors []string
	switch {
	case *all:
		selectors = append(selectors, nativefetch.DesktopTargets...)
	case len(targetFlags) > 0:
		selectors = targetFlags
	default:
		selectors = []string{runtime.GOOS + "/" + runtime.GOARCH}
	}
	sort.Strings(selectors)

	err := nativefetch.Fetch(selectors, nativefetch.Options{
		Version:  *version,
		OutDir:   *outDir,
		CacheDir: *cacheDir,
		Proxy:    *proxy,
		Force:    *force,
		Logf:     func(f string, a ...any) { fmt.Printf(f+"\n", a...) },
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
