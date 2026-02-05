//go:build integration

package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	oldModule = "github.com/testold/myproject"
	newModule = "github.com/testnew/myproject"
)

func TestIntegration_RealProject(t *testing.T) {
	// Find the testdata/integration fixture
	fixtureDir := findFixtureDir(t)

	// Copy fixture to temp directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	copyDir(t, fixtureDir, projectDir)

	// Verify project builds before changes
	runGoBuild(t, projectDir, "project should build before replacement")

	// Run gomod-rename
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := New(stdout, stderr, strings.NewReader(""), "test")

	exitCode := c.Run(context.Background(), []string{"-d", projectDir, "-w", "-y", oldModule, newModule})
	if exitCode != 0 {
		t.Fatalf("gomod-rename failed with exit code %d\nstdout: %s\nstderr: %s",
			exitCode, stdout.String(), stderr.String())
	}

	// Verify project still builds after changes
	runGoBuild(t, projectDir, "project should build after replacement")

	// Verify go.mod was updated
	goModPath := filepath.Join(projectDir, "go.mod")
	goModContent := readFile(t, goModPath)
	if strings.Contains(goModContent, oldModule) {
		t.Errorf("go.mod still contains old module path: %s", oldModule)
	}
	if !strings.Contains(goModContent, newModule) {
		t.Errorf("go.mod does not contain new module path: %s", newModule)
	}

	// Verify main.go imports were updated
	mainPath := filepath.Join(projectDir, "main.go")
	mainContent := readFile(t, mainPath)
	if strings.Contains(mainContent, oldModule) {
		t.Errorf("main.go still contains old import path")
	}
	if !strings.Contains(mainContent, newModule) {
		t.Errorf("main.go does not contain new import path")
	}

	// Verify pkg/greeter/greeter.go was updated
	greeterPath := filepath.Join(projectDir, "pkg", "greeter", "greeter.go")
	greeterContent := readFile(t, greeterPath)
	if strings.Contains(greeterContent, oldModule) {
		t.Errorf("greeter.go still contains old import path")
	}
	if !strings.Contains(greeterContent, newModule) {
		t.Errorf("greeter.go does not contain new import path")
	}

	// Verify internal/helper/helper.go was NOT modified (no imports to change)
	helperPath := filepath.Join(projectDir, "internal", "helper", "helper.go")
	helperContent := readFile(t, helperPath)
	if strings.Contains(helperContent, oldModule) || strings.Contains(helperContent, newModule) {
		t.Errorf("helper.go should not contain module paths")
	}
}

func TestIntegration_DryRunDoesNotModify(t *testing.T) {
	fixtureDir := findFixtureDir(t)
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	copyDir(t, fixtureDir, projectDir)

	// Get original content
	goModPath := filepath.Join(projectDir, "go.mod")
	originalContent := readFile(t, goModPath)

	// Run in dry-run mode (no -w flag)
	stdout := &bytes.Buffer{}
	c := New(stdout, &bytes.Buffer{}, strings.NewReader(""), "test")
	exitCode := c.Run(context.Background(), []string{"-d", projectDir, oldModule, newModule})

	if exitCode != 0 {
		t.Fatalf("dry-run failed with exit code %d", exitCode)
	}

	// Verify file was NOT modified
	afterContent := readFile(t, goModPath)
	if afterContent != originalContent {
		t.Errorf("dry-run modified go.mod file")
	}
}

func TestIntegration_PartialPathMatch(t *testing.T) {
	fixtureDir := findFixtureDir(t)
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	copyDir(t, fixtureDir, projectDir)

	// Replace only the org name, not the full module path
	stdout := &bytes.Buffer{}
	c := New(stdout, &bytes.Buffer{}, strings.NewReader(""), "test")
	exitCode := c.Run(context.Background(), []string{"-d", projectDir, "-w", "-y", "github.com/testold", "github.com/testnew"})

	if exitCode != 0 {
		t.Fatalf("gomod-rename failed with exit code %d", exitCode)
	}

	// Should still build
	runGoBuild(t, projectDir, "project should build after org rename")

	// Verify the change
	goModContent := readFile(t, filepath.Join(projectDir, "go.mod"))
	if !strings.Contains(goModContent, "github.com/testnew/myproject") {
		t.Errorf("org rename did not work correctly")
	}
}

// Helper functions

func findFixtureDir(t *testing.T) string {
	t.Helper()

	// Try relative paths from different locations
	candidates := []string{
		"../../testdata/integration",
		"testdata/integration",
		"../testdata/integration",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}

	// Try finding from module root
	wd, _ := os.Getwd()
	for wd != "/" {
		candidate := filepath.Join(wd, "testdata", "integration")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		wd = filepath.Dir(wd)
	}

	t.Fatal("could not find testdata/integration fixture")
	return ""
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})

	if err != nil {
		t.Fatalf("failed to copy directory: %v", err)
	}
}

func runGoBuild(t *testing.T, dir, msg string) {
	t.Helper()

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: go build failed: %v\nOutput: %s", msg, err, output)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

// Ensure io import is used
var _ io.Reader
