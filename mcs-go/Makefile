.PHONY: build install clean test test-unit test-integration test-e2e test-coverage test-clean test-bench run dev

# Variables
BINARY_NAME=mcs
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG=$(shell git describe --exact-match --tags 2>/dev/null || echo "")
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "true" || echo "false")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags
LDFLAGS=-ldflags "\
	-X github.com/michaelkeevildown/mcs/internal/version.Version=$(VERSION) \
	-X github.com/michaelkeevildown/mcs/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/michaelkeevildown/mcs/internal/version.GitTag=$(GIT_TAG) \
	-X github.com/michaelkeevildown/mcs/internal/version.GitDirty=$(GIT_DIRTY) \
	-X github.com/michaelkeevildown/mcs/internal/version.BuildTime=$(BUILD_TIME)"

INSTALL_PATH=$(HOME)/.mcs/bin

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME) $(VERSION)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/mcs/main.go
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

# Install to user's .mcs/bin directory
install: build
	@echo "📦 Installing to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp bin/$(BINARY_NAME) $(INSTALL_PATH)/
	@echo "✅ Installed to $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo ""
	@echo "🎯 Add this to your PATH if not already done:"
	@echo "  export PATH=\"$(INSTALL_PATH):$$PATH\""

# Development build (with race detector)
dev:
	@echo "🔧 Building development version..."
	@go build -race $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/mcs/main.go
	@echo "✅ Development build complete"

# Run the binary
run: build
	@./bin/$(BINARY_NAME)

# Run all tests
test:
	@echo "🧪 Running all tests..."
	@go test -v ./...

# Run unit tests with coverage
test-unit:
	@echo "🧪 Running unit tests with coverage..."
	@mkdir -p test/coverage/unit test/reports
	@go test -race -coverprofile=test/coverage/unit/coverage.out -covermode=atomic ./...
	@go tool cover -html=test/coverage/unit/coverage.out -o test/reports/coverage.html
	@echo "✅ Unit tests complete - coverage report: test/reports/coverage.html"

# Run integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	@mkdir -p test/coverage/integration test/reports
	@go test -tags=integration -race -coverprofile=test/coverage/integration/coverage.out -covermode=atomic ./...
	@echo "✅ Integration tests complete"

# Run end-to-end tests
test-e2e:
	@echo "🧪 Running end-to-end tests..."
	@mkdir -p test/coverage/e2e test/reports
	@go test -tags=e2e -race -coverprofile=test/coverage/e2e/coverage.out -covermode=atomic ./...
	@echo "✅ End-to-end tests complete"

# Generate comprehensive coverage report
test-coverage:
	@echo "📊 Generating coverage reports..."
	@./test/scripts/coverage.sh
	@echo "✅ Coverage reports generated"

# Run benchmarks
test-bench:
	@echo "⚡ Running benchmarks..."
	@mkdir -p test/reports
	@go test -bench=. -benchmem -cpuprofile=test/reports/cpu.prof -memprofile=test/reports/mem.prof ./...
	@echo "✅ Benchmarks complete"

# Clean test artifacts
test-clean:
	@echo "🧹 Cleaning test artifacts..."
	@rm -rf test/coverage/unit/*.out test/coverage/integration/*.out test/coverage/e2e/*.out
	@rm -rf test/reports/*.html test/reports/*.xml test/reports/*.json test/reports/*.prof
	@rm -rf test/fixtures/temp test/integration/temp .test-cache
	@echo "✅ Test artifacts cleaned"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@echo "✅ Clean complete"

# Build for all platforms
build-all:
	@echo "🌍 Building for all platforms..."
	@mkdir -p dist
	
	@echo "  • Building Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 cmd/mcs/main.go
	
	@echo "  • Building Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 cmd/mcs/main.go
	
	@echo "  • Building macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 cmd/mcs/main.go
	
	@echo "  • Building macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 cmd/mcs/main.go
	
	@echo "  • Building Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/mcs/main.go
	
	@echo "✅ All builds complete"

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	@go mod download
	@echo "✅ Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "🧹 Tidying dependencies..."
	@go mod tidy
	@echo "✅ Dependencies tidied"

# Version helpers
.PHONY: version version-patch version-minor version-major

# Display current version
version:
	@echo "Current version: $(VERSION)"

# Create a patch release (v0.1.0 -> v0.1.1)
version-patch:
	@echo "Creating patch release..."
	@./scripts/release.sh patch

# Create a minor release (v0.1.0 -> v0.2.0)
version-minor:
	@echo "Creating minor release..."
	@./scripts/release.sh minor

# Create a major release (v0.1.0 -> v1.0.0)
version-major:
	@echo "Creating major release..."
	@./scripts/release.sh major

# Create a pre-release
version-pre:
	@echo "Creating pre-release..."
	@./scripts/release.sh pre