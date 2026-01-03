BINARY_NAME=tokctl
BUILD_DIR=bin
CMD_DIR=./cmd/tokctl
PKG=./...

.PHONY: all build clean test coverage lint run-init run-build

all: clean build test

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf dist
	@rm -rf test-system
	@rm -f coverage.out

test:
	@echo "Running tests..."
	go test -v $(PKG)

coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -func=coverage.out

lint:
	@echo "Linting..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed"; \
		go vet $(PKG); \
	fi

# Integration / Demo Tasks

run-init: build
	@echo "Running init demo..."
	@rm -rf test-system
	./$(BUILD_DIR)/$(BINARY_NAME) init test-system

run-build: run-init
	@echo "Running build demo..."
	@rm -rf dist
	./$(BUILD_DIR)/$(BINARY_NAME) build test-system --output dist --format tailwind
	@echo "Output generated at dist/tokens.css"
	@head -n 10 dist/tokens.css
