package replacer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lancekrogers/gomod-rename/internal/finder"
)

func TestReplaceInFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		oldStr      string
		newStr      string
		wantContent string
		wantCount   int
		wantErr     bool
	}{
		{
			name:        "single replacement",
			content:     "module github.com/old/repo\ngo 1.21\n",
			oldStr:      "github.com/old/repo",
			newStr:      "github.com/new/repo",
			wantContent: "module github.com/new/repo\ngo 1.21\n",
			wantCount:   1,
		},
		{
			name: "multiple replacements",
			content: `package main

import (
	"github.com/old/repo/pkg1"
	"github.com/old/repo/pkg2"
)
`,
			oldStr: "github.com/old/repo",
			newStr: "github.com/new/repo",
			wantContent: `package main

import (
	"github.com/new/repo/pkg1"
	"github.com/new/repo/pkg2"
)
`,
			wantCount: 2,
		},
		{
			name:        "no matches returns zero count",
			content:     "module github.com/other/repo\n",
			oldStr:      "github.com/old/repo",
			newStr:      "github.com/new/repo",
			wantContent: "module github.com/other/repo\n",
			wantCount:   0,
		},
		{
			name:        "preserves line endings",
			content:     "line1\nline2\nline3\n",
			oldStr:      "line2",
			newStr:      "replaced",
			wantContent: "line1\nreplaced\nline3\n",
			wantCount:   1,
		},
		{
			name:        "handles empty file",
			content:     "",
			oldStr:      "anything",
			newStr:      "nothing",
			wantContent: "",
			wantCount:   0,
		},
		{
			name:        "partial path match",
			content:     "github.com/old/repo/subpkg",
			oldStr:      "github.com/old/repo",
			newStr:      "github.com/new/repo",
			wantContent: "github.com/new/repo/subpkg",
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.go")

			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Run replacement
			result := ReplaceInFile(path, tt.oldStr, tt.newStr)

			// Check for unexpected errors
			if (result.Err != nil) != tt.wantErr {
				t.Errorf("ReplaceInFile() error = %v, wantErr %v", result.Err, tt.wantErr)
				return
			}

			// Check count
			if result.Count != tt.wantCount {
				t.Errorf("ReplaceInFile() count = %d, want %d", result.Count, tt.wantCount)
			}

			// Check file content
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read result file: %v", err)
			}

			if string(got) != tt.wantContent {
				t.Errorf("file content = %q, want %q", string(got), tt.wantContent)
			}
		})
	}
}

func TestReplaceInFilePreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.go")

	// Write file with specific permissions
	content := "github.com/old/repo"
	wantMode := os.FileMode(0600)

	if err := os.WriteFile(path, []byte(content), wantMode); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run replacement
	result := ReplaceInFile(path, "github.com/old/repo", "github.com/new/repo")
	if result.Err != nil {
		t.Fatalf("ReplaceInFile() error = %v", result.Err)
	}

	// Check permissions are preserved
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	gotMode := info.Mode().Perm()
	if gotMode != wantMode {
		t.Errorf("file mode = %o, want %o", gotMode, wantMode)
	}
}

func TestReplaceInFileHandlesReadOnlyError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "readonly.go")

	content := "github.com/old/repo"
	if err := os.WriteFile(path, []byte(content), 0444); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := ReplaceInFile(path, "github.com/old/repo", "github.com/new/repo")

	// Should return an error because file is read-only
	if result.Err == nil {
		t.Error("ReplaceInFile() expected error for read-only file, got nil")
	}
}

func TestReplaceInFileHandlesNonExistentFile(t *testing.T) {
	result := ReplaceInFile("/nonexistent/path/file.go", "old", "new")

	if result.Err == nil {
		t.Error("ReplaceInFile() expected error for nonexistent file, got nil")
	}
}

func TestReplace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []struct {
		name    string
		content string
	}{
		{"go.mod", "module github.com/old/repo\ngo 1.21\n"},
		{"main.go", `package main
import "github.com/old/repo/pkg"
`},
		{"other.go", `package main
import "github.com/other/repo"
`},
	}

	var matches []finder.FileMatch
	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", f.name, err)
		}

		// Only add files that contain the pattern
		if f.name != "other.go" {
			matches = append(matches, finder.FileMatch{
				Path:    path,
				IsGoMod: f.name == "go.mod",
			})
		}
	}

	results := Replace(matches, "github.com/old/repo", "github.com/new/repo")

	if len(results) != 2 {
		t.Fatalf("Replace() returned %d results, want 2", len(results))
	}

	// Check all replacements succeeded
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("Replace() error for %s: %v", r.Path, r.Err)
		}
		if r.Count != 1 {
			t.Errorf("Replace() count for %s = %d, want 1", r.Path, r.Count)
		}
	}

	// Verify file contents
	goMod, _ := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if want := "module github.com/new/repo\ngo 1.21\n"; string(goMod) != want {
		t.Errorf("go.mod = %q, want %q", string(goMod), want)
	}

	mainGo, _ := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	wantMain := `package main
import "github.com/new/repo/pkg"
`
	if string(mainGo) != wantMain {
		t.Errorf("main.go = %q, want %q", string(mainGo), wantMain)
	}

	// other.go should be unchanged
	otherGo, _ := os.ReadFile(filepath.Join(tmpDir, "other.go"))
	wantOther := `package main
import "github.com/other/repo"
`
	if string(otherGo) != wantOther {
		t.Errorf("other.go = %q, want %q", string(otherGo), wantOther)
	}
}

func TestSummarize(t *testing.T) {
	tests := []struct {
		name    string
		results []Result
		want    Summary
	}{
		{
			name:    "empty results",
			results: []Result{},
			want:    Summary{},
		},
		{
			name: "all successful",
			results: []Result{
				{Path: "a.go", Count: 2},
				{Path: "b.go", Count: 1},
				{Path: "c.go", Count: 3},
			},
			want: Summary{FilesModified: 3, TotalReplaced: 6, Errors: 0},
		},
		{
			name: "some errors",
			results: []Result{
				{Path: "a.go", Count: 2},
				{Path: "b.go", Err: os.ErrPermission},
				{Path: "c.go", Count: 1},
			},
			want: Summary{FilesModified: 2, TotalReplaced: 3, Errors: 1},
		},
		{
			name: "some zero counts",
			results: []Result{
				{Path: "a.go", Count: 2},
				{Path: "b.go", Count: 0},
				{Path: "c.go", Count: 1},
			},
			want: Summary{FilesModified: 2, TotalReplaced: 3, Errors: 0},
		},
		{
			name: "all errors",
			results: []Result{
				{Path: "a.go", Err: os.ErrPermission},
				{Path: "b.go", Err: os.ErrNotExist},
			},
			want: Summary{FilesModified: 0, TotalReplaced: 0, Errors: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Summarize(tt.results)

			if got.FilesModified != tt.want.FilesModified {
				t.Errorf("Summarize().FilesModified = %d, want %d", got.FilesModified, tt.want.FilesModified)
			}
			if got.TotalReplaced != tt.want.TotalReplaced {
				t.Errorf("Summarize().TotalReplaced = %d, want %d", got.TotalReplaced, tt.want.TotalReplaced)
			}
			if got.Errors != tt.want.Errors {
				t.Errorf("Summarize().Errors = %d, want %d", got.Errors, tt.want.Errors)
			}
		})
	}
}
