# PocketConcierge DNS Server Makefile
# Build and deployment automation

# Project information
PROJECT_NAME := pocketconcierge
BINARY_NAME := pocketconcierge
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
BUILD_FLAGS := -v $(LDFLAGS)

# Directories
BUILD_DIR := ./build
DIST_DIR := ./dist
CMD_DIR := ./cmd
INTERNAL_DIR := ./internal

# Operating System Detection
ifeq ($(OS),Windows_NT)
		BINARY_EXT := .exe
		SHELL_EXT := .bat
		RM := del /Q
		MKDIR := mkdir
		COPY := copy
else
		BINARY_EXT :=
		SHELL_EXT := .sh
		RM := rm -f
		MKDIR := mkdir -p
		COPY := cp
endif

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
		@echo "🏨 PocketConcierge DNS Server - Build Commands"
		@echo "=============================================="
		@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "	\033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## Development Commands

.PHONY: run
run: ## Run the DNS server with default config
		@echo "🚀 Starting PocketConcierge..."
		$(GOCMD) run $(CMD_DIR)/$(PROJECT_NAME)/main.go

.PHONY: run-config
run-config: ## Run with custom config file (make run-config CONFIG=myconfig.yaml)
		@echo "🚀 Starting PocketConcierge with $(CONFIG)..."
		$(GOCMD) run $(CMD_DIR)/$(PROJECT_NAME)/main.go $(CONFIG)

.PHONY: dev
dev: clean format lint test build ## Full development cycle: clean, format, lint, test, build

## Build Commands

.PHONY: build
build: deps ## Build binary for current platform
		@echo "🔨 Building $(BINARY_NAME) $(VERSION)..."
		@$(MKDIR) $(BUILD_DIR) 2>/dev/null || true
		$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(CMD_DIR)/$(PROJECT_NAME)/main.go

.PHONY: build-all
build-all: ## Build binaries for all platforms
		@echo "🔨 Building for all platforms..."
		@$(MKDIR) $(DIST_DIR) 2>/dev/null || true
		@echo "Building for Windows (amd64)..."
		GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/$(PROJECT_NAME)/main.go
		@echo "Building for Linux (amd64)..."
		GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/$(PROJECT_NAME)/main.go
		@echo "Building for macOS (amd64)..."
		GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/$(PROJECT_NAME)/main.go
		@echo "Building for macOS (arm64)..."
		GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/$(PROJECT_NAME)/main.go
		@echo "Building for Linux (arm64)..."
		GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)/$(PROJECT_NAME)/main.go
		@echo "✅ All builds complete!"

## Test Commands

.PHONY: test
test: ## Run all tests
		@echo "🧪 Running tests..."
		$(GOTEST) -v ./...

.PHONY: test-race
test-race: ## Run tests with race detection
		@echo "🏁 Running tests with race detection..."
		$(GOTEST) -race -v ./...

.PHONY: benchmark
benchmark: ## Run DNS performance benchmark
		@echo "📊 Running DNS benchmarks..."
		$(GOCMD) run cmd/benchmark/main.go 127.0.0.1:8053 500 20 mixed

## Quality Commands

.PHONY: format
format: ## Format Go code
		@echo "🎨 Formatting code..."
		$(GOFMT) ./...

.PHONY: lint
lint: ## Run linting tools
		@echo "🔍 Running linter..."
		@which golangci-lint > /dev/null || echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
		golangci-lint run || echo "Linter not found, install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

.PHONY: vet
vet: ## Run go vet
		@echo "🔍 Running go vet..."
		$(GOCMD) vet ./...

## Dependency Commands

.PHONY: deps
deps: ## Download dependencies
		@echo "📦 Downloading dependencies..."
		$(GOMOD) tidy
		$(GOMOD) download

.PHONY: deps-update
deps-update: ## Update dependencies
		@echo "📦 Updating dependencies..."
		$(GOMOD) get -u ./...
		$(GOMOD) tidy

## Utility Commands

.PHONY: clean
clean: ## Clean build artifacts
		@echo "🧹 Cleaning build artifacts..."
		$(GOCLEAN)
		@$(RM) $(BUILD_DIR)/* 2>/dev/null || true
		@$(RM) $(DIST_DIR)/* 2>/dev/null || true

.PHONY: version
version: ## Show version information
		@echo "PocketConcierge DNS Server"
		@echo "Version: $(VERSION)"
		@echo "Build Time: $(BUILD_TIME)"
		@echo "Git Commit: $(GIT_COMMIT)"

.PHONY: config-example
config-example: ## Create example configuration file
		@echo "📋 Creating example config..."
		@$(COPY) config.yaml config-example.yaml

.PHONY: release
release: clean format vet test build ## Create a release build
		@echo "🎉 Release $(VERSION) ready!"
