# tokctl/Makefile

BINARY_NAME=tokctl
BUILD_DIR=bin
CMD_DIR=./cmd/tokctl
PKG=./...

.PHONY: all build clean test test-unit test-integration coverage lint install dev-deps help

all: clean build test

## Build

build: ## Build the tokctl binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

install: build ## Install tokctl to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(CMD_DIR)

## Clean

clean: ## Remove build artifacts and test outputs
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf dist
	@rm -rf .build
	@rm -rf test-*-out
	@rm -f coverage.out
	@rm -f $(BINARY_NAME)

## Test

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v ./pkg/...

test-integration: build ## Run integration tests
	@echo "Running integration tests..."
	go test -v ./cmd/tokctl -run TestIntegration

coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report:"
	@echo "  go tool cover -html=coverage.out"

coverage-html: coverage ## Generate and open HTML coverage report
	go tool cover -html=coverage.out

## Quality

lint: ## Run linters
	@echo "Linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, falling back to go vet"; \
		go vet $(PKG); \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	go fmt $(PKG)

vet: ## Run go vet
	@echo "Running go vet..."
	go vet $(PKG)

## Development

dev-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

## Examples

demo: build ## Run full demo workflow (init -> validate -> build)
	@echo "=== Demo: Full Workflow ==="
	@mkdir -p .demo
	@echo "\n→ Initializing token system..."
	./$(BUILD_DIR)/$(BINARY_NAME) init .demo
	@echo "\n→ Validating tokens..."
	./$(BUILD_DIR)/$(BINARY_NAME) validate .demo
	@echo "\n→ Building CSS output..."
	./$(BUILD_DIR)/$(BINARY_NAME) build .demo --output .demo/dist
	@echo "\n→ Generated output:"
	@head -n 20 .demo/dist/tokens.css
	@echo "\n✅ Demo complete! Output in .demo/dist/"
	@rm -rf .demo

example-basic: build ## Build basic example
	@echo "Building basic example..."
	./$(BUILD_DIR)/$(BINARY_NAME) build examples/basic --output dist/basic
	@echo "Output: dist/basic/tokens.css"

example-themes: build ## Build themes example
	@echo "Building themes example..."
	./$(BUILD_DIR)/$(BINARY_NAME) build examples/themes --output dist/themes
	@echo "Output: dist/themes/tokens.css"

example-components: build ## Build components example
	@echo "Building components example..."
	./$(BUILD_DIR)/$(BINARY_NAME) build examples/components --output dist/components
	@echo "Output: dist/components/tokens.css"

examples: example-basic example-themes example-components ## Build all examples

## CI/CD

ci: lint test coverage ## Run CI pipeline (lint, test, coverage)

## Help

help: ## Show this help message
	@echo "tokctl Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
