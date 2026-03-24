// Package table provides a human-readable table output formatter for terminal display.
package table

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/cemililik/leakwatch/pkg/finding"
)

// Formatter outputs findings as a human-readable table for terminal display.
type Formatter struct {
	// ShowRaw, when true, includes the Raw field in output.
	// When false, the Raw field is stripped for defense in depth.
	ShowRaw bool
}

// Format writes findings as a formatted table to the given writer.
// Columns: SEVERITY | DETECTOR | FILE | REDACTED | STATUS | REMEDIATION
// A summary line is appended at the bottom.
// When ShowRaw is false, the Raw field is actively stripped from the output.
func (f *Formatter) Format(w io.Writer, findings []finding.Finding) error {
	output := make([]finding.Finding, len(findings))
	copy(output, findings)

	if !f.ShowRaw {
		for i := range output {
			output[i].Raw = ""
		}
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Write header.
	if _, err := fmt.Fprintln(tw, "SEVERITY\tDETECTOR\tFILE\tREDACTED\tSTATUS\tREMEDIATION"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}

	// Write separator.
	if _, err := fmt.Fprintln(tw, "--------\t--------\t----\t--------\t------\t-----------"); err != nil {
		return fmt.Errorf("failed to write table separator: %w", err)
	}

	// Write rows.
	for _, fd := range output {
		remediation := "-"
		if fd.Remediation != nil && fd.Remediation.Title != "" {
			remediation = fd.Remediation.Title
		}

		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
			strings.ToUpper(fd.Severity.String()),
			fd.DetectorID,
			fd.SourceMetadata.FilePath,
			fd.Redacted,
			fd.Verification.Status.String(),
			remediation,
		)
		if _, err := fmt.Fprintln(tw, line); err != nil {
			return fmt.Errorf("failed to write table row: %w", err)
		}
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("failed to flush table output: %w", err)
	}

	// Write summary line.
	summary := buildSummary(output)
	if _, err := fmt.Fprintln(w, ""); err != nil {
		return fmt.Errorf("failed to write table summary: %w", err)
	}
	if _, err := fmt.Fprintln(w, summary); err != nil {
		return fmt.Errorf("failed to write table summary: %w", err)
	}

	return nil
}

// buildSummary generates the summary line: "Found X secrets (Y critical, Z high, ...)"
func buildSummary(findings []finding.Finding) string {
	counts := map[finding.Severity]int{}
	for _, fd := range findings {
		counts[fd.Severity]++
	}

	total := len(findings)
	if total == 0 {
		return "Found 0 secrets."
	}

	var parts []string
	// Order: critical, high, medium, low.
	for _, sev := range []finding.Severity{
		finding.SeverityCritical,
		finding.SeverityHigh,
		finding.SeverityMedium,
		finding.SeverityLow,
	} {
		if c, ok := counts[sev]; ok && c > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c, sev.String()))
		}
	}

	return fmt.Sprintf("Found %d secrets (%s).", total, strings.Join(parts, ", "))
}

// FileExtension returns the text file extension.
func (f *Formatter) FileExtension() string {
	return ".txt"
}
