// Package finder provides functionality to discover Go files containing specific import paths.
package finder

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Match represents a single occurrence of a pattern in a file.
type Match struct {
	LineNum int
	Line    string
}

// FileMatch represents a file containing one or more matches.
type FileMatch struct {
	Path    string
	IsGoMod bool
	Matches []Match
}

// Config holds configuration options for the finder.
type Config struct {
	// SkipVendor controls whether vendor directories are skipped (default: true)
	SkipVendor bool
	// SkipHidden controls whether hidden directories are skipped (default: true)
	SkipHidden bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		SkipVendor: true,
		SkipHidden: true,
	}
}

// Find walks the directory tree starting at root and returns all files
// containing the specified pattern. It only searches go.mod and .go files.
func Find(root, pattern string, cfg *Config) ([]FileMatch, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var results []FileMatch

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access rather than failing entirely
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		if info.IsDir() {
			return handleDirectory(info.Name(), cfg)
		}

		match, err := processFile(path, info.Name(), pattern)
		if err != nil {
			return err
		}

		if match != nil {
			results = append(results, *match)
		}

		return nil
	})

	return results, err
}

// handleDirectory determines whether to skip a directory based on config.
func handleDirectory(name string, cfg *Config) error {
	// Skip hidden directories (except current dir)
	if cfg.SkipHidden && strings.HasPrefix(name, ".") && name != "." {
		return filepath.SkipDir
	}

	// Skip vendor directories
	if cfg.SkipVendor && name == "vendor" {
		return filepath.SkipDir
	}

	return nil
}

// processFile checks if a file should be searched and returns matches if found.
func processFile(path, name, pattern string) (*FileMatch, error) {
	isGoMod := name == "go.mod"
	isGoFile := strings.HasSuffix(name, ".go")

	if !isGoMod && !isGoFile {
		return nil, nil
	}

	matches, err := findMatchesInFile(path, pattern)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	return &FileMatch{
		Path:    path,
		IsGoMod: isGoMod,
		Matches: matches,
	}, nil
}

// findMatchesInFile scans a file for occurrences of the pattern.
func findMatchesInFile(path, pattern string) ([]Match, error) {
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
		if strings.Contains(line, pattern) {
			matches = append(matches, Match{
				LineNum: lineNum,
				Line:    line,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}
