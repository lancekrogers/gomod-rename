# gomod-rename

Rename Go module import paths across an entire codebase in one command.

## Features

- **Dry-run by default** — preview every match before writing anything
- **Confirmation prompt** — asks before applying changes (skip with `-y`)
- **Targets go.mod and .go files** — handles both module declarations and import statements
- **Skips vendor/ and hidden directories** — won't touch files you don't own
- **Preserves file permissions** — replacement writes respect original file modes
- **Zero external dependencies** — stdlib only

## Installation

Requires Go 1.23+.

```bash
go install github.com/lancekrogers/gomod-rename/cmd/gomod-rename@latest
```

## Usage

```bash
# Dry-run (default) — see what would change
gomod-rename github.com/old/module github.com/new/module

# Apply changes
gomod-rename -w github.com/old/module github.com/new/module

# Skip confirmation prompt
gomod-rename -w -y github.com/old/module github.com/new/module

# Target a specific directory
gomod-rename -d ./myproject -w github.com/old/module github.com/new/module

# Verbose output (show all matches per file)
gomod-rename -v github.com/old/module github.com/new/module
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--write` | `-w` | Apply changes (default is dry-run preview) |
| `--dir` | `-d` | Target directory to search (default: `.`) |
| `--yes` | `-y` | Skip confirmation prompt |
| `--verbose` | `-v` | Show detailed output |
| `--version` | | Show version |

## Building from source

```bash
git clone https://github.com/lancekrogers/gomod-rename.git
cd gomod-rename

# With just (recommended)
just build

# Or directly with go
go build -o bin/gomod-rename ./cmd/gomod-rename
```

## Testing

```bash
# Run tests
just test

# Lint + test + build
just check

# Coverage report
just test-coverage
```

## License

[MIT](LICENSE)
