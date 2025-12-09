// Package cli provides the command-line interface for gomod-rename.
package cli

import (
	"bufio"
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
func (c *CLI) Run(args []string) int {
	cfg, err := c.parseFlags(args)
	if err != nil {
		if err == errShowVersion {
			fmt.Fprintf(c.Stdout, "gomod-rename %s\n", c.Version)
			return 0
		}
		if err == errShowUsage {
			return 0
		}
		fmt.Fprintf(c.Stderr, "Error: %v\n", err)
		return 1
	}

	// Validate arguments
	if err := c.validate(cfg); err != nil {
		fmt.Fprintf(c.Stderr, "Error: %v\n", err)
		return 1
	}

	// Find matching files
	matches, err := finder.Find(cfg.Dir, cfg.OldPath, nil)
	if err != nil {
		fmt.Fprintf(c.Stderr, "Error scanning files: %v\n", err)
		return 1
	}

	if len(matches) == 0 {
		fmt.Fprintf(c.Stdout, "No files in '%s' contain '%s'\n", cfg.Dir, cfg.OldPath)
		return 0
	}

	// Show preview
	c.printPreview(matches, cfg)

	// If dry-run, we're done
	if !cfg.Write {
		fmt.Fprintln(c.Stdout)
		fmt.Fprintln(c.Stdout, "This is a dry-run. Use --write (-w) to apply changes.")
		return 0
	}

	// Confirm before writing
	if !cfg.Yes {
		fmt.Fprintln(c.Stdout)
		fmt.Fprint(c.Stdout, "Apply these changes? [y/N]: ")
		if !c.confirm() {
			fmt.Fprintln(c.Stdout, "Aborted.")
			return 0
		}
	}

	// Apply changes
	fmt.Fprintln(c.Stdout)
	fmt.Fprintln(c.Stdout, "Applying changes...")
	results := replacer.Replace(matches, cfg.OldPath, cfg.NewPath)
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
		fmt.Fprintf(c.Stderr, "gomod-rename - Replace Go module import paths\n\n")
		fmt.Fprintf(c.Stderr, "Usage: gomod-rename [flags] <old-path> <new-path>\n\n")
		fmt.Fprintf(c.Stderr, "Examples:\n")
		fmt.Fprintf(c.Stderr, "  gomod-rename github.com/old/repo github.com/new/repo\n")
		fmt.Fprintf(c.Stderr, "  gomod-rename -w github.com/old/repo github.com/new/repo\n")
		fmt.Fprintf(c.Stderr, "  gomod-rename -d ./myproject -w -y github.com/old/repo github.com/new/repo\n\n")
		fmt.Fprintf(c.Stderr, "Flags:\n")
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
	fmt.Fprintf(c.Stdout, "Files containing '%s':\n\n", cfg.OldPath)

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
		fmt.Fprintf(c.Stdout, "  [%s] %s\n", fileType, fm.Path)

		// Show matches (limit to 5 per file unless verbose)
		limit := 5
		if cfg.Verbose {
			limit = len(fm.Matches)
		}

		for i, m := range fm.Matches {
			if i >= limit {
				remaining := len(fm.Matches) - limit
				fmt.Fprintf(c.Stdout, "         ... and %d more matches\n", remaining)
				break
			}
			fmt.Fprintf(c.Stdout, "    %4d: %s\n", m.LineNum, strings.TrimSpace(m.Line))
		}
		fmt.Fprintln(c.Stdout)
	}

	// Summary
	fmt.Fprintln(c.Stdout, "---")
	fmt.Fprintf(c.Stdout, "Summary: %d matches in %d files (%d go.mod, %d .go)\n",
		totalMatches, len(matches), goModCount, goFileCount)
	fmt.Fprintf(c.Stdout, "Will replace: '%s' -> '%s'\n", cfg.OldPath, cfg.NewPath)
}

func (c *CLI) printResults(results []replacer.Result) {
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(c.Stderr, "  ERROR: %s: %v\n", r.Path, r.Err)
		} else if r.Count > 0 {
			fmt.Fprintf(c.Stdout, "  Updated: %s (%d replacements)\n", r.Path, r.Count)
		}
	}

	summary := replacer.Summarize(results)
	fmt.Fprintln(c.Stdout)
	fmt.Fprintf(c.Stdout, "Done: %d files updated, %d total replacements", summary.FilesModified, summary.TotalReplaced)
	if summary.Errors > 0 {
		fmt.Fprintf(c.Stdout, ", %d errors", summary.Errors)
	}
	fmt.Fprintln(c.Stdout)
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
