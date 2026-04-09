BINARY := a365
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/sozercan/a365cli/internal/version.Version=$(VERSION) -X github.com/sozercan/a365cli/internal/version.Commit=$(COMMIT)"

.PHONY: build build-cgo build-static install install-cgo install-static test clean lint fmt vet

## build: Build the a365 binary with OS-backed token cache support when available
build: build-cgo

## build-cgo: Build the a365 binary with default native auth behavior
build-cgo:
	go build $(LDFLAGS) -o $(BINARY) .
ifeq ($(shell uname),Darwin)
	@codesign --sign - --force $(BINARY) 2>/dev/null || true
endif

## build-static: Build a pure-Go binary without OS-backed token cache support
build-static:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) .
ifeq ($(shell uname),Darwin)
	@codesign --sign - --force $(BINARY) 2>/dev/null || true
endif

## install: Install a365 with OS-backed token cache support when available
install: install-cgo

## install-cgo: Install a365 with default native auth behavior
install-cgo:
	go install $(LDFLAGS) .

## install-static: Install a pure-Go a365 binary without OS-backed token cache support
install-static:
	CGO_ENABLED=0 go install $(LDFLAGS) .

## test: Run all tests
test:
	go test ./... -v

## test-short: Run tests without verbose output
test-short:
	go test ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)

## fmt: Format Go source files
fmt:
	go fmt ./...

## vet: Run go vet
vet:
	go vet ./...

## lint: Run fmt and vet
lint: fmt vet

## help: Show this help
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
