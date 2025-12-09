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
    go build -ldflags "-X main.version={{VERSION}}" -o gomod-rename .

# Install to GOPATH/bin
install:
    go install -ldflags "-X main.version={{VERSION}}" .

# Run tests
test:
    go test -v ./...

# Run linter
lint:
    go vet ./...

# Format code
fmt:
    go fmt ./...

# Clean build artifacts
clean:
    rm -f gomod-rename

# Build and run with arguments
run *ARGS:
    go run -ldflags "-X main.version={{VERSION}}" . {{ARGS}}
