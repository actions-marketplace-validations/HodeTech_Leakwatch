package filter

import (
	"os"
	"path/filepath"
	"testing"
)

func writeIgnoreFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".leakwatchignore")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing ignore file: %v", err)
	}
	return path
}

func TestLoadIgnoreFile_CommentLines_Skipped(t *testing.T) {
	path := writeIgnoreFile(t, "# this is a comment\n*.go\n")
	rules, err := LoadIgnoreFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules.patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(rules.patterns))
	}
	if rules.patterns[0].pattern != "*.go" {
		t.Errorf("expected pattern *.go, got %q", rules.patterns[0].pattern)
	}
}

func TestLoadIgnoreFile_EmptyLines_Skipped(t *testing.T) {
	path := writeIgnoreFile(t, "\n\n*.md\n\n")
	rules, err := LoadIgnoreFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules.patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(rules.patterns))
	}
}

func TestShouldIgnore_GlobPattern_Matches(t *testing.T) {
	path := writeIgnoreFile(t, "*.md\n*.log\n")
	rules, err := LoadIgnoreFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		filePath string
		want     bool
	}{
		{"README.md", true},
		{"docs/guide.md", true},
		{"main.go", false},
		{"app.log", true},
	}
	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := rules.ShouldIgnore(tt.filePath)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestShouldIgnore_DoubleStarPattern_Matches(t *testing.T) {
	path := writeIgnoreFile(t, "vendor/**\ntest/**\n")
	rules, err := LoadIgnoreFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		filePath string
		want     bool
	}{
		{"vendor/github.com/pkg/errors/errors.go", true},
		{"vendor/module/sub/deep.go", true},
		{"test/integration/scan_test.go", true},
		{"src/main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := rules.ShouldIgnore(tt.filePath)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestShouldIgnore_NegationPattern_ReIncludes(t *testing.T) {
	path := writeIgnoreFile(t, "*.log\n!important.log\n")
	rules, err := LoadIgnoreFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		filePath string
		want     bool
	}{
		{"debug.log", true},
		{"important.log", false},
		{"main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := rules.ShouldIgnore(tt.filePath)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestLoadIgnoreFile_FileNotFound_ReturnsError(t *testing.T) {
	_, err := LoadIgnoreFile("/nonexistent/.leakwatchignore")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}
