// Package cli provides the command-line interface for gomod-rename.
package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/gomod-rename/internal/finder"
	"github.com/lancekrogers/gomod-rename/internal/replacer"
)

// CLI encapsulates the command-line interface.
type CLI struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Stdin   io.Reader
	Version string
}

// New creates a new CLI instance with the given I/O streams.
func New(stdout, stderr io.Writer, stdin io.Reader, version string) *CLI {
	return &CLI{
		Stdout:  stdout,
		Stderr:  stderr,
		Stdin:   stdin,
		Version: version,
	}
}

// Output helpers — stdout/stderr writes in a CLI cannot meaningfully recover
// from write errors, so we explicitly discard them (similar to log.Println).

func (c *CLI) printf(format string, a ...any) {
	_, _ = fmt.Fprintf(c.Stdout, format, a...)
}

func (c *CLI) println(a ...any) {
	_, _ = fmt.Fprintln(c.Stdout, a...)
}

func (c *CLI) print(a ...any) {
	_, _ = fmt.Fprint(c.Stdout, a...)
}

func (c *CLI) errf(format string, a ...any) {
	_, _ = fmt.Fprintf(c.Stderr, format, a...)
}

// Config holds the parsed command-line configuration.
type Config struct {
	Write   bool
	Dir     string
	Yes     bool
	Verbose bool
	OldPath string
	NewPath string
}

// Run executes the CLI with the given arguments and returns an exit code.
func (c *CLI) Run(ctx context.Context, args []string) int {
	cfg, err := c.parseFlags(args)
	if err != nil {
		if err == errShowVersion {
			c.printf("gomod-rename %s\n", c.Version)
			return 0
		}
		if err == errShowUsage {
			return 0
		}
		c.errf("Error: %v\n", err)
		return 1
	}

	// Validate arguments
	if err := c.validate(cfg); err != nil {
		c.errf("Error: %v\n", err)
		return 1
	}

	// Find matching files
	matches, err := finder.Find(ctx, cfg.Dir, cfg.OldPath, nil)
	if err != nil {
		c.errf("Error scanning files: %v\n", err)
		return 1
	}

	if len(matches) == 0 {
		c.printf("No files in '%s' contain '%s'\n", cfg.Dir, cfg.OldPath)
		return 0
	}

	// Show preview
	c.printPreview(matches, cfg)

	// If dry-run, we're done
	if !cfg.Write {
		c.println()
		c.println("This is a dry-run. Use --write (-w) to apply changes.")
		return 0
	}

	// Confirm before writing
	if !cfg.Yes {
		c.println()
		c.print("Apply these changes? [y/N]: ")
		if !c.confirm() {
			c.println("Aborted.")
			return 0
		}
	}

	// Apply changes
	c.println()
	c.println("Applying changes...")
	results := replacer.Replace(ctx, matches, cfg.OldPath, cfg.NewPath)
	c.printResults(results)

	// Return error code if any replacements failed
	summary := replacer.Summarize(results)
	if summary.Errors > 0 {
		return 1
	}

	return 0
}

var (
	errShowVersion = fmt.Errorf("show version")
	errShowUsage   = fmt.Errorf("show usage")
)

func (c *CLI) parseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("gomod-rename", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	cfg := &Config{}

	fs.BoolVar(&cfg.Write, "write", false, "Apply changes (default is dry-run preview)")
	fs.BoolVar(&cfg.Write, "w", false, "Apply changes (shorthand)")

	fs.StringVar(&cfg.Dir, "dir", ".", "Target directory to search")
	fs.StringVar(&cfg.Dir, "d", ".", "Target directory (shorthand)")

	fs.BoolVar(&cfg.Yes, "yes", false, "Skip confirmation prompt")
	fs.BoolVar(&cfg.Yes, "y", false, "Skip confirmation (shorthand)")

	fs.BoolVar(&cfg.Verbose, "verbose", false, "Show detailed output")
	fs.BoolVar(&cfg.Verbose, "v", false, "Verbose output (shorthand)")

	showVersion := fs.Bool("version", false, "Show version")

	fs.Usage = func() {
		c.errf("gomod-rename - Replace Go module import paths\n\n")
		c.errf("Usage: gomod-rename [flags] <old-path> <new-path>\n\n")
		c.errf("Examples:\n")
		c.errf("  gomod-rename github.com/old/repo github.com/new/repo\n")
		c.errf("  gomod-rename -w github.com/old/repo github.com/new/repo\n")
		c.errf("  gomod-rename -d ./myproject -w -y github.com/old/repo github.com/new/repo\n\n")
		c.errf("Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil, errShowUsage
		}
		return nil, err
	}

	if *showVersion {
		return nil, errShowVersion
	}

	// Get positional arguments
	posArgs := fs.Args()
	if len(posArgs) != 2 {
		fs.Usage()
		return nil, fmt.Errorf("requires exactly 2 arguments: old-path and new-path")
	}

	cfg.OldPath = posArgs[0]
	cfg.NewPath = posArgs[1]

	return cfg, nil
}

func (c *CLI) validate(cfg *Config) error {
	if cfg.OldPath == cfg.NewPath {
		return fmt.Errorf("old-path and new-path are identical")
	}

	// Resolve and validate directory
	absDir, err := filepath.Abs(cfg.Dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}
	cfg.Dir = absDir

	return nil
}

func (c *CLI) printPreview(matches []finder.FileMatch, cfg *Config) {
	c.printf("Files containing '%s':\n\n", cfg.OldPath)

	totalMatches := 0
	goModCount := 0
	goFileCount := 0

	for _, fm := range matches {
		totalMatches += len(fm.Matches)
		if fm.IsGoMod {
			goModCount++
		} else {
			goFileCount++
		}

		// File header
		fileType := ".go"
		if fm.IsGoMod {
			fileType = "go.mod"
		}
		c.printf("  [%s] %s\n", fileType, fm.Path)

		// Show matches (limit to 5 per file unless verbose)
		limit := 5
		if cfg.Verbose {
			limit = len(fm.Matches)
		}

		for i, m := range fm.Matches {
			if i >= limit {
				remaining := len(fm.Matches) - limit
				c.printf("         ... and %d more matches\n", remaining)
				break
			}
			c.printf("    %4d: %s\n", m.LineNum, strings.TrimSpace(m.Line))
		}
		c.println()
	}

	// Summary
	c.println("---")
	c.printf("Summary: %d matches in %d files (%d go.mod, %d .go)\n",
		totalMatches, len(matches), goModCount, goFileCount)
	c.printf("Will replace: '%s' -> '%s'\n", cfg.OldPath, cfg.NewPath)
}

func (c *CLI) printResults(results []replacer.Result) {
	for _, r := range results {
		if r.Err != nil {
			c.errf("  ERROR: %s: %v\n", r.Path, r.Err)
		} else if r.Count > 0 {
			c.printf("  Updated: %s (%d replacements)\n", r.Path, r.Count)
		}
	}

	summary := replacer.Summarize(results)
	c.println()
	c.printf("Done: %d files updated, %d total replacements", summary.FilesModified, summary.TotalReplaced)
	if summary.Errors > 0 {
		c.printf(", %d errors", summary.Errors)
	}
	c.println()
}

func (c *CLI) confirm() bool {
	reader := bufio.NewReader(c.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
