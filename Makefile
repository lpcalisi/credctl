.PHONY: build install uninstall clean

# Binary name
BINARY=credctl

# Installation directory (defaults to GOPATH/bin)
GOPATH?=$(shell go env GOPATH)
PREFIX?=$(GOPATH)
INSTALL_DIR=$(PREFIX)/bin

# Build binary
build:
	@echo "Building $(BINARY)..."
	@go build -o $(BINARY) .
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

