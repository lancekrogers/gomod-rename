package finder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFind(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		setup         func(t *testing.T, root string)
		pattern       string
		cfg           *Config
		wantFiles     int
		wantMatches   int
		wantGoMod     int
		wantGoFiles   int
	}{
		{
			name: "finds matches in go.mod",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "go.mod"), "module github.com/old/repo\ngo 1.21\n")
			},
			pattern:     "github.com/old/repo",
			wantFiles:   1,
			wantMatches: 1,
			wantGoMod:   1,
			wantGoFiles: 0,
		},
		{
			name: "finds matches in .go files",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "main.go"), `package main

import "github.com/old/repo/pkg"

func main() {}
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   1,
			wantMatches: 1,
			wantGoMod:   0,
			wantGoFiles: 1,
		},
		{
			name: "finds multiple matches in single file",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "main.go"), `package main

import (
	"github.com/old/repo/pkg1"
	"github.com/old/repo/pkg2"
	"github.com/old/repo/pkg3"
)

func main() {}
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   1,
			wantMatches: 3,
			wantGoMod:   0,
			wantGoFiles: 1,
		},
		{
			name: "finds matches across multiple files",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "go.mod"), "module github.com/old/repo\ngo 1.21\n")
				writeFile(t, filepath.Join(root, "main.go"), `package main
import "github.com/old/repo/pkg"
`)
				writeFile(t, filepath.Join(root, "other.go"), `package main
import "github.com/old/repo/other"
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   3,
			wantMatches: 3,
			wantGoMod:   1,
			wantGoFiles: 2,
		},
		{
			name: "skips vendor directory by default",
			setup: func(t *testing.T, root string) {
				vendorDir := filepath.Join(root, "vendor")
				mustMkdir(t, vendorDir)
				writeFile(t, filepath.Join(vendorDir, "dep.go"), `package dep
import "github.com/old/repo/internal"
`)
				writeFile(t, filepath.Join(root, "main.go"), `package main
import "github.com/old/repo/pkg"
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   1,
			wantMatches: 1,
			wantGoMod:   0,
			wantGoFiles: 1,
		},
		{
			name: "includes vendor when configured",
			setup: func(t *testing.T, root string) {
				vendorDir := filepath.Join(root, "vendor")
				mustMkdir(t, vendorDir)
				writeFile(t, filepath.Join(vendorDir, "dep.go"), `package dep
import "github.com/old/repo/internal"
`)
				writeFile(t, filepath.Join(root, "main.go"), `package main
import "github.com/old/repo/pkg"
`)
			},
			pattern:     "github.com/old/repo",
			cfg:         &Config{SkipVendor: false, SkipHidden: true},
			wantFiles:   2,
			wantMatches: 2,
			wantGoMod:   0,
			wantGoFiles: 2,
		},
		{
			name: "skips hidden directories by default",
			setup: func(t *testing.T, root string) {
				hiddenDir := filepath.Join(root, ".hidden")
				mustMkdir(t, hiddenDir)
				writeFile(t, filepath.Join(hiddenDir, "secret.go"), `package secret
import "github.com/old/repo/internal"
`)
				writeFile(t, filepath.Join(root, "main.go"), `package main
import "github.com/old/repo/pkg"
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   1,
			wantMatches: 1,
			wantGoMod:   0,
			wantGoFiles: 1,
		},
		{
			name: "includes hidden when configured",
			setup: func(t *testing.T, root string) {
				hiddenDir := filepath.Join(root, ".hidden")
				mustMkdir(t, hiddenDir)
				writeFile(t, filepath.Join(hiddenDir, "secret.go"), `package secret
import "github.com/old/repo/internal"
`)
				writeFile(t, filepath.Join(root, "main.go"), `package main
import "github.com/old/repo/pkg"
`)
			},
			pattern:     "github.com/old/repo",
			cfg:         &Config{SkipVendor: true, SkipHidden: false},
			wantFiles:   2,
			wantMatches: 2,
			wantGoMod:   0,
			wantGoFiles: 2,
		},
		{
			name: "returns empty for no matches",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "go.mod"), "module github.com/other/repo\ngo 1.21\n")
				writeFile(t, filepath.Join(root, "main.go"), `package main
func main() {}
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   0,
			wantMatches: 0,
			wantGoMod:   0,
			wantGoFiles: 0,
		},
		{
			name: "returns empty for empty directory",
			setup: func(t *testing.T, root string) {
				// No files created
			},
			pattern:     "github.com/old/repo",
			wantFiles:   0,
			wantMatches: 0,
			wantGoMod:   0,
			wantGoFiles: 0,
		},
		{
			name: "ignores non-go files",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "readme.md"), "github.com/old/repo")
				writeFile(t, filepath.Join(root, "config.yaml"), "repo: github.com/old/repo")
				writeFile(t, filepath.Join(root, "main.go"), `package main
import "github.com/old/repo/pkg"
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   1,
			wantMatches: 1,
			wantGoMod:   0,
			wantGoFiles: 1,
		},
		{
			name: "finds matches in nested directories",
			setup: func(t *testing.T, root string) {
				subDir := filepath.Join(root, "pkg", "sub")
				mustMkdir(t, subDir)
				writeFile(t, filepath.Join(root, "go.mod"), "module github.com/old/repo\n")
				writeFile(t, filepath.Join(root, "main.go"), `package main
import "github.com/old/repo/pkg"
`)
				writeFile(t, filepath.Join(subDir, "sub.go"), `package sub
import "github.com/old/repo/internal"
`)
			},
			pattern:     "github.com/old/repo",
			wantFiles:   3,
			wantMatches: 3,
			wantGoMod:   1,
			wantGoFiles: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh subdirectory for each test
			testDir := filepath.Join(tmpDir, tt.name)
			mustMkdir(t, testDir)

			// Setup test files
			tt.setup(t, testDir)

			// Run finder
			results, err := Find(testDir, tt.pattern, tt.cfg)
			if err != nil {
				t.Fatalf("Find() error = %v", err)
			}

			// Count results
			gotFiles := len(results)
			gotMatches := 0
			gotGoMod := 0
			gotGoFiles := 0

			for _, r := range results {
				gotMatches += len(r.Matches)
				if r.IsGoMod {
					gotGoMod++
				} else {
					gotGoFiles++
				}
			}

			if gotFiles != tt.wantFiles {
				t.Errorf("Find() got %d files, want %d", gotFiles, tt.wantFiles)
			}
			if gotMatches != tt.wantMatches {
				t.Errorf("Find() got %d matches, want %d", gotMatches, tt.wantMatches)
			}
			if gotGoMod != tt.wantGoMod {
				t.Errorf("Find() got %d go.mod files, want %d", gotGoMod, tt.wantGoMod)
			}
			if gotGoFiles != tt.wantGoFiles {
				t.Errorf("Find() got %d .go files, want %d", gotGoFiles, tt.wantGoFiles)
			}
		})
	}
}

func TestFindMatchLineNumbers(t *testing.T) {
	tmpDir := t.TempDir()

	content := `package main

import (
	"fmt"
	"github.com/old/repo/pkg"
)

func main() {
	fmt.Println("github.com/old/repo")
}
`
	writeFile(t, filepath.Join(tmpDir, "main.go"), content)

	results, err := Find(tmpDir, "github.com/old/repo", nil)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 file, got %d", len(results))
	}

	if len(results[0].Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results[0].Matches))
	}

	// First match should be on line 5 (import)
	if results[0].Matches[0].LineNum != 5 {
		t.Errorf("first match line = %d, want 5", results[0].Matches[0].LineNum)
	}

	// Second match should be on line 9 (string literal)
	if results[0].Matches[1].LineNum != 9 {
		t.Errorf("second match line = %d, want 9", results[0].Matches[1].LineNum)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.SkipVendor {
		t.Error("DefaultConfig().SkipVendor should be true")
	}

	if !cfg.SkipHidden {
		t.Error("DefaultConfig().SkipHidden should be true")
	}
}

// Helper functions

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}
