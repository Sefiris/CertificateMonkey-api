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
version: ## Show current version information
	@echo "📋 Certificate Monkey Version Information"
	@echo "========================================"
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Go version: $(GO_VERSION)"
	@echo ""
	@echo "🔍 For detailed version analysis, run: make version-preview"
	@echo "🐳 For Docker/Helm tags, run: make version-docker-tags"

version-preview: ## Preview next version based on conventional commits
	@./scripts/version-manager.sh preview

version-bump-auto: ## Automatically bump version based on conventional commits
	@./scripts/version-manager.sh bump auto

version-bump-patch: ## Bump patch version (0.1.0 -> 0.1.1)
	@./scripts/version-manager.sh bump patch

version-bump-minor: ## Bump minor version (0.1.0 -> 0.2.0)
	@./scripts/version-manager.sh bump minor

version-bump-major: ## Bump major version (0.1.0 -> 1.0.0)
	@./scripts/version-manager.sh bump major

version-tag: ## Create git tag for current version
	@./scripts/version-manager.sh tag

version-release: ## Complete release process (bump + commit + tag)
	@./scripts/version-manager.sh release

version-docker-tags: ## Preview Docker/Helm tags for current version
	@./scripts/version-manager.sh docker-tags

# Legacy version commands (deprecated but maintained for compatibility)
version-patch: version-bump-patch ## [DEPRECATED] Use version-bump-patch instead

version-minor: version-bump-minor ## [DEPRECATED] Use version-bump-minor instead

version-major: version-bump-major ## [DEPRECATED] Use version-bump-major instead

changelog-prepare: ## Validate changelog is ready for current version
	@echo "📝 Validating CHANGELOG.md for version $(CURRENT_VERSION)..."
	@if ! grep -q "## \[$(CURRENT_VERSION)\]" CHANGELOG.md; then \
		echo "❌ Version $(CURRENT_VERSION) not found in CHANGELOG.md"; \
		echo "💡 Run 'make version-bump-auto' to automatically update changelog"; \
		exit 1; \
	fi
	@echo "✅ CHANGELOG.md is ready for version $(CURRENT_VERSION)"

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
	@rm -f CHANGELOG.md.backup
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

# Helm commands (future implementation)
# When Helm charts are created, uncomment and customize these commands:
#
# helm-lint: ## Lint Helm chart
# 	@echo "🔍 Linting Helm chart..."
# 	@helm lint helm/certificate-monkey
#
# helm-package: ## Package Helm chart with current version
# 	@echo "📦 Packaging Helm chart..."
# 	@sed -i.bak "s/^appVersion:.*$$/appVersion: \"$(CURRENT_VERSION)\"/" helm/certificate-monkey/Chart.yaml
# 	@helm package helm/certificate-monkey
# 	@rm -f helm/certificate-monkey/Chart.yaml.bak
#
# helm-install: ## Install chart locally for testing
# 	@echo "🚀 Installing Helm chart..."
# 	@helm install certificate-monkey-test ./helm/certificate-monkey \
# 		--set image.tag=$(CURRENT_VERSION) \
# 		--set apiKeys.primary=test-key
#
# helm-test: ## Run Helm chart tests
# 	@echo "🧪 Testing Helm chart..."
# 	@helm test certificate-monkey-test
#
# helm-uninstall: ## Uninstall test chart
# 	@helm uninstall certificate-monkey-test
#
# See docs/HELM_INTEGRATION.md for complete Helm integration guide

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
release-prepare: version-preview swagger-gen test lint-full ## Prepare for release (run tests, generate docs, preview version)
	@echo ""
	@echo "🚀 Release Preparation Complete"
	@echo "==============================="
	@echo "✅ Version analysis completed"
	@echo "✅ Tests passed"
	@echo "✅ Linting passed"
	@echo "✅ Documentation generated"
	@echo ""
	@echo "💡 Next steps:"
	@echo "1. Review version preview above"
	@echo "2. Run 'make version-bump-auto' to bump version and update changelog"
	@echo "3. Review and commit changes"
	@echo "4. Run 'make version-tag' to create git tag"
	@echo "5. Push with 'git push origin main --tags'"

release-auto: ## Automated release process with conventional commits
	@echo "🚀 Starting automated release process..."
	@make test
	@make lint-full
	@make swagger-gen
	@./scripts/version-manager.sh release
	@echo ""
	@echo "🎉 Release completed! Don't forget to push:"
	@echo "   git push origin main --tags"

# Conventional commits helpers
commit-help: ## Show conventional commit format help
	@echo "📝 Conventional Commit Format"
	@echo "============================"
	@echo ""
	@echo "Format: <type>[optional scope]: <description>"
	@echo ""
	@echo "Types:"
	@echo "  feat:     ✨ A new feature (minor version bump)"
	@echo "  fix:      🐛 A bug fix (patch version bump)"
	@echo "  docs:     📚 Documentation only changes"
	@echo "  style:    💄 Code style changes (formatting, etc.)"
	@echo "  refactor: ♻️  Code refactoring"
	@echo "  perf:     ⚡ Performance improvements"
	@echo "  test:     🧪 Adding or updating tests"
	@echo "  build:    🔧 Build system or dependencies"
	@echo "  ci:       👷 CI/CD changes"
	@echo "  chore:    🔨 Other changes (maintenance, etc.)"
	@echo "  revert:   ⏪ Reverting previous changes"
	@echo ""
	@echo "Breaking changes:"
	@echo "  Add '!' after type: feat!: breaking change"
	@echo "  Or add 'BREAKING CHANGE:' in commit body"
	@echo ""
	@echo "Examples:"
	@echo "  feat: add user authentication system"
	@echo "  fix: resolve memory leak in certificate processing"
	@echo "  docs: update API documentation with new endpoints"
	@echo "  feat!: redesign API endpoints (breaking change)"

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
	@echo ""
	@echo "📝 Version Management:"
	@echo "  make version-preview     - Preview next version"
	@echo "  make version-bump-auto   - Auto-bump based on commits"
	@echo "  make commit-help         - Show commit format help"
	@echo "  make release-auto        - Complete automated release"
