# Modelplex development tasks

# Default recipe to display available commands
default:
    @just --list

# Build the modelplex binary
build:
    go build -o bin/modelplex ./cmd/modelplex

# Build for multiple platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -o bin/modelplex-linux-amd64 ./cmd/modelplex
    GOOS=darwin GOARCH=amd64 go build -o bin/modelplex-darwin-amd64 ./cmd/modelplex
    GOOS=darwin GOARCH=arm64 go build -o bin/modelplex-darwin-arm64 ./cmd/modelplex
    GOOS=windows GOARCH=amd64 go build -o bin/modelplex-windows-amd64.exe ./cmd/modelplex

# Run the application with default config
run:
    go run ./cmd/modelplex --config config.toml --socket ./modelplex.socket

# Run with verbose logging
run-verbose:
    go run ./cmd/modelplex --config config.toml --socket ./modelplex.socket

# Install dependencies
deps:
    go mod tidy
    go mod download

# Run tests
test:
    go test -v ./...

# Run unit tests only
test-unit:
    go test -v -short ./...

# Run integration tests only  
test-integration:
    go test -v -run Integration ./test/integration/...

# Run tests with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
test-race:
    go test -v -race ./...

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f modelplex.socket
    rm -f coverage.out coverage.html

# Create example config from template
init-config:
    cp config.toml config.local.toml
    @echo "Created config.local.toml - edit with your API keys"

# Test API with curl examples
test-api:
    ./examples/curl/test_api.sh

# Run Python example
test-python:
    cd examples/python && python basic_usage.py

# Start development environment (requires Docker)
dev-start:
    docker run -d --name ollama -p 11434:11434 ollama/ollama
    @echo "Started Ollama container for local model testing"

# Stop development environment
dev-stop:
    docker stop ollama || true
    docker rm ollama || true

# Install development tools
dev-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/goreleaser/goreleaser@latest

# Generate documentation
docs:
    @echo "Generating API documentation..."
    @echo "Visit http://localhost:8080/docs when running"

# Docker build
docker-build:
    docker build -t modelplex:latest .

# Docker run
docker-run:
    docker run -v $(pwd)/config.toml:/app/config.toml -v /tmp/modelplex.socket:/app/modelplex.socket modelplex:latest

# Test release build locally
test-release:
    goreleaser release --snapshot --clean

# Validate GitHub Actions locally (requires act)
test-actions:
    act pull_request

# Run all checks (tests, lint, build)
check-all: test lint build
    @echo "✅ All checks passed!"