.PHONY: all build clean install help test

# Default target
all: build

# Build the executable
build:
	@echo "Building claudex-session..."
	@go build -o claudex-session main.go
	@echo "✓ Built: claudex-session"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@echo "✓ Dependencies installed"

# Install the binary
install: build
	@echo "Installing claudex-session..."
	@chmod +x claudex-session
	@echo "✓ Installed: claudex-session"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f claudex-session
	@echo "✓ Cleaned"

# Run the program
run: build
	@./claudex-session

# Show help
help:
	@echo "Available targets:"
	@echo "  make all      - Build the executable (default)"
	@echo "  make build    - Build claudex-session"
	@echo "  make deps     - Install/update dependencies"
	@echo "  make install  - Build and mark as executable"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make run      - Build and run"
	@echo "  make help     - Show this help"
