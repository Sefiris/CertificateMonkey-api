# Certificate Monkey - Development Makefile

.PHONY: help build test test-cover swagger-install swagger-gen swagger-serve clean run dev lint

# Default target
help: ## Show this help message
	@echo "Certificate Monkey - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build commands
build: ## Build the application
	@echo "🔧 Building Certificate Monkey..."
	@go build -o certificate-monkey cmd/server/main.go
	@echo "✅ Build complete"

build-linux: ## Build for Linux (useful for Docker)
	@echo "🔧 Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o certificate-monkey-linux cmd/server/main.go
	@echo "✅ Linux build complete"

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
	@echo "💡 Press Ctrl+C to stop"
	@echo ""
	@./scripts/start-swagger-demo.sh

# Development commands
run: build ## Build and run the application
	@echo "🚀 Starting Certificate Monkey..."
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
	@golangci-lint run
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

# Docker commands (if needed)
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t certificate-monkey:latest .
	@echo "✅ Docker image built"

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

# Information
info: ## Show project information
	@echo "📋 Certificate Monkey Project Information"
	@echo "========================================"
	@echo "Go version: $$(go version)"
	@echo "Project: Certificate Management API"
	@echo "License: MIT"
	@echo ""
	@echo "🔗 Key URLs:"
	@echo "  Health:     http://localhost:8080/health"
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
