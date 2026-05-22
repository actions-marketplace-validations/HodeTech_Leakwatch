package filter

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/cemililik/leakwatch/pkg/finding"
)

const (
	// inlineIgnoreTag is the marker that disables leak detection for a line.
	inlineIgnoreTag = "leakwatch:ignore"
)

// HasInlineIgnore reports whether line contains the generic inline ignore
// marker "leakwatch:ignore". The marker may appear anywhere in the line
// (typically inside a comment).
func HasInlineIgnore(line string) bool {
	return strings.Contains(line, inlineIgnoreTag)
}

// HasInlineIgnoreForDetector reports whether line contains the detector-
// specific inline ignore marker "leakwatch:ignore:<detectorID>".
// It also returns true when the generic "leakwatch:ignore" marker (without a
// detector suffix) is present.
func HasInlineIgnoreForDetector(line string, detectorID string) bool {
	// Check for detector-specific marker first.
	specific := inlineIgnoreTag + ":" + detectorID
	if strings.Contains(line, specific) {
		return true
	}

	// A bare "leakwatch:ignore" (not followed by ':') covers all detectors.
	idx := strings.Index(line, inlineIgnoreTag)
	if idx == -1 {
		return false
	}
	afterTag := idx + len(inlineIgnoreTag)
	if afterTag >= len(line) {
		// Tag is at the end of the line — generic ignore.
		return true
	}
	// If the character right after the tag is not ':', it is a generic ignore.
	return line[afterTag] != ':'
}

// LineHasInlineIgnore reports whether the 1-based lineNum in data carries an
// inline ignore marker (generic or detector-specific) for detectorID.
// It returns false when lineNum is out of range or non-positive, which lets
// callers use it as a single guard regardless of whether line tracking is
// available for a given source.
func LineHasInlineIgnore(data []byte, lineNum int, detectorID string) bool {
	if lineNum <= 0 {
		return false
	}
	line := getLine(data, lineNum)
	if line == "" {
		return false
	}
	return HasInlineIgnoreForDetector(line, detectorID)
}

// FilterFindingsByInlineIgnore returns a filtered slice of findings, removing
// any finding whose source line contains an inline ignore marker.
// sourceData maps file paths to their raw content; findings whose file is
// missing from the map are kept as-is.
func FilterFindingsByInlineIgnore(findings []finding.Finding, sourceData map[string][]byte) []finding.Finding {
	var kept []finding.Finding
	for _, f := range findings {
		data, ok := sourceData[f.SourceMetadata.FilePath]
		if !ok {
			kept = append(kept, f)
			continue
		}
		if LineHasInlineIgnore(data, f.SourceMetadata.Line, f.DetectorID) {
			continue
		}
		kept = append(kept, f)
	}
	return kept
}

// getLine returns the content of the 1-based line number from data.
// If the line number is out of range, an empty string is returned.
func getLine(data []byte, lineNum int) string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	current := 0
	for scanner.Scan() {
		current++
		if current == lineNum {
			return scanner.Text()
		}
	}
	return ""
}
