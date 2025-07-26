.PHONY: build install clean test run dev

# Variables
BINARY_NAME=mcs
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
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

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

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