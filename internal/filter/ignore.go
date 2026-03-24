// Package filter provides file filtering helpers.
package filter

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ignorePattern represents a single parsed pattern from a .leakwatchignore file.
type ignorePattern struct {
	// pattern is the glob pattern string.
	pattern string
	// negated indicates the pattern was prefixed with '!' (un-ignore).
	negated bool
}

// IgnoreRules holds the parsed patterns from a .leakwatchignore file.
type IgnoreRules struct {
	patterns []ignorePattern
}

// LoadIgnoreFile reads and parses a .leakwatchignore file at the given path.
// Lines starting with '#' are treated as comments and blank lines are skipped.
// A '!' prefix negates the pattern (un-ignores a previously ignored path).
func LoadIgnoreFile(path string) (*IgnoreRules, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening ignore file %q: %w", path, err)
	}
	defer f.Close()

	rules := &IgnoreRules{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		negated := false
		if strings.HasPrefix(line, "!") {
			negated = true
			line = strings.TrimPrefix(line, "!")
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
		}

		rules.patterns = append(rules.patterns, ignorePattern{
			pattern: line,
			negated: negated,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading ignore file %q: %w", path, err)
	}

	return rules, nil
}

// ShouldIgnore reports whether path matches the ignore rules.
// Patterns are evaluated in order; the last matching pattern wins.
// A negated pattern (!) re-includes a previously ignored path.
func (r *IgnoreRules) ShouldIgnore(path string) bool {
	ignored := false
	for _, p := range r.patterns {
		if MatchesGlob(path, []string{p.pattern}) {
			ignored = !p.negated
		}
	}
	return ignored
}
