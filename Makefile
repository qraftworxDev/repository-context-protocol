.PHONY: build test install clean lint fmt vet tidy deps dev-setup integration-test coverage pre-commit help

# Binary names
BINARY_NAME=repocontext
LSP_BINARY_NAME=repocontext-lsp

# Build info
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build binaries
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/repocontext/main.go
	go build $(LDFLAGS) -o bin/$(LSP_BINARY_NAME) cmd/lsp/main.go

test: ## Run unit tests
	go test -v -race ./...

coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

integration-test: build ## Run integration tests
	./scripts/test-integration.sh

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt: ## Format code
	go fmt ./...
	@which goimports > /dev/null || (echo "goimports not found. Installing..." && go install golang.org/x/tools/cmd/goimports@latest)
	goimports -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy dependencies
	go mod tidy

deps: ## Download dependencies
	go mod download

tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/goreleaser/goreleaser@latest

dev-setup: build ## Setup development environment
	@echo "Setting up development environment..."
	@mkdir -p testdata/simple-go testdata/complex-go testdata/multi-lang
	cd testdata/simple-go && ../../bin/$(BINARY_NAME) init
	cd testdata/simple-go && ../../bin/$(BINARY_NAME) build

install: build ## Install binaries to system
	cp bin/$(BINARY_NAME) /usr/local/bin/
	cp bin/$(LSP_BINARY_NAME) /usr/local/bin/

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

pre-commit: fmt vet lint test ## Run pre-commit checks

setup: tools deps ## Install tools and dependencies
	@echo "Development environment setup complete!"

all: clean deps fmt vet lint test build ## Run full build pipeline 