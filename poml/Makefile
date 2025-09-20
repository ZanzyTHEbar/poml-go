# Makefile for POML Go module with linting support

.PHONY: help lint lint-clean build test clean dev-setup

# Default target
help:
	@echo "POML Go Module Build Targets (run from go/ directory):"
	@echo "  lint       - Run linters with struct tag validation disabled"
	@echo "  lint-clean - Run linters with full validation (may show struct tag warnings)"
	@echo "  build      - Build the module"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  dev-setup  - Show development workflow guide"

# Lint without struct tag validation (recommended for CI)
lint:
	@echo "Running linters with participle tag compatibility..."
	cd sdk && go vet -tags=no_participle_tags ./...
	cd sdk && go fmt ./...
	go mod tidy

# Lint with full validation (shows participle struct tag warnings)
lint-clean:
	@echo "Running linters with full validation..."
	cd sdk && go vet ./...
	cd sdk && go fmt ./...
	go mod tidy

# Build the module
build:
	@echo "Building POML module..."
	cd sdk && go build ./...

# Run tests
test:
	@echo "Running tests..."
	cd sdk && go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	cd sdk && go clean ./...
	rm -f *.out *.test

# Development workflow reminder
dev-setup:
	@echo "Development Setup:"
	@echo "1. Use 'make lint' for normal development (no struct tag warnings)"
	@echo "2. Use 'make lint-clean' to see all warnings including participle tags"
	@echo "3. Participle struct tags (@@, @@*, @@?) are intentional for parsing"
	@echo "4. Build tags separate linting from runtime functionality"
	@echo "5. For IDE configuration, set build tags manually:"
	@echo "   export GOFLAGS=\"-tags=no_participle_tags\""


