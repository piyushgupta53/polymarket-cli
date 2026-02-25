BINARY_NAME := polymarket
MODULE := github.com/piyushgupta/polymarket-cli

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

GO := go
GOFLAGS := -trimpath

.PHONY: build install test lint vet clean release help

## build: Build the binary
build:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY_NAME) .

## install: Install to $GOPATH/bin
install:
	$(GO) install $(GOFLAGS) -ldflags '$(LDFLAGS)' .

## test: Run all tests
test:
	$(GO) test -race -count=1 ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ dist/

## release: Cross-compile for all platforms
release: clean
	@mkdir -p dist
	GOOS=darwin  GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o dist/$(BINARY_NAME)_darwin_amd64 .
	GOOS=darwin  GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o dist/$(BINARY_NAME)_darwin_arm64 .
	GOOS=linux   GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o dist/$(BINARY_NAME)_linux_amd64 .
	GOOS=linux   GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o dist/$(BINARY_NAME)_linux_arm64 .
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o dist/$(BINARY_NAME)_windows_amd64.exe .

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
