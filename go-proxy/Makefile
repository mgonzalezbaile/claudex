# Makefile for claudex-proxy Go implementation

BINARY_NAME=claudex-proxy
GO=go
GOFLAGS=-ldflags="-s -w"

# Detect OS and architecture
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-arm64 .

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-arm64 .

.PHONY: build-windows
build-windows:
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Install dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Verifying dependencies..."
	$(GO) mod verify

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	@echo "Clean complete"

# Install to system (requires appropriate permissions)
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	cp $(BINARY_NAME) /usr/local/bin/
	chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall from system
.PHONY: uninstall
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Display help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Build the binary (default)"
	@echo "  build        - Build the binary for current platform"
	@echo "  build-all    - Build for all supported platforms"
	@echo "  build-linux  - Build for Linux (amd64 and arm64)"
	@echo "  build-darwin - Build for macOS (amd64 and arm64)"
	@echo "  build-windows- Build for Windows (amd64)"
	@echo "  deps         - Download and verify dependencies"
	@echo "  test         - Run tests"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install binary to /usr/local/bin"
	@echo "  uninstall    - Remove binary from /usr/local/bin"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  help         - Display this help message"

