.PHONY: build install uninstall clean

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
	@rm -f $(BINARY)
	@echo "✓ Clean complete"

# Show help
help:
	@echo "credctl Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build      - Build the binary"
	@echo "  make install    - Install credctl to \$$(go env GOPATH)/bin"
	@echo "  make uninstall  - Remove credctl from \$$(go env GOPATH)/bin"
	@echo "  make clean      - Clean build artifacts"
	@echo ""
	@echo "Installation directory can be changed with PREFIX:"
	@echo "  make install PREFIX=/usr/local"
	@echo "  make install PREFIX=~/.local"
	@echo ""
	@echo "Version can be set explicitly:"
	@echo "  make build VERSION=1.0.0"

