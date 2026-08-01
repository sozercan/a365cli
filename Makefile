GO ?= go
BINARY ?= a365
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
A365_TOKEN_CACHE_NAME ?= a365-dev
LDFLAGS := -ldflags "-X github.com/sozercan/a365cli/internal/version.Version=$(VERSION) -X github.com/sozercan/a365cli/internal/version.Commit=$(COMMIT) -X github.com/sozercan/a365cli/internal/auth.persistentTokenCacheName=$(A365_TOKEN_CACHE_NAME)"
TARGET_GOOS := $(or $(GOOS),$(shell $(GO) env GOOS))
TARGET_GOARCH := $(or $(GOARCH),$(shell $(GO) env GOARCH))
TARGET_GOEXE := $(shell GOOS=$(TARGET_GOOS) GOARCH=$(TARGET_GOARCH) $(GO) env GOEXE)
HOST_GOOS := $(shell $(GO) env GOHOSTOS)
HOST_GOARCH := $(shell $(GO) env GOHOSTARCH)
GOENV := GOOS=$(TARGET_GOOS) GOARCH=$(TARGET_GOARCH)
A365_CODESIGN_IDENTITY ?=
A365_CODESIGN_IDENTIFIER ?= com.github.sozercan.a365
A365_REQUIRE_CERT_SIGNING ?= 1
A365_ALLOW_ADHOC_SIGNING ?= 0
GORELEASER ?= goreleaser

# Canonical macOS builds require an explicitly selected certificate-backed
# identity. Ad-hoc signing is available only through an explicitly named target.
define codesign_binary
	@set -e; \
	if [ "$(TARGET_GOOS)" != "darwin" ]; then \
		exit 0; \
	fi; \
	if [ "$(HOST_GOOS)" != "darwin" ]; then \
		if [ "$(A365_REQUIRE_CERT_SIGNING)" = "1" ]; then \
			echo "ERROR: canonical Darwin builds must be produced and certificate-signed on macOS." >&2; \
			rm -f "$(1)"; \
			exit 1; \
		fi; \
		echo "WARNING: leaving cross-built Darwin artifact unsigned; it is not a canonical a365 binary." >&2; \
		exit 0; \
	fi; \
	if ! command -v codesign >/dev/null 2>&1; then \
		echo "ERROR: codesign is required for Darwin builds." >&2; \
		rm -f "$(1)"; \
		exit 1; \
	fi; \
	identity="$(A365_CODESIGN_IDENTITY)"; \
	if [ -n "$$identity" ] && [ "$$identity" != "-" ]; then \
		if command -v security >/dev/null 2>&1 && \
			! security find-identity -v -p codesigning 2>/dev/null | grep -Fqi -- "$$identity"; then \
			echo "ERROR: A365_CODESIGN_IDENTITY does not match a valid local codesigning identity." >&2; \
			rm -f "$(1)"; \
			exit 1; \
		fi; \
		echo "Signing $(1) with the explicitly configured certificate-backed identity"; \
		codesign --force --options runtime --identifier "$(A365_CODESIGN_IDENTIFIER)" --sign "$$identity" "$(1)"; \
		codesign --verify --strict --verbose=2 "$(1)"; \
		details="$$(codesign -dv --verbose=4 "$(1)" 2>&1)"; \
		if ! printf '%s\n' "$$details" | grep -q '^Authority=' || \
			! printf '%s\n' "$$details" | grep -q '^Identifier=$(A365_CODESIGN_IDENTIFIER)$$'; then \
			echo "ERROR: certificate-backed signature verification failed." >&2; \
			rm -f "$(1)"; \
			exit 1; \
		fi; \
	elif [ "$(A365_ALLOW_ADHOC_SIGNING)" = "1" ]; then \
		echo "WARNING: signing $(1) ad-hoc; use only as a disposable development binary." >&2; \
		codesign --force --identifier "$(A365_CODESIGN_IDENTIFIER)" --sign - "$(1)"; \
		codesign --verify --strict --verbose=2 "$(1)"; \
	else \
		echo "ERROR: set A365_CODESIGN_IDENTITY to a certificate SHA-1 or exact identity name." >&2; \
		echo "       For a disposable build, run 'make build-adhoc' instead." >&2; \
		rm -f "$(1)"; \
		exit 1; \
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
	staged_binary="$$install_binary.tmp.$$$$"; \
	mkdir -p "$$bin_dir"; \
	trap 'rm -f "$$staged_binary"' 0 1 2 15; \
	$(1) $(GOENV) $(GO) build $(LDFLAGS) -o "$$staged_binary" .; \
	if [ "$(TARGET_GOOS)" = "darwin" ]; then \
		if [ "$(HOST_GOOS)" != "darwin" ]; then \
			if [ "$(A365_REQUIRE_CERT_SIGNING)" = "1" ]; then \
				echo "ERROR: canonical Darwin installs must be produced and certificate-signed on macOS." >&2; \
				exit 1; \
			fi; \
			echo "WARNING: installing an unsigned, non-canonical cross-built Darwin artifact." >&2; \
		else \
			identity="$(A365_CODESIGN_IDENTITY)"; \
			if [ -n "$$identity" ] && [ "$$identity" != "-" ]; then \
				if command -v security >/dev/null 2>&1 && \
					! security find-identity -v -p codesigning 2>/dev/null | grep -Fqi -- "$$identity"; then \
					echo "ERROR: A365_CODESIGN_IDENTITY does not match a valid local codesigning identity." >&2; \
					exit 1; \
				fi; \
				echo "Signing staged a365 binary with the explicitly configured certificate-backed identity"; \
				codesign --force --options runtime --identifier "$(A365_CODESIGN_IDENTIFIER)" --sign "$$identity" "$$staged_binary"; \
				codesign --verify --strict --verbose=2 "$$staged_binary"; \
				details="$$(codesign -dv --verbose=4 "$$staged_binary" 2>&1)"; \
				if ! printf '%s\n' "$$details" | grep -q '^Authority=' || \
					! printf '%s\n' "$$details" | grep -q '^Identifier=$(A365_CODESIGN_IDENTIFIER)$$'; then \
					echo "ERROR: certificate-backed signature verification failed." >&2; \
					exit 1; \
				fi; \
			elif [ "$(A365_ALLOW_ADHOC_SIGNING)" = "1" ]; then \
				echo "WARNING: signing staged a365 binary ad-hoc; use only as a disposable development binary." >&2; \
				codesign --force --identifier "$(A365_CODESIGN_IDENTIFIER)" --sign - "$$staged_binary"; \
				codesign --verify --strict --verbose=2 "$$staged_binary"; \
			else \
				echo "ERROR: set A365_CODESIGN_IDENTITY before installing the canonical macOS binary." >&2; \
				exit 1; \
			fi; \
		fi; \
	fi; \
	mv -f "$$staged_binary" "$$install_binary"; \
	trap - 0 1 2 15
endef

.PHONY: build build-cgo build-adhoc build-static install install-cgo install-adhoc install-static release-check test test-short clean lint fmt vet help

## build: Build the canonical native binary (macOS requires explicit certificate signing)
build: build-cgo

## build-cgo: Build with CGO-backed OS token caching on supported native platforms
build-cgo:
	CGO_ENABLED=1 $(GOENV) $(GO) build $(LDFLAGS) -o $(BINARY).tmp .
	$(call codesign_binary,$(BINARY).tmp)
	mv -f $(BINARY).tmp $(BINARY)

## build-adhoc: Build a disposable CGO-enabled macOS binary with explicit ad-hoc signing
build-adhoc: A365_REQUIRE_CERT_SIGNING=0
build-adhoc: A365_ALLOW_ADHOC_SIGNING=1
build-adhoc: build-cgo

## build-static: Build a portable pure-Go binary without persistent OS token caching
build-static: A365_REQUIRE_CERT_SIGNING=0
build-static: A365_ALLOW_ADHOC_SIGNING=1
build-static:
	CGO_ENABLED=0 $(GOENV) $(GO) build $(LDFLAGS) -o $(BINARY).tmp .
	$(call codesign_binary,$(BINARY).tmp)
	mv -f $(BINARY).tmp $(BINARY)

## install: Install the canonical native binary (macOS requires explicit certificate signing)
install: install-cgo

## install-cgo: Install with CGO-backed OS token caching on supported native platforms
install-cgo:
	$(call install_binary,CGO_ENABLED=1)

## install-adhoc: Install a disposable CGO-enabled macOS binary with explicit ad-hoc signing
install-adhoc: A365_REQUIRE_CERT_SIGNING=0
install-adhoc: A365_ALLOW_ADHOC_SIGNING=1
install-adhoc:
	$(call install_binary,CGO_ENABLED=1)

## install-static: Install a portable pure-Go binary without persistent OS token caching
install-static: A365_REQUIRE_CERT_SIGNING=0
install-static: A365_ALLOW_ADHOC_SIGNING=1
install-static:
	$(call install_binary,CGO_ENABLED=0)

## release-check: Validate .goreleaser.yml with the installed GoReleaser
release-check:
	@command -v $(GORELEASER) >/dev/null 2>&1 || { echo "ERROR: goreleaser is required for release-check." >&2; exit 1; }
	$(GORELEASER) check

## test: Run all tests
test:
	A365_DISABLE_PERSISTENT_TOKEN_CACHE=1 $(GO) test ./... -v

## test-short: Run tests without verbose output
test-short:
	A365_DISABLE_PERSISTENT_TOKEN_CACHE=1 $(GO) test ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) $(BINARY).tmp

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
