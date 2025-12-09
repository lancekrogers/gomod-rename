package main

import (
	"os"
	"strings"
)

// ReplaceResult holds the result of a replacement operation
type ReplaceResult struct {
	Path         string
	ReplacedCount int
	Err          error
}

// ReplaceInFiles performs the actual replacement in all matched files
func ReplaceInFiles(matches []FileMatch, oldPath, newPath string) []ReplaceResult {
	results := make([]ReplaceResult, 0, len(matches))

	for _, fm := range matches {
		result := replaceInFile(fm.Path, oldPath, newPath)
		results = append(results, result)
	}

	return results
}

// replaceInFile replaces all occurrences of oldPath with newPath in a single file
func replaceInFile(path, oldPath, newPath string) ReplaceResult {
	// Read entire file
	content, err := os.ReadFile(path)
	if err != nil {
		return ReplaceResult{Path: path, Err: err}
	}

	original := string(content)
	replaced := strings.ReplaceAll(original, oldPath, newPath)

	// Count replacements
	count := strings.Count(original, oldPath)

	if count == 0 {
		return ReplaceResult{Path: path, ReplacedCount: 0}
	}

	// Get original file permissions
	info, err := os.Stat(path)
	if err != nil {
		return ReplaceResult{Path: path, Err: err}
	}

	// Write back with same permissions
	err = os.WriteFile(path, []byte(replaced), info.Mode())
	if err != nil {
		return ReplaceResult{Path: path, Err: err}
	}

	return ReplaceResult{Path: path, ReplacedCount: count}
}
