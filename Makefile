# tokenctl/Makefile

# tokenctl is a self-contained module. When this checkout sits inside a
# parent go.work that doesn't list it (e.g. the fastatic monorepo),
# `go install` refuses with a workspace-membership error. Force the
# module path regardless of where the directory lives on disk.
export GOWORK := off

BINARY_NAME=tokenctl
BUILD_DIR=bin
CMD_DIR=./cmd/tokenctl
PKG=./...

# Version info for builds
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
VERSION_PKG := github.com/dmoose/tokenctl/pkg/version
LDFLAGS := -X $(VERSION_PKG).Value=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).BuildTime=$(BUILD_TIME)

# Single source of truth: the exact version CI pins in the workflow file.
# Never hand-sync a copy here — that drift is how CI broke on 2026-08-12.
GOLANGCI_LINT_VERSION := $(shell sed -n 's/.*version: \(v[0-9][0-9.]*\).*/\1/p' .github/workflows/go.yml | head -1)

.PHONY: all build clean test test-unit test-integration coverage lint install dev-deps help

all: clean build test

## Build

build: ## Build the tokenctl binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

install: ## Install tokenctl to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install -ldflags "$(LDFLAGS)" $(CMD_DIR)

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
	go test -v ./cmd/tokenctl -run TestIntegration

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

lint: ## Lint at the exact version CI runs (read from the workflow file)
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --timeout=5m

fmt: ## Format code
	@echo "Formatting code..."
	go fmt $(PKG)

vet: ## Run go vet
	@echo "Running go vet..."
	go vet $(PKG)

## Development

dev-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

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

example-baseline: build ## Build baseline example (complete design system)
	@echo "Building baseline example..."
	./$(BUILD_DIR)/$(BINARY_NAME) build examples/baseline --format=css --output examples/baseline/dist
	@echo "Output: examples/baseline/dist/tokens.css"
	@echo "Demo:   open examples/baseline/demo.html"

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

example-computed: build ## Build computed example
	@echo "Building computed example..."
	./$(BUILD_DIR)/$(BINARY_NAME) build examples/computed --output dist/computed
	@echo "Output: dist/computed/tokens.css"

example-validation: build ## Build validation example
	@echo "Building validation example..."
	./$(BUILD_DIR)/$(BINARY_NAME) build examples/validation --output dist/validation
	@echo "Output: dist/validation/tokens.css"

example-daisyui: build ## Build DaisyUI 5 theme example
	@echo "Building DaisyUI example..."
	./$(BUILD_DIR)/$(BINARY_NAME) build examples/daisyui --output dist/daisyui
	@echo "Output: dist/daisyui/tokens.css"

examples: example-baseline example-basic example-themes example-components example-computed example-validation example-daisyui ## Build all examples

## CI/CD

ci: ## Exactly what GitHub CI runs — the pre-push gate (also wired as .git/hooks/pre-push)
	go build -v ./...
	go test -race ./...
	$(MAKE) lint

check: fmt vet ## Quick quality check (format + vet)

## Help

help: ## Show this help message
	@echo "tokenctl Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
