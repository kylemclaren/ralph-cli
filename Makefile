.PHONY: build install clean test lint

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o ralph ./cmd/ralph

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/ralph

clean:
	rm -f ralph
	rm -rf dist/

test:
	go test -v ./...

lint:
	golangci-lint run

# Build for multiple platforms
dist: clean
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/ralph-darwin-arm64 ./cmd/ralph
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ralph-darwin-amd64 ./cmd/ralph
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ralph-linux-amd64 ./cmd/ralph
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/ralph-linux-arm64 ./cmd/ralph
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ralph-windows-amd64.exe ./cmd/ralph

# Run ralph init in current directory
init:
	./ralph init

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build ralph binary"
	@echo "  install  - Install ralph to GOPATH/bin"
	@echo "  clean    - Remove built binaries"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linter"
	@echo "  dist     - Build for all platforms"
	@echo "  init     - Run ralph init in current directory"
