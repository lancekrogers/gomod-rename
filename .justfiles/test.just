# Testing recipes

# Show available test commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Run unit tests
[no-cd]
all:
    go test -v ./...

# Run integration tests
[no-cd]
integration:
    go test -v -tags=integration ./...

# Run tests with coverage
[no-cd]
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    @echo ""
    @echo "HTML report: go tool cover -html=coverage.out"

# Generate HTML coverage report
[no-cd]
coverage-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Run tests matching pattern
[no-cd]
match PATTERN:
    go test -v -run "{{PATTERN}}" ./...

# Run tests and stop on first failure
[no-cd]
fail-fast:
    go test -v -failfast ./...
