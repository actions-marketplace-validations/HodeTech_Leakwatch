package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemSource_Type(t *testing.T) {
	s := New("/tmp")
	assert.Equal(t, "filesystem", s.Type())
}

func TestFilesystemSource_New_CleansPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantAbs  bool
		wantClean bool
	}{
		{
			name:      "trailing slash removed",
			input:     "/tmp/foo/",
			wantClean: true,
		},
		{
			name:      "double slash cleaned",
			input:     "/tmp//foo",
			wantClean: true,
		},
		{
			name:      "dot segments resolved",
			input:     "/tmp/foo/../bar",
			wantClean: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.input)
			cleaned := filepath.Clean(tt.input)
			abs, err := filepath.Abs(cleaned)
			if err != nil {
				abs = cleaned
			}
			assert.Equal(t, abs, s.root, "root should be cleaned and absolute")
		})
	}
}

func TestFilesystemSource_Validate_ValidDir(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	assert.NoError(t, s.Validate())
}

func TestFilesystemSource_Validate_NonExistentDir(t *testing.T) {
	s := New("/nonexistent/path")
	assert.Error(t, s.Validate())
}

func TestFilesystemSource_Validate_FileNotDir(t *testing.T) {
	f, err := os.CreateTemp("", "test")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	s := New(f.Name())
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestFilesystemSource_Chunks_ReadsFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("AKIAIOSFODNN7EXAMPLE"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("api_key: test123"), 0644))

	s := New(dir)
	ctx := context.Background()

	var chunks []string
	for chunk := range s.Chunks(ctx) {
		chunks = append(chunks, chunk.SourceMetadata.FilePath)
	}

	assert.Len(t, chunks, 2)
}

func TestFilesystemSource_Chunks_SkipsBinaryFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "text.txt"), []byte("hello world"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "binary.dat"), []byte("hello\x00world"), 0644))

	s := New(dir)
	ctx := context.Background()

	var chunks []string
	for chunk := range s.Chunks(ctx) {
		chunks = append(chunks, chunk.SourceMetadata.FilePath)
	}

	assert.Len(t, chunks, 1)
	assert.Equal(t, "text.txt", chunks[0])
}

func TestFilesystemSource_Chunks_SkipsBinaryExtensions(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "code.go"), []byte("package main"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "image.png"), []byte("fakepng"), 0644))

	s := New(dir)
	ctx := context.Background()

	var chunks []string
	for chunk := range s.Chunks(ctx) {
		chunks = append(chunks, chunk.SourceMetadata.FilePath)
	}

	assert.Len(t, chunks, 1)
	assert.Equal(t, "code.go", chunks[0])
}

func TestFilesystemSource_Chunks_RespectsMaxFileSize(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "small.txt"), []byte("small"), 0644))

	bigData := make([]byte, 1024)
	for i := range bigData {
		bigData[i] = 'A'
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "big.txt"), bigData, 0644))

	s := New(dir, WithMaxFileSize(512))
	ctx := context.Background()

	var chunks []string
	for chunk := range s.Chunks(ctx) {
		chunks = append(chunks, chunk.SourceMetadata.FilePath)
	}

	assert.Len(t, chunks, 1)
	assert.Equal(t, "small.txt", chunks[0])
}

func TestFilesystemSource_Chunks_ContextCancellation(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 100; i++ {
		name := filepath.Join(dir, "file"+string(rune('a'+i%26))+".txt")
		_ = os.WriteFile(name, []byte("content"), 0644)
	}

	s := New(dir, WithBufferSize(1))
	ctx, cancel := context.WithCancel(context.Background())

	ch := s.Chunks(ctx)

	// Get the first chunk, then cancel.
	<-ch
	cancel()

	// Verify the channel eventually closes.
	count := 0
	for range ch {
		count++
	}
	// Buffered chunks may arrive but the channel must close.
	assert.Less(t, count, 100)
}

func TestFilesystemSource_Chunks_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	s := New(dir)
	ctx := context.Background()

	var count int
	for range s.Chunks(ctx) {
		count++
	}

	assert.Equal(t, 0, count)
}

func TestFilesystemSource_Chunks_ExcludePaths(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.sum"), []byte("checksum"), 0644))

	s := New(dir, WithExcludePaths([]string{"go.sum"}))
	ctx := context.Background()

	var chunks []string
	for chunk := range s.Chunks(ctx) {
		chunks = append(chunks, chunk.SourceMetadata.FilePath)
	}

	assert.Len(t, chunks, 1)
	assert.Equal(t, "main.go", chunks[0])
}

func TestFilesystemSource_Chunks_SkipsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated privileges on Windows")
	}

	dir := t.TempDir()

	// Create a real file.
	realFile := filepath.Join(dir, "real.txt")
	require.NoError(t, os.WriteFile(realFile, []byte("real content"), 0644))

	// Create a symlink pointing to the real file.
	symlinkFile := filepath.Join(dir, "link.txt")
	require.NoError(t, os.Symlink(realFile, symlinkFile))

	s := New(dir)
	ctx := context.Background()

	var chunks []string
	for chunk := range s.Chunks(ctx) {
		chunks = append(chunks, chunk.SourceMetadata.FilePath)
	}

	// Only the real file should be scanned; the symlink should be skipped.
	assert.Len(t, chunks, 1)
	assert.Equal(t, "real.txt", chunks[0])
}
