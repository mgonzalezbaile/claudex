.PHONY: all build clean install install-project uninstall help test run deps

# Configuration
SRC_DIR = src
CONFIG_DIR = $(HOME)/.config/claudex
BIN_DIR = $(HOME)/.local/bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Default target
all: build

# Build the executable
build:
	@echo "Building claudex $(VERSION)..."
	@cd $(SRC_DIR) && go build -ldflags "-X main.Version=$(VERSION)" -o ../claudex ./cmd/claudex
	@echo "✓ Built: claudex $(VERSION)"

# Build hooks binary
build-hooks:
	@echo "Building claudex-hooks..."
	@cd $(SRC_DIR) && go build -o ../bin/claudex-hooks ./cmd/claudex-hooks
	@echo "✓ Built: bin/claudex-hooks"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@cd $(SRC_DIR) && go mod tidy
	@echo "✓ Dependencies installed"

# Run tests
test:
	@echo "Running tests..."
	@cd $(SRC_DIR) && go test -v ./...
	@echo "✓ Tests complete"

# Install hooks (binary + proxies)
install-hooks: build-hooks
	@echo "Installing claudex-hooks..."
	@mkdir -p $(BIN_DIR) $(CONFIG_DIR)/hooks
	@install -m 755 bin/claudex-hooks $(BIN_DIR)/claudex-hooks
	@install -m 755 $(SRC_DIR)/scripts/proxies/*.sh $(CONFIG_DIR)/hooks/
	@echo "✓ Installed claudex-hooks to $(BIN_DIR)"
	@echo "✓ Installed hook proxies to $(CONFIG_DIR)/hooks"

# Configure recommended MCPs (opt-in)
install-mcp: build
	@echo "Configuring recommended MCPs..."
	@./claudex --setup-mcp

# Install claudex for current user
install: build build-hooks
	@echo "Installing claudex..."
	@mkdir -p $(CONFIG_DIR) $(BIN_DIR)
	@cp -r $(SRC_DIR)/profiles $(CONFIG_DIR)/
	@ln -sf $(CURDIR)/claudex $(BIN_DIR)/claudex
	@install -m 755 bin/claudex-hooks $(BIN_DIR)/claudex-hooks
	@mkdir -p $(CONFIG_DIR)/hooks
	@install -m 755 $(SRC_DIR)/scripts/proxies/*.sh $(CONFIG_DIR)/hooks/
	@echo "✓ Installed to $(CONFIG_DIR)"
	@echo "✓ Binary linked at $(BIN_DIR)/claudex"
	@echo "✓ Hooks installed to $(CONFIG_DIR)/hooks"
	@if ! echo "$$PATH" | grep -q "$(BIN_DIR)"; then \
		echo "⚠ Add to your shell config: export PATH=\"\$$HOME/.local/bin:\$$PATH\""; \
	fi

# Uninstall claudex
uninstall:
	@echo "Uninstalling claudex..."
	@rm -rf $(CONFIG_DIR)
	@rm -f $(BIN_DIR)/claudex
	@echo "✓ Uninstalled"

# Install to current project directory
install-project:
	@echo "Installing claudex to current project..."
	@mkdir -p .claude
	@if [ -d "$(CONFIG_DIR)/profiles" ]; then \
		cp -r $(CONFIG_DIR)/profiles .claude/; \
		echo "✓ Copied profiles from $(CONFIG_DIR)"; \
	elif [ -d "profiles" ]; then \
		cp -r profiles .claude/; \
		echo "✓ Copied profiles from local directory"; \
	else \
		echo "✗ No profiles directory found"; \
		exit 1; \
	fi
	@if [ -d "$(CONFIG_DIR)/hooks" ]; then \
		cp -r $(CONFIG_DIR)/hooks .claude/; \
		echo "✓ Copied hooks from $(CONFIG_DIR)"; \
	elif [ -d ".claude/hooks" ]; then \
		echo "✓ Hooks already exist in .claude/"; \
	else \
		echo "⚠ No hooks directory found"; \
	fi
	@echo "✓ Project installation complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f claudex
	@rm -rf bin/
	@echo "✓ Cleaned"

# Run the program
run: build
	@./claudex

# Show help
help:
	@echo "Available targets:"
	@echo "  make all             - Build the executable (default)"
	@echo "  make build           - Build claudex"
	@echo "  make build-hooks     - Build claudex-hooks binary"
	@echo "  make deps            - Install/update dependencies"
	@echo "  make test            - Run tests"
	@echo "  make install         - Install claudex and hooks to ~/.local/bin and ~/.config/claudex"
	@echo "  make install-hooks   - Install only hooks binary and proxies"
	@echo "  make install-mcp     - Configure recommended MCP servers (sequential-thinking, context7)"
	@echo "  make uninstall       - Remove claudex installation"
	@echo "  make install-project - Install profiles/hooks to current project .claude/"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make run             - Build and run"
	@echo "  make help            - Show this help"
