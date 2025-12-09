package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Match represents a single occurrence of the old path in a file
type Match struct {
	LineNum int
	Line    string
}

// FileMatch represents a file containing matches
type FileMatch struct {
	Path    string
	IsGoMod bool
	Matches []Match
}

// FindFiles walks the directory tree and finds all files containing the old path
func FindFiles(root, oldPath string) ([]FileMatch, error) {
	var results []FileMatch

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories (except current dir)
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			// Skip vendor directories
			if name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process go.mod and .go files
		name := info.Name()
		isGoMod := name == "go.mod"
		isGoFile := strings.HasSuffix(name, ".go")

		if !isGoMod && !isGoFile {
			return nil
		}

		// Check if file contains the old path
		matches, err := findMatchesInFile(path, oldPath)
		if err != nil {
			return err
		}

		if len(matches) > 0 {
			results = append(results, FileMatch{
				Path:    path,
				IsGoMod: isGoMod,
				Matches: matches,
			})
		}

		return nil
	})

	return results, err
}

// findMatchesInFile scans a file for occurrences of the search string
func findMatchesInFile(path, search string) ([]Match, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []Match
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(line, search) {
			matches = append(matches, Match{
				LineNum: lineNum,
				Line:    line,
			})
		}
	}

	return matches, scanner.Err()
}
