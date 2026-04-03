BINARY := a365
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/sozercan/a365cli/internal/version.Version=$(VERSION) -X github.com/sozercan/a365cli/internal/version.Commit=$(COMMIT)"

.PHONY: build install test clean lint fmt vet

## build: Build the a365 binary
build:
	go build $(LDFLAGS) -o $(BINARY) .

## install: Install a365 to $GOPATH/bin
install:
	go install $(LDFLAGS) .

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
