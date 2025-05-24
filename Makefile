# Certificate Monkey - Development Makefile

# Version management
CURRENT_VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0-dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags for version information
LDFLAGS := -X 'certificate-monkey/internal/version.Version=$(CURRENT_VERSION)' \
           -X 'certificate-monkey/internal/version.BuildTime=$(BUILD_TIME)' \
           -X 'certificate-monkey/internal/version.GitCommit=$(GIT_COMMIT)' \
           -X 'certificate-monkey/internal/version.GoVersion=$(GO_VERSION)'

.PHONY: help build test test-cover swagger-install swagger-gen swagger-serve clean run dev lint version

# Default target
help: ## Show this help message
	@echo "Certificate Monkey - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Version management commands
version: ## Show current version
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Go version: $(GO_VERSION)"

version-patch: ## Bump patch version (0.1.0 -> 0.1.1)
	@echo "Bumping patch version..."
	@current=$$(cat VERSION); \
	new_version=$$(echo $$current | awk -F. '{$$3=$$3+1; print $$1"."$$2"."$$3}'); \
	echo $$new_version > VERSION; \
	echo "Version bumped from $$current to $$new_version"

version-minor: ## Bump minor version (0.1.0 -> 0.2.0)
	@echo "Bumping minor version..."
	@current=$$(cat VERSION); \
	new_version=$$(echo $$current | awk -F. '{$$2=$$2+1; $$3=0; print $$1"."$$2"."$$3}'); \
	echo $$new_version > VERSION; \
	echo "Version bumped from $$current to $$new_version"

version-major: ## Bump major version (0.1.0 -> 1.0.0)
	@echo "Bumping major version..."
	@current=$$(cat VERSION); \
	new_version=$$(echo $$current | awk -F. '{$$1=$$1+1; $$2=0; $$3=0; print $$1"."$$2"."$$3}'); \
	echo $$new_version > VERSION; \
	echo "Version bumped from $$current to $$new_version"

changelog-prepare: ## Prepare changelog for new release
	@echo "Preparing CHANGELOG.md for version $(CURRENT_VERSION)..."
	@if ! grep -q "## \[$(CURRENT_VERSION)\]" CHANGELOG.md; then \
		echo "Version $(CURRENT_VERSION) not found in CHANGELOG.md"; \
		echo "Please add an entry for this version in CHANGELOG.md"; \
		exit 1; \
	fi
	@echo "CHANGELOG.md is ready for version $(CURRENT_VERSION)"

# Build commands
build: ## Build the application with version information
	@echo "🔧 Building Certificate Monkey v$(CURRENT_VERSION)..."
	@go build -ldflags "$(LDFLAGS)" -o certificate-monkey cmd/server/main.go
	@echo "✅ Build complete"

build-linux: ## Build for Linux with version information
	@echo "🔧 Building for Linux v$(CURRENT_VERSION)..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o certificate-monkey-linux cmd/server/main.go
	@echo "✅ Linux build complete"

build-release: ## Build optimized release binary
	@echo "🔧 Building release binary v$(CURRENT_VERSION)..."
	@go build -ldflags "$(LDFLAGS) -s -w" -o certificate-monkey cmd/server/main.go
	@echo "✅ Release build complete"

# Test commands
test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test ./...

test-cover: ## Run tests with coverage report
	@echo "🧪 Running tests with coverage..."
	@go test -cover ./...
	@echo ""
	@echo "📊 Generating detailed coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-verbose: ## Run tests with verbose output
	@echo "🧪 Running verbose tests..."
	@go test -v ./...

# Swagger documentation
swagger-install: ## Install swag CLI tool
	@echo "📦 Installing swag CLI..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Swag CLI installed"

swagger-gen: ## Generate Swagger documentation
	@echo "📝 Generating Swagger documentation..."
	@$(HOME)/go/bin/swag init -g cmd/server/main.go -o docs --parseInternal
	@echo "✅ Swagger docs generated in docs/ directory"

swagger-serve: swagger-gen build ## Generate docs and start server with Swagger UI
	@echo "🚀 Starting server with Swagger UI..."
	@echo "📖 Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "🏥 Health Check: http://localhost:8080/health"
	@echo "📊 Build Info: http://localhost:8080/build-info"
	@echo "💡 Press Ctrl+C to stop"
	@echo ""
	@./scripts/start-swagger-demo.sh

# Development commands
run: build ## Build and run the application
	@echo "🚀 Starting Certificate Monkey v$(CURRENT_VERSION)..."
	@./certificate-monkey

dev: swagger-gen ## Start development environment
	@echo "🔄 Starting development environment..."
	@echo "📝 Swagger docs regenerated"
	@echo "🚀 Starting server..."
	@./scripts/start-swagger-demo.sh

# Code quality
lint: ## Run linting
	@echo "🔍 Running linters..."
	@go vet ./...
	@go fmt ./...
	@echo "✅ Linting complete"

lint-install: ## Install golangci-lint
	@echo "📦 Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ golangci-lint installed"

lint-full: ## Run comprehensive linting (requires golangci-lint)
	@echo "🔍 Running comprehensive linting..."
	@$(HOME)/go/bin/golangci-lint run --exclude="fieldalignment:"
	@echo "✅ Comprehensive linting complete"

# Utility commands
clean: ## Clean build artifacts and temporary files
	@echo "🧹 Cleaning up..."
	@rm -f certificate-monkey certificate-monkey-linux
	@rm -f coverage.out coverage.html
	@rm -rf docs/docs.go docs/swagger.json docs/swagger.yaml
	@echo "✅ Cleanup complete"

deps: ## Download and tidy dependencies
	@echo "📦 Managing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

# Docker commands
docker-build: ## Build Docker image with version tags
	@echo "🐳 Building Docker image v$(CURRENT_VERSION)..."
	@docker build \
		--build-arg VERSION=$(CURRENT_VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		-t certificate-monkey:latest \
		-t certificate-monkey:$(CURRENT_VERSION) .
	@echo "✅ Docker image built with tags: latest, $(CURRENT_VERSION)"

docker-run: docker-build ## Build and run Docker container
	@echo "🐳 Running Certificate Monkey container..."
	@docker run -d --name certificate-monkey -p 8080:8080 \
		-e API_KEY_1=demo-api-key-12345 \
		-e DYNAMODB_TABLE=certificate-monkey-dev \
		-e KMS_KEY_ID=alias/certificate-monkey-dev \
		certificate-monkey:latest
	@echo "✅ Container started on http://localhost:8080"
	@echo "🏥 Health Check: http://localhost:8080/health"
	@echo "📖 Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "📊 Build Info: http://localhost:8080/build-info"

docker-stop: ## Stop and remove Docker container
	@echo "🐳 Stopping Certificate Monkey container..."
	@docker stop certificate-monkey || true
	@docker rm certificate-monkey || true
	@echo "✅ Container stopped and removed"

docker-logs: ## View Docker container logs
	@echo "📋 Certificate Monkey container logs:"
	@docker logs certificate-monkey

docker-test: docker-build ## Test Docker container health
	@echo "🧪 Testing Docker container..."
	@docker run -d --name certificate-monkey-test -p 8081:8080 \
		-e API_KEY_1=test-key \
		-e DYNAMODB_TABLE=test-table \
		-e KMS_KEY_ID=test-key \
		certificate-monkey:latest
	@sleep 5
	@if curl -f http://localhost:8081/health; then \
		echo "✅ Container health check passed"; \
	else \
		echo "❌ Container health check failed"; \
		docker logs certificate-monkey-test; \
		exit 1; \
	fi
	@docker stop certificate-monkey-test
	@docker rm certificate-monkey-test

docker-clean: ## Clean up Docker images and containers
	@echo "🧹 Cleaning Docker artifacts..."
	@docker stop certificate-monkey certificate-monkey-test 2>/dev/null || true
	@docker rm certificate-monkey certificate-monkey-test 2>/dev/null || true
	@docker rmi certificate-monkey:latest certificate-monkey:$(CURRENT_VERSION) 2>/dev/null || true
	@echo "✅ Docker cleanup complete"

# Scripts
demo: ## Run the complete demo
	@echo "🎪 Starting Certificate Monkey demo..."
	@./scripts/start-swagger-demo.sh

test-tags: ## Test tag search functionality
	@echo "🔍 Testing tag search..."
	@./scripts/test-tag-search.sh

test-workflow: ## Test complete certificate workflow
	@echo "📋 Testing certificate workflow..."
	@./scripts/test-pfx-workflow.sh

test-private-key: ## Test private key export functionality
	@echo "🔐 Testing private key export..."
	@./scripts/test-private-key-export.sh

# Release management
release-prepare: version changelog-prepare swagger-gen test ## Prepare for release (run tests, generate docs)
	@echo "🚀 Preparing release v$(CURRENT_VERSION)..."
	@echo "✅ All checks passed - ready for release!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Review CHANGELOG.md"
	@echo "2. git add ."
	@echo "3. git commit -m 'Release v$(CURRENT_VERSION)'"
	@echo "4. git tag v$(CURRENT_VERSION)"
	@echo "5. git push origin main --tags"

# Information
info: ## Show project information
	@echo "📋 Certificate Monkey Project Information"
	@echo "========================================"
	@echo "Version: $(CURRENT_VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "License: MIT"
	@echo ""
	@echo "🔗 Key URLs:"
	@echo "  Health:     http://localhost:8080/health"
	@echo "  Build Info: http://localhost:8080/build-info"
	@echo "  Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  API Base:   http://localhost:8080/api/v1"
	@echo ""
	@echo "📋 API Endpoints:"
	@echo "  POST   /api/v1/keys                      - Create private key and CSR"
	@echo "  GET    /api/v1/keys                      - List certificates with filtering"
	@echo "  GET    /api/v1/keys/{id}                 - Get certificate details"
	@echo "  GET    /api/v1/keys/{id}/private-key     - Export private key (SENSITIVE)"
	@echo "  PUT    /api/v1/keys/{id}/certificate     - Upload certificate"
	@echo "  POST   /api/v1/keys/{id}/pfx             - Generate PFX file"
	@echo ""
	@echo "🔑 Demo API Keys:"
	@echo "  demo-api-key-12345"
	@echo "  swagger-test-key"
