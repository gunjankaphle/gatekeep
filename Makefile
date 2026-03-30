.PHONY: help build build-all test test-integration test-e2e coverage lint clean run run-server docker-build docker-up docker-down install-tools

# Default target
help:
	@echo "GateKeep - Makefile Commands"
	@echo ""
	@echo "Build:"
	@echo "  make build         - Build CLI and server binaries"
	@echo "  make build-all     - Build binaries for multiple platforms"
	@echo ""
	@echo "Test:"
	@echo "  make test          - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-e2e      - Run end-to-end tests"
	@echo "  make coverage      - Generate coverage report"
	@echo ""
	@echo "Development:"
	@echo "  make lint          - Run linters"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make run           - Run CLI (use ARGS for arguments)"
	@echo "  make run-server    - Run API server"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start Docker services"
	@echo "  make docker-down   - Stop Docker services"
	@echo ""
	@echo "Tools:"
	@echo "  make install-tools - Install development tools"

# Build targets
build:
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o bin/gatekeep ./cmd/cli
	go build -o bin/gatekeep-server ./cmd/server
	@echo "✓ Binaries built: bin/gatekeep, bin/gatekeep-server"

build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/gatekeep-linux-amd64 ./cmd/cli
	GOOS=darwin GOARCH=amd64 go build -o bin/gatekeep-darwin-amd64 ./cmd/cli
	GOOS=darwin GOARCH=arm64 go build -o bin/gatekeep-darwin-arm64 ./cmd/cli
	GOOS=windows GOARCH=amd64 go build -o bin/gatekeep-windows-amd64.exe ./cmd/cli
	@echo "✓ Cross-platform binaries built in bin/"

# Test targets
test:
	@echo "Running unit tests..."
	go test -v -race -short ./...

test-integration:
	@echo "Running integration tests..."
	go test -v -race -run Integration ./...

test-e2e:
	@echo "Running end-to-end tests..."
	go test -v -race -run E2E ./test/...

coverage:
	@echo "Generating coverage report..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Development targets
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m ./...
	@echo "✓ Linting complete"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache
	@echo "✓ Clean complete"

run:
	@go run ./cmd/cli $(ARGS)

run-server:
	@go run ./cmd/server

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t gatekeep:latest .
	@echo "✓ Docker image built: gatekeep:latest"

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "✓ Services started"

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "✓ Services stopped"

# Tool installation
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "✓ Tools installed"
