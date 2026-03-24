// Package csv provides a CSV output formatter for Leakwatch findings.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/cemililik/leakwatch/pkg/finding"
)

// Formatter outputs findings in CSV format.
type Formatter struct {
	// ShowRaw, when true, includes the Raw field as a column.
	// When false, the Raw field is stripped for defense in depth.
	ShowRaw bool
}

// Format writes findings as CSV to the given writer.
// Header: id,detector_id,severity,redacted,file_path,commit,verification_status
// When ShowRaw is false, the Raw field is actively stripped from the output.
func (f *Formatter) Format(w io.Writer, findings []finding.Finding) error {
	output := make([]finding.Finding, len(findings))
	copy(output, findings)

	if !f.ShowRaw {
		for i := range output {
			output[i].Raw = ""
		}
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header row.
	header := []string{"id", "detector_id", "severity", "redacted", "file_path", "commit", "verification_status"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write one row per finding.
	for _, fd := range output {
		row := []string{
			fd.ID,
			fd.DetectorID,
			fd.Severity.String(),
			fd.Redacted,
			fd.SourceMetadata.FilePath,
			fd.SourceMetadata.Commit,
			fd.Verification.Status.String(),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV output: %w", err)
	}
	return nil
}

// FileExtension returns the CSV file extension.
func (f *Formatter) FileExtension() string {
	return ".csv"
}
