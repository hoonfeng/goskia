# goskia — common tasks.
#
# The portable, cross-platform entrypoint is the goskia tool:
#   go run ./cmd/goskia <fetch|build|targets>
# This Makefile is a convenience wrapper for Unix shells.

GO      ?= go
GOOS    := $(shell $(GO) env GOOS)
GOARCH  := $(shell $(GO) env GOARCH)
LIBDIR  := $(CURDIR)/skia/lib/$(GOOS)_$(GOARCH)

.PHONY: help fetch fetch-all targets test examples hello showcase pdf clean

help:
	@echo "targets: fetch fetch-all targets test examples clean"
	@echo "cross-compile: $(GO) run ./cmd/goskia build -target os/arch -o OUT PKG"

fetch:        ## download native lib for the host
	$(GO) run ./cmd/goskia fetch

fetch-all:    ## download native libs for all desktop targets
	$(GO) run ./cmd/goskia fetch -all

targets:
	$(GO) run ./cmd/goskia targets

# Tests need the native library on the loader path. We copy it next to the
# package (found via the test's working directory on Windows) and also export
# the Unix loader paths.
test: fetch
	cp -f $(LIBDIR)/libSkiaSharp.* skia/ 2>/dev/null || true
	CGO_ENABLED=1 \
		LD_LIBRARY_PATH="$(LIBDIR):$$LD_LIBRARY_PATH" \
		DYLD_LIBRARY_PATH="$(LIBDIR):$$DYLD_LIBRARY_PATH" \
		$(GO) test ./skia/...

examples: hello showcase pdf

hello: fetch
	$(GO) run ./cmd/goskia build -o bin/hello ./examples/hello

showcase: fetch
	$(GO) run ./cmd/goskia build -o bin/showcase ./examples/showcase

pdf: fetch
	$(GO) run ./cmd/goskia build -o bin/pdf ./examples/pdf

clean:
	rm -rf bin dist
