# gomod-rename - Go module import path replacement tool

set dotenv-load := true

VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

# Modules
[doc('Testing (unit, integration, coverage)')]
mod test '.justfiles/test.just'

[doc('Release (tag, publish, GitHub release)')]
mod release '.justfiles/release.just'

# Show available commands
[private]
@default:
    @echo "gomod-rename - Go module import path replacement tool"
    @echo ""
    @just --list --unsorted

# Build the binary
build:
    @mkdir -p bin
    go build -ldflags "-X main.version={{VERSION}}" -o bin/gomod-rename ./cmd/gomod-rename

# Install to GOPATH/bin
install:
    go install -ldflags "-X main.version={{VERSION}}" ./cmd/gomod-rename

# Run linter
lint:
    golangci-lint run ./...

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
check: lint _test build
    @echo "All checks passed!"

[private]
_test:
    go test -v ./...
