GO ?= go
BINARY := a365
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/sozercan/a365cli/internal/version.Version=$(VERSION) -X github.com/sozercan/a365cli/internal/version.Commit=$(COMMIT)"
TARGET_GOOS := $(or $(GOOS),$(shell $(GO) env GOOS))
TARGET_GOARCH := $(or $(GOARCH),$(shell $(GO) env GOARCH))
TARGET_GOEXE := $(shell GOOS=$(TARGET_GOOS) GOARCH=$(TARGET_GOARCH) $(GO) env GOEXE)
HOST_GOOS := $(shell $(GO) env GOHOSTOS)
HOST_GOARCH := $(shell $(GO) env GOHOSTARCH)
GOENV := GOOS=$(TARGET_GOOS) GOARCH=$(TARGET_GOARCH)
A365_DETECTED_CODESIGN_IDENTITY := $(shell if [ "$$(uname)" = "Darwin" ] && command -v security >/dev/null 2>&1; then security find-identity -v -p codesigning 2>/dev/null | awk '/"Apple Development:/ { print $$2; exit }'; fi)
A365_CODESIGN_IDENTITY ?= $(A365_DETECTED_CODESIGN_IDENTITY)

define codesign_binary
	@if [ "$(TARGET_GOOS)" = "darwin" ] && [ "$$(uname)" = "Darwin" ] && command -v codesign >/dev/null 2>&1; then \
		identity="$(A365_CODESIGN_IDENTITY)"; \
		if [ -n "$$identity" ]; then \
			echo "Signing $(1) with configured codesigning identity"; \
			codesign --force --sign "$$identity" "$(1)"; \
		else \
			echo "Signing $(1) ad-hoc (set A365_CODESIGN_IDENTITY to use a stable cert)"; \
			codesign --force --sign - "$(1)" 2>/dev/null || true; \
		fi; \
	fi
endef

define install_binary
	@set -e; \
	gobin="$$( $(GOENV) $(GO) env GOBIN )"; \
	gopath="$$( $(GOENV) $(GO) env GOPATH )"; \
	if [ -n "$$gobin" ]; then \
		if [ "$(TARGET_GOOS)_$(TARGET_GOARCH)" != "$(HOST_GOOS)_$(HOST_GOARCH)" ]; then \
			echo "go install: cannot install cross-compiled binaries when GOBIN is set" >&2; \
			exit 1; \
		fi; \
		bin_dir="$$gobin"; \
	else \
		path_sep=":"; \
		if [ "$(HOST_GOOS)" = "windows" ]; then path_sep=";"; fi; \
		first_gopath="$${gopath%%$${path_sep}*}"; \
		bin_dir="$$first_gopath/bin"; \
		if [ "$(TARGET_GOOS)_$(TARGET_GOARCH)" != "$(HOST_GOOS)_$(HOST_GOARCH)" ]; then \
			bin_dir="$$bin_dir/$(TARGET_GOOS)_$(TARGET_GOARCH)"; \
		fi; \
	fi; \
	install_binary="$$bin_dir/$(BINARY)$(TARGET_GOEXE)"; \
	mkdir -p "$$bin_dir"; \
	$(1) $(GOENV) $(GO) build $(LDFLAGS) -o "$$install_binary" .; \
	if [ "$(TARGET_GOOS)" = "darwin" ] && [ "$$(uname)" = "Darwin" ] && command -v codesign >/dev/null 2>&1; then \
		identity="$(A365_CODESIGN_IDENTITY)"; \
		if [ -n "$$identity" ]; then \
			echo "Signing $$install_binary with configured codesigning identity"; \
			codesign --force --sign "$$identity" "$$install_binary"; \
		else \
			echo "Signing $$install_binary ad-hoc (set A365_CODESIGN_IDENTITY to use a stable cert)"; \
			codesign --force --sign - "$$install_binary" 2>/dev/null || true; \
		fi; \
	fi
endef

.PHONY: build build-cgo build-static install install-cgo install-static test clean lint fmt vet

## build: Build the a365 binary with OS-backed token cache support when available
build: build-cgo

## build-cgo: Build the a365 binary with default native auth behavior
build-cgo:
	$(GOENV) $(GO) build $(LDFLAGS) -o $(BINARY) .
	$(call codesign_binary,$(BINARY))

## build-static: Build a pure-Go binary without OS-backed token cache support
build-static:
	CGO_ENABLED=0 $(GOENV) $(GO) build $(LDFLAGS) -o $(BINARY) .
	$(call codesign_binary,$(BINARY))

## install: Install a365 with OS-backed token cache support when available
install: install-cgo

## install-cgo: Install a365 with default native auth behavior
install-cgo:
	$(call install_binary,)

## install-static: Install a pure-Go a365 binary without OS-backed token cache support
install-static:
	$(call install_binary,CGO_ENABLED=0)

## test: Run all tests
test:
	$(GO) test ./... -v

## test-short: Run tests without verbose output
test-short:
	$(GO) test ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)

## fmt: Format Go source files
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## lint: Run fmt and vet
lint: fmt vet

## help: Show this help
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
