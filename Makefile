.PHONY: build install uninstall clean build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-all

# Binary name
BINARY=credctl

# Installation directory (defaults to GOPATH/bin)
GOPATH?=$(shell go env GOPATH)
PREFIX?=$(GOPATH)
INSTALL_DIR=$(PREFIX)/bin

# Package for version variables
PKG=sigs.k8s.io/release-utils/version

# Version information
GIT_VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_HASH?=$(shell git rev-parse HEAD 2>/dev/null || echo "none")
GIT_TREESTATE?=$(shell test -z "$$(git status --porcelain 2>/dev/null)" && echo "clean" || echo "dirty")
BUILD_DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags
LDFLAGS=-w -s -buildid= \
	-X $(PKG).gitVersion=$(GIT_VERSION) \
	-X $(PKG).gitCommit=$(GIT_HASH) \
	-X $(PKG).gitTreeState=$(GIT_TREESTATE) \
	-X $(PKG).buildDate=$(BUILD_DATE)

# Build binary
build:
	@echo "Building $(BINARY) $(GIT_VERSION)..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY) .
	@echo "Build complete: ./$(BINARY)"

# Install binary to PATH
install: build
	@echo "Installing $(BINARY) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@chmod +x $(INSTALL_DIR)/$(BINARY)
	@echo "✓ $(BINARY) installed successfully!"
	@echo "Run 'credctl --version' to verify"

# Uninstall binary from PATH
uninstall:
	@echo "Uninstalling $(BINARY)..."
	@rm -f $(INSTALL_DIR)/$(BINARY)
	@echo "✓ $(BINARY) uninstalled"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY) $(BINARY)_*
	@echo "✓ Clean complete"

# Cross-compilation targets
build-linux-amd64:
	@echo "Building $(BINARY) for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)_linux_amd64 .
	@echo "✓ Build complete: ./$(BINARY)_linux_amd64"

build-linux-arm64:
	@echo "Building $(BINARY) for linux/arm64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)_linux_arm64 .
	@echo "✓ Build complete: ./$(BINARY)_linux_arm64"

build-darwin-amd64:
	@echo "Building $(BINARY) for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)_darwin_amd64 .
	@echo "✓ Build complete: ./$(BINARY)_darwin_amd64"

build-darwin-arm64:
	@echo "Building $(BINARY) for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)_darwin_arm64 .
	@echo "✓ Build complete: ./$(BINARY)_darwin_arm64"

# Build for all common platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64
	@echo ""
	@echo "✓ All builds complete"
	@ls -lh $(BINARY)_*

# Show help
help:
	@echo "credctl Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build              - Build for current platform"
	@echo "  make install            - Install credctl to \$$(go env GOPATH)/bin"
	@echo "  make uninstall          - Remove credctl from \$$(go env GOPATH)/bin"
	@echo "  make clean              - Clean build artifacts"
	@echo ""
	@echo "Cross-compilation:"
	@echo "  make build-linux-amd64  - Build for Linux x86_64"
	@echo "  make build-linux-arm64  - Build for Linux ARM64"
	@echo "  make build-darwin-amd64 - Build for macOS Intel"
	@echo "  make build-darwin-arm64 - Build for macOS Apple Silicon"
	@echo "  make build-all          - Build for all platforms"
	@echo ""
	@echo "Installation directory can be changed with PREFIX:"
	@echo "  make install PREFIX=/usr/local"
	@echo "  make install PREFIX=~/.local"
	@echo ""
	@echo "Version can be set explicitly:"
	@echo "  make build VERSION=1.0.0"

