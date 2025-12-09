# gomod-rename - Go module import path replacement tool

set dotenv-load := true

VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

# Show available commands
[private]
@default:
    @echo "gomod-rename - Go module import path replacement tool"
    @echo ""
    @just --list

# Build the binary
build:
    @mkdir -p bin
    go build -ldflags "-X main.version={{VERSION}}" -o bin/gomod-rename ./cmd/gomod-rename

# Install to GOPATH/bin
install:
    go install -ldflags "-X main.version={{VERSION}}" ./cmd/gomod-rename

# Run all tests
test:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    @echo ""
    @echo "HTML report: go tool cover -html=coverage.out"

# Run tests with coverage and open HTML report
test-coverage-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    open coverage.html

# Run linter
lint:
    go vet ./...

# Format code
fmt:
    go fmt ./...

# Clean build artifacts
clean:
    rm -rf bin/ coverage.out coverage.html

# Build and run with arguments
run *ARGS:
    go run -ldflags "-X main.version={{VERSION}}" ./cmd/gomod-rename {{ARGS}}

# Verify everything passes (lint, test, build)
check: lint test build
    @echo "All checks passed!"
