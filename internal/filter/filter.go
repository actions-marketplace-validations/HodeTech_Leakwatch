// Package filter provides file filtering helpers.
package filter

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// binaryCheckLen is the number of bytes to inspect for null bytes.
	binaryCheckLen = 8192
)

// defaultBinaryExtensions lists file extensions that are always skipped.
var defaultBinaryExtensions = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".bin": true, ".o": true, ".a": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".bmp": true, ".ico": true, ".svg": true, ".webp": true,
	".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true,
	".rar": true, ".7z": true, ".xz": true,
	".pdf": true, ".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
}

// defaultSkipFilenames lists filenames that are always skipped.
// These are auto-generated files that contain hashes/checksums
// which frequently trigger false positives.
var defaultSkipFilenames = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"composer.lock":     true,
	"Gemfile.lock":      true,
	"Cargo.lock":        true,
	"poetry.lock":       true,
	"go.sum":            true,
	"Pipfile.lock":      true,
}

// IsSkippedFilename checks whether a filename should be skipped.
func IsSkippedFilename(path string) bool {
	return defaultSkipFilenames[filepath.Base(path)]
}

// IsExcludedExtension checks whether a file extension should be excluded.
func IsExcludedExtension(path string, extraExts []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if defaultBinaryExtensions[ext] {
		return true
	}
	for _, e := range extraExts {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

// IsBinaryFile checks whether data appears to be a binary file.
// If a null byte is found within the first 8KB, it is considered binary.
func IsBinaryFile(data []byte) bool {
	checkLen := binaryCheckLen
	if len(data) < checkLen {
		checkLen = len(data)
	}
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

// MatchesGlob checks whether a path matches any of the given glob patterns.
// Supports ** (double-star) patterns by splitting path segments.
// Returns an error if any pattern has invalid glob syntax.
func MatchesGlob(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchGlob(pattern, path) {
			return true
		}
		// Also match against the base filename for simple patterns.
		if matchGlob(pattern, filepath.Base(path)) {
			return true
		}
	}
	return false
}

// MatchesGlobStrict is like MatchesGlob but returns an error on invalid patterns.
func MatchesGlobStrict(path string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		matched, err := matchGlobStrict(pattern, path)
		if err != nil {
			return false, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
		matched, err = matchGlobStrict(pattern, filepath.Base(path))
		if err != nil {
			return false, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// matchGlob matches a single pattern against a path, supporting ** (double-star).
func matchGlob(pattern, path string) bool {
	matched, _ := matchGlobStrict(pattern, path)
	return matched
}

// matchGlobStrict matches a pattern with ** support, returning errors.
func matchGlobStrict(pattern, path string) (bool, error) {
	// If pattern contains **, use segment-based matching.
	if strings.Contains(pattern, "**") {
		return matchDoubleStar(pattern, path), nil
	}
	return filepath.Match(pattern, path)
}

// matchDoubleStar handles ** glob patterns.
// ** matches zero or more directory segments.
func matchDoubleStar(pattern, path string) bool {
	// Split both on separator
	patternParts := splitPath(pattern)
	pathParts := splitPath(path)
	return matchSegments(patternParts, pathParts)
}

func matchSegments(pattern, path []string) bool {
	// Base cases
	if len(pattern) == 0 {
		return len(path) == 0
	}

	head := pattern[0]
	rest := pattern[1:]

	if head == "**" {
		// ** matches zero or more segments
		// Try matching rest of pattern from every position in path
		for i := 0; i <= len(path); i++ {
			if matchSegments(rest, path[i:]) {
				return true
			}
		}
		return false
	}

	if len(path) == 0 {
		return false
	}

	matched, _ := filepath.Match(head, path[0])
	if !matched {
		return false
	}
	return matchSegments(rest, path[1:])
}

func splitPath(p string) []string {
	// Normalize separators
	p = filepath.ToSlash(p)
	parts := strings.Split(p, "/")
	// Remove empty parts
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
