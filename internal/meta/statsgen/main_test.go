package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestManagedAssetsUpToDate fails when a marketing asset's stat block no longer
// matches internal/meta — i.e. a count was bumped but `go generate ./...` was
// not run (or the PNG was re-rendered from a stale source). It runs as part of
// `go test ./...`, so CI catches the drift without a dedicated workflow step.
func TestManagedAssetsUpToDate(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	for _, rel := range managedFiles {
		orig, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		updated, err := rewrite(string(orig))
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		if updated != string(orig) {
			t.Errorf("%s stat block is stale; run `go generate ./...` and re-render its PNG", rel)
		}
	}
}
