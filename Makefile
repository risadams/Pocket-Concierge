# PocketConcierge DNS Server Makefile
# Build and deployment automation

# Project information
PROJECT_NAME := pocketconcierge
BINARY_NAME := pocketconcierge
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker parameters
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := $(VERSION)
DOCKER_REGISTRY := # Set this for your registry, e.g., docker.io/username
DOCKER_FULL_IMAGE := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)$(DOCKER_IMAGE):$(DOCKER_TAG)

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

## Docker Commands

.PHONY: docker-build
docker-build: ## Build Docker image
		@echo "🐳 Building Docker image $(DOCKER_FULL_IMAGE)..."
		docker build \
			--build-arg VERSION=$(VERSION) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			-t $(DOCKER_FULL_IMAGE) \
			-t $(DOCKER_IMAGE):latest \
			.

.PHONY: docker-build-no-cache
docker-build-no-cache: ## Build Docker image without cache
		@echo "🐳 Building Docker image $(DOCKER_FULL_IMAGE) (no cache)..."
		docker build --no-cache \
			--build-arg VERSION=$(VERSION) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			-t $(DOCKER_FULL_IMAGE) \
			-t $(DOCKER_IMAGE):latest \
			.

.PHONY: docker-run
docker-run: ## Run Docker container with default config
		@echo "🚀 Running $(DOCKER_IMAGE):latest..."
		docker run --rm -it \
			-p 8053:8053/udp \
			-p 8053:8053/tcp \
			--name $(PROJECT_NAME) \
			$(DOCKER_IMAGE):latest

.PHONY: docker-run-daemon
docker-run-daemon: ## Run Docker container as daemon
		@echo "🚀 Starting $(DOCKER_IMAGE):latest as daemon..."
		docker run -d \
			-p 8053:8053/udp \
			-p 8053:8053/tcp \
			--name $(PROJECT_NAME) \
			--restart unless-stopped \
			$(DOCKER_IMAGE):latest

.PHONY: docker-run-custom
docker-run-custom: ## Run Docker container with custom config (CONFIG=path/to/config.yaml)
		@echo "🚀 Running $(DOCKER_IMAGE):latest with custom config..."
		docker run --rm -it \
			-p 8053:8053/udp \
			-p 8053:8053/tcp \
			-v "$(shell pwd)/$(CONFIG):/app/config.yaml:ro" \
			--name $(PROJECT_NAME) \
			$(DOCKER_IMAGE):latest

.PHONY: docker-stop
docker-stop: ## Stop Docker container
		@echo "🛑 Stopping $(PROJECT_NAME) container..."
		docker stop $(PROJECT_NAME) || true

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
		@echo "📋 Showing logs for $(PROJECT_NAME)..."
		docker logs -f $(PROJECT_NAME)

.PHONY: docker-shell
docker-shell: ## Open shell in Docker container
		@echo "🐚 Opening shell in $(DOCKER_IMAGE):latest..."
		docker run --rm -it \
			--entrypoint /bin/sh \
			$(DOCKER_IMAGE):latest

.PHONY: docker-push
docker-push: ## Push Docker image to registry
		@echo "📤 Pushing $(DOCKER_FULL_IMAGE) to registry..."
		@if [ -z "$(DOCKER_REGISTRY)" ]; then \
			echo "❌ DOCKER_REGISTRY not set. Use: make docker-push DOCKER_REGISTRY=your-registry.com/username"; \
			exit 1; \
		fi
		docker push $(DOCKER_FULL_IMAGE)
		docker push $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)$(DOCKER_IMAGE):latest

.PHONY: docker-clean
docker-clean: ## Remove Docker images and containers
		@echo "🧹 Cleaning Docker artifacts..."
		docker rm -f $(PROJECT_NAME) 2>/dev/null || true
		docker rmi $(DOCKER_IMAGE):latest $(DOCKER_FULL_IMAGE) 2>/dev/null || true

.PHONY: docker-inspect
docker-inspect: ## Inspect Docker image
		@echo "🔍 Inspecting $(DOCKER_IMAGE):latest..."
		docker inspect $(DOCKER_IMAGE):latest

.PHONY: docker-all
docker-all: docker-build docker-run ## Build and run Docker container

.PHONY: docker-buildx-setup
docker-buildx-setup: ## Set up Docker buildx for multi-architecture builds
		@echo "🔧 Setting up Docker buildx..."
		docker buildx create --name $(PROJECT_NAME)-builder --use || true
		docker buildx inspect --bootstrap

.PHONY: docker-buildx-multiarch
docker-buildx-multiarch: ## Build multi-architecture Docker images
		@echo "🐳 Building multi-architecture Docker images..."
		docker buildx build \
			--platform linux/amd64,linux/arm64,linux/arm/v7 \
			--build-arg VERSION=$(VERSION) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			-f Dockerfile.multiarch \
			-t $(DOCKER_FULL_IMAGE) \
			-t $(DOCKER_IMAGE):latest \
			--push \
			.

.PHONY: docker-buildx-local
docker-buildx-local: ## Build multi-architecture images locally (no push)
		@echo "🐳 Building multi-architecture Docker images locally..."
		docker buildx build \
			--platform linux/amd64,linux/arm64,linux/arm/v7 \
			--build-arg VERSION=$(VERSION) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			-f Dockerfile.multiarch \
			-t $(DOCKER_FULL_IMAGE) \
			-t $(DOCKER_IMAGE):latest \
			--load \
			.

## Docker Compose Commands

.PHONY: compose-up
compose-up: ## Start services with Docker Compose
		@echo "🚀 Starting services with Docker Compose..."
		VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) docker-compose up -d

.PHONY: compose-up-build
compose-up-build: ## Build and start services with Docker Compose
		@echo "🔨 Building and starting services with Docker Compose..."
		VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) docker-compose up -d --build

.PHONY: compose-down
compose-down: ## Stop and remove Docker Compose services
		@echo "🛑 Stopping Docker Compose services..."
		docker-compose down

.PHONY: compose-logs
compose-logs: ## Show Docker Compose logs
		@echo "📋 Showing Docker Compose logs..."
		docker-compose logs -f

.PHONY: compose-restart
compose-restart: ## Restart Docker Compose services
		@echo "🔄 Restarting Docker Compose services..."
		docker-compose restart

.PHONY: compose-clean
compose-clean: ## Stop and remove Docker Compose services and volumes
		@echo "🧹 Cleaning Docker Compose resources..."
		docker-compose down -v --remove-orphans

.PHONY: compose-dev
compose-dev: ## Start development environment with Docker Compose
		@echo "🚀 Starting development environment..."
		VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) docker-compose -f docker-compose.dev.yml up -d --build

.PHONY: compose-dev-logs
compose-dev-logs: ## Show development environment logs
		@echo "📋 Showing development logs..."
		docker-compose -f docker-compose.dev.yml logs -f

.PHONY: compose-dev-down
compose-dev-down: ## Stop development environment
		@echo "🛑 Stopping development environment..."
		docker-compose -f docker-compose.dev.yml down

## Test Commands

.PHONY: test
test: ## Run all tests
		@echo "🧪 Running tests..."
		$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run tests with short flag (skip integration tests)
		@echo "🧪 Running short tests..."
		$(GOTEST) -short -v ./...

.PHONY: test-race
test-race: ## Run tests with race detection
		@echo "🏁 Running tests with race detection..."
		$(GOTEST) -race -v ./...

.PHONY: test-integration
test-integration: ## Run integration tests only
		@echo "🔗 Running integration tests..."
		$(GOTEST) -v ./test/...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
		@echo "📊 Running tests with coverage..."
		$(GOTEST) -v -coverprofile=coverage.out ./...
		$(GOCMD) tool cover -html=coverage.out -o coverage.html
		@echo "📊 Coverage report generated: coverage.html"

.PHONY: test-benchmark
test-benchmark: ## Run Go benchmark tests
		@echo "⚡ Running benchmark tests..."
		$(GOTEST) -bench=. -benchmem ./...

.PHONY: test-all
test-all: test-short test-race test-coverage ## Run all test types

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
