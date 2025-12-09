package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var version = "dev"

func main() {
	// Define flags
	write := flag.Bool("write", false, "Apply changes (default is dry-run preview)")
	flag.BoolVar(write, "w", false, "Apply changes (shorthand)")

	dir := flag.String("dir", ".", "Target directory to search")
	flag.StringVar(dir, "d", ".", "Target directory (shorthand)")

	yes := flag.Bool("yes", false, "Skip confirmation prompt")
	flag.BoolVar(yes, "y", false, "Skip confirmation (shorthand)")

	verbose := flag.Bool("verbose", false, "Show detailed output")
	flag.BoolVar(verbose, "v", false, "Verbose output (shorthand)")

	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gomod-rename - Replace Go module import paths\n\n")
		fmt.Fprintf(os.Stderr, "Usage: gomod-rename [flags] <old-path> <new-path>\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  gomod-rename github.com/old/repo github.com/new/repo\n")
		fmt.Fprintf(os.Stderr, "  gomod-rename -w github.com/old/repo github.com/new/repo\n")
		fmt.Fprintf(os.Stderr, "  gomod-rename -d ./myproject -w -y github.com/old/repo github.com/new/repo\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("gomod-rename %s\n", version)
		os.Exit(0)
	}

	// Validate arguments
	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	oldPath := args[0]
	newPath := args[1]

	if oldPath == newPath {
		fmt.Fprintln(os.Stderr, "Error: old-path and new-path are identical")
		os.Exit(1)
	}

	// Resolve target directory
	targetDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
		os.Exit(1)
	}

	// Check directory exists
	info, err := os.Stat(targetDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid directory\n", targetDir)
		os.Exit(1)
	}

	// Find matching files
	matches, err := FindFiles(targetDir, oldPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning files: %v\n", err)
		os.Exit(1)
	}

	if len(matches) == 0 {
		fmt.Printf("No files in '%s' contain '%s'\n", targetDir, oldPath)
		os.Exit(0)
	}

	// Always show preview
	PrintPreview(matches, oldPath, newPath, *verbose)

	// If not writing, we're done
	if !*write {
		fmt.Println()
		fmt.Println("This is a dry-run. Use --write (-w) to apply changes.")
		os.Exit(0)
	}

	// Confirm before writing
	if !*yes {
		fmt.Println()
		fmt.Print("Apply these changes? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if response != "y\n" && response != "Y\n" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	// Apply changes
	fmt.Println()
	fmt.Println("Applying changes...")
	results := ReplaceInFiles(matches, oldPath, newPath)
	PrintResults(results)
}
