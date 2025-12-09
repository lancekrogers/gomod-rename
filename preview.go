package main

import (
	"fmt"
	"strings"
)

// PrintPreview displays what would be changed without modifying files
func PrintPreview(matches []FileMatch, oldPath, newPath string, verbose bool) {
	if len(matches) == 0 {
		fmt.Println("No files contain the specified path.")
		return
	}

	fmt.Printf("Files containing '%s':\n\n", oldPath)

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
		fmt.Printf("  [%s] %s\n", fileType, fm.Path)

		// Show matches (limit to 5 per file unless verbose)
		limit := 5
		if verbose {
			limit = len(fm.Matches)
		}

		for i, m := range fm.Matches {
			if i >= limit {
				remaining := len(fm.Matches) - limit
				fmt.Printf("         ... and %d more matches\n", remaining)
				break
			}

			// Highlight the match in the line
			highlighted := highlightMatch(m.Line, oldPath, newPath)
			fmt.Printf("    %4d: %s\n", m.LineNum, highlighted)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("---")
	fmt.Printf("Summary: %d matches in %d files (%d go.mod, %d .go)\n",
		totalMatches, len(matches), goModCount, goFileCount)
	fmt.Printf("Will replace: '%s' -> '%s'\n", oldPath, newPath)
}

// PrintResults displays the results after replacement
func PrintResults(results []ReplaceResult) {
	successCount := 0
	errorCount := 0
	totalReplaced := 0

	for _, r := range results {
		if r.Err != nil {
			errorCount++
			fmt.Printf("  ERROR: %s: %v\n", r.Path, r.Err)
		} else if r.ReplacedCount > 0 {
			successCount++
			totalReplaced += r.ReplacedCount
			fmt.Printf("  Updated: %s (%d replacements)\n", r.Path, r.ReplacedCount)
		}
	}

	fmt.Println()
	fmt.Printf("Done: %d files updated, %d total replacements", successCount, totalReplaced)
	if errorCount > 0 {
		fmt.Printf(", %d errors", errorCount)
	}
	fmt.Println()
}

// highlightMatch shows the transformation that will occur
func highlightMatch(line, oldPath, newPath string) string {
	// For terminal output, show the line with old path
	// In a more sophisticated version, we could use ANSI colors
	return strings.TrimSpace(line)
}
