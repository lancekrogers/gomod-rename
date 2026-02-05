package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Run_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), "module github.com/old/repo\ngo 1.21\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("")

	c := New(stdout, stderr, stdin, "test")
	exitCode := c.Run(context.Background(), []string{"-d", tmpDir, "github.com/old/repo", "github.com/new/repo"})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "github.com/old/repo") {
		t.Error("output should contain old path")
	}
	if !strings.Contains(output, "dry-run") {
		t.Error("output should mention dry-run")
	}

	// File should NOT be modified
	content, _ := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if !strings.Contains(string(content), "github.com/old/repo") {
		t.Error("file should not be modified in dry-run")
	}
}

func TestCLI_Run_Write(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), "module github.com/old/repo\ngo 1.21\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("y\n")

	c := New(stdout, stderr, stdin, "test")
	exitCode := c.Run(context.Background(), []string{"-d", tmpDir, "-w", "github.com/old/repo", "github.com/new/repo"})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0", exitCode)
	}

	// File should be modified
	content, _ := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if !strings.Contains(string(content), "github.com/new/repo") {
		t.Error("file should be modified after --write")
	}
}

func TestCLI_Run_WriteWithYes(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), "module github.com/old/repo\ngo 1.21\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("") // No input needed with -y

	c := New(stdout, stderr, stdin, "test")
	exitCode := c.Run(context.Background(), []string{"-d", tmpDir, "-w", "-y", "github.com/old/repo", "github.com/new/repo"})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0", exitCode)
	}

	// File should be modified
	content, _ := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if !strings.Contains(string(content), "github.com/new/repo") {
		t.Error("file should be modified with -w -y")
	}
}

func TestCLI_Run_Abort(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), "module github.com/old/repo\ngo 1.21\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("n\n") // User says no

	c := New(stdout, stderr, stdin, "test")
	exitCode := c.Run(context.Background(), []string{"-d", tmpDir, "-w", "github.com/old/repo", "github.com/new/repo"})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0 (abort is not an error)", exitCode)
	}

	if !strings.Contains(stdout.String(), "Aborted") {
		t.Error("output should mention abort")
	}

	// File should NOT be modified
	content, _ := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if !strings.Contains(string(content), "github.com/old/repo") {
		t.Error("file should not be modified after abort")
	}
}

func TestCLI_Run_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "go.mod"), "module github.com/other/repo\ngo 1.21\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("")

	c := New(stdout, stderr, stdin, "test")
	exitCode := c.Run(context.Background(), []string{"-d", tmpDir, "github.com/old/repo", "github.com/new/repo"})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0", exitCode)
	}

	if !strings.Contains(stdout.String(), "No files") {
		t.Error("output should mention no files found")
	}
}

func TestCLI_Run_Version(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("")

	c := New(stdout, stderr, stdin, "v1.2.3")
	exitCode := c.Run(context.Background(), []string{"--version"})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0", exitCode)
	}

	if !strings.Contains(stdout.String(), "v1.2.3") {
		t.Errorf("output = %q, should contain version", stdout.String())
	}
}

func TestCLI_Run_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{}},
		{"one argument", []string{"github.com/old/repo"}},
		{"three arguments", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			stdin := strings.NewReader("")

			c := New(stdout, stderr, stdin, "test")
			exitCode := c.Run(context.Background(), tt.args)

			if exitCode != 1 {
				t.Errorf("Run() exit code = %d, want 1", exitCode)
			}
		})
	}
}

func TestCLI_Run_SamePaths(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("")

	c := New(stdout, stderr, stdin, "test")
	exitCode := c.Run(context.Background(), []string{"github.com/same/repo", "github.com/same/repo"})

	if exitCode != 1 {
		t.Errorf("Run() exit code = %d, want 1", exitCode)
	}

	if !strings.Contains(stderr.String(), "identical") {
		t.Error("stderr should mention paths are identical")
	}
}

func TestCLI_Run_InvalidDirectory(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("")

	c := New(stdout, stderr, stdin, "test")
	exitCode := c.Run(context.Background(), []string{"-d", "/nonexistent/path", "old", "new"})

	if exitCode != 1 {
		t.Errorf("Run() exit code = %d, want 1", exitCode)
	}
}

func TestCLI_Run_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a file with many matches
	content := strings.Repeat("github.com/old/repo\n", 10)
	writeTestFile(t, filepath.Join(tmpDir, "main.go"), "package main\n"+content)

	// Without verbose - should truncate
	stdout1 := &bytes.Buffer{}
	c1 := New(stdout1, &bytes.Buffer{}, strings.NewReader(""), "test")
	c1.Run(context.Background(), []string{"-d", tmpDir, "github.com/old/repo", "github.com/new/repo"})

	if !strings.Contains(stdout1.String(), "more matches") {
		t.Error("non-verbose output should truncate matches")
	}

	// With verbose - should show all
	stdout2 := &bytes.Buffer{}
	c2 := New(stdout2, &bytes.Buffer{}, strings.NewReader(""), "test")
	c2.Run(context.Background(), []string{"-d", tmpDir, "-v", "github.com/old/repo", "github.com/new/repo"})

	if strings.Contains(stdout2.String(), "more matches") {
		t.Error("verbose output should show all matches")
	}
}

func TestCLI_Confirm(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
		{"anything\n", false},
		{"\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c := New(&bytes.Buffer{}, &bytes.Buffer{}, strings.NewReader(tt.input), "test")
			got := c.confirm()

			if got != tt.want {
				t.Errorf("confirm() with input %q = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
