// Package replacer provides functionality to replace strings in files.
package replacer

import (
	"os"
	"strings"

	"github.com/lancekrogers/gomod-rename/internal/finder"
)

// Result holds the outcome of a replacement operation on a single file.
type Result struct {
	Path    string
	Count   int
	Err     error
}

// Replace performs string replacement on all files in the given matches.
// It returns a Result for each file indicating success or failure.
func Replace(matches []finder.FileMatch, oldStr, newStr string) []Result {
	results := make([]Result, 0, len(matches))

	for _, fm := range matches {
		result := ReplaceInFile(fm.Path, oldStr, newStr)
		results = append(results, result)
	}

	return results
}

// ReplaceInFile replaces all occurrences of oldStr with newStr in the specified file.
// It preserves the original file permissions.
func ReplaceInFile(path, oldStr, newStr string) Result {
	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return Result{Path: path, Err: err}
	}

	original := string(content)

	// Count occurrences before replacement
	count := strings.Count(original, oldStr)
	if count == 0 {
		return Result{Path: path, Count: 0}
	}

	// Perform replacement
	replaced := strings.ReplaceAll(original, oldStr, newStr)

	// Get original file permissions
	info, err := os.Stat(path)
	if err != nil {
		return Result{Path: path, Err: err}
	}

	// Write back with same permissions
	err = os.WriteFile(path, []byte(replaced), info.Mode())
	if err != nil {
		return Result{Path: path, Err: err}
	}

	return Result{Path: path, Count: count}
}

// Summary holds aggregate statistics from a batch replacement operation.
type Summary struct {
	FilesModified  int
	TotalReplaced  int
	Errors         int
}

// Summarize calculates aggregate statistics from a slice of Results.
func Summarize(results []Result) Summary {
	var s Summary

	for _, r := range results {
		if r.Err != nil {
			s.Errors++
		} else if r.Count > 0 {
			s.FilesModified++
			s.TotalReplaced += r.Count
		}
	}

	return s
}
