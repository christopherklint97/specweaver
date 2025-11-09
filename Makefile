# SpecWeaver Makefile
# Comprehensive build and development automation

.PHONY: help build install clean test test-coverage test-race test-verbose \
        fmt vet lint check \
        example-server example-library example-custom-router \
        generate generate-examples \
        deps update-deps \
        all

# Default target
.DEFAULT_GOAL := help

# Binary and package settings
BINARY_NAME := specweaver
MAIN_PATH := ./cmd/specweaver
PKG := github.com/christopherklint97/specweaver

# Build settings
GOFLAGS := -v
LDFLAGS := -w -s

# Coverage settings
COVERAGE_DIR := coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

##@ Help

help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

build: ## Build the specweaver binary
	@echo "Building $(BINARY_NAME)..."
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Build complete: ./$(BINARY_NAME)"

install: ## Install the specweaver binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(MAIN_PATH)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

clean: ## Clean build artifacts and generated files
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf generated/
	@rm -rf $(COVERAGE_DIR)
	@find . -name "*.test" -type f -delete
	@echo "✓ Clean complete"

all: clean fmt vet test build ## Run clean, fmt, vet, test, and build

##@ Testing

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...
	@echo "✓ All tests passed"

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@go tool cover -func=$(COVERAGE_FILE)
	@echo "✓ Coverage report: $(COVERAGE_HTML)"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -v -race ./...
	@echo "✓ Race tests passed"

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	@go test -v -count=1 ./...

test-bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

##@ Code Quality

fmt: ## Format all Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet passed"

lint: ## Run golangci-lint (if installed)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "✓ Lint passed"; \
	else \
		echo "⚠ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

check: fmt vet test ## Run fmt, vet, and test (pre-commit checks)
	@echo "✓ All checks passed"

##@ Code Generation

generate: build ## Generate code from petstore example
	@echo "Generating code from examples/petstore.yaml..."
	@./$(BINARY_NAME) -spec examples/petstore.yaml -output generated -package api
	@echo "✓ Code generated in ./generated"

generate-examples: build ## Regenerate code for all examples
	@echo "Regenerating code for examples/server..."
	@./$(BINARY_NAME) -spec examples/petstore.yaml -output examples/server/api -package api
	@echo "✓ Examples regenerated"

##@ Examples

example-server: build ## Run the example server
	@echo "Starting example server..."
	@cd examples/server && go run main.go

example-library: ## Run the library usage example
	@echo "Running library example..."
	@cd examples/library && go run main.go

example-custom-router: ## Run the custom router example
	@echo "Starting custom router example..."
	@cd examples/custom-router && go run .

##@ Dependencies

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@echo "✓ Dependencies downloaded"

update-deps: ## Update dependencies to latest versions
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✓ Dependencies updated"

tidy: ## Run go mod tidy
	@echo "Tidying go.mod..."
	@go mod tidy
	@echo "✓ go.mod tidied"

##@ Development

dev: clean fmt vet ## Quick development cycle (clean, fmt, vet, build)
	@$(MAKE) build

watch: ## Watch for changes and rebuild (requires entr)
	@if command -v find >/dev/null 2>&1 && command -v entr >/dev/null 2>&1; then \
		echo "Watching for changes..."; \
		find . -name "*.go" | entr -c make build; \
	else \
		echo "⚠ This target requires 'find' and 'entr'. Install entr from http://eradman.com/entrproject/"; \
	fi

version: ## Show Go and module versions
	@echo "Go version: $(shell go version)"
	@echo "Module: $(PKG)"
	@echo "Binary: $(BINARY_NAME)"

##@ Docker (Optional)

docker-build: ## Build Docker image (if Dockerfile exists)
	@if [ -f Dockerfile ]; then \
		echo "Building Docker image..."; \
		docker build -t $(BINARY_NAME):latest .; \
		echo "✓ Docker image built: $(BINARY_NAME):latest"; \
	else \
		echo "⚠ Dockerfile not found"; \
	fi

##@ Utilities

tree: ## Show project structure
	@if command -v tree >/dev/null 2>&1; then \
		tree -I 'generated|vendor|.git|coverage' -L 3; \
	else \
		find . -type d -not -path '*/.*' -not -path '*/generated*' -not -path '*/vendor*' | sed 's|[^/]*/|  |g'; \
	fi

size: build ## Show binary size
	@echo "Binary size:"
	@ls -lh $(BINARY_NAME) | awk '{print $$5 "\t" $$9}'

todo: ## Show TODO and FIXME comments in code
	@echo "TODOs and FIXMEs:"
	@grep -rn "TODO\|FIXME" --include="*.go" . || echo "None found"
