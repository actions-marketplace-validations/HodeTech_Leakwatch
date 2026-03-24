// Package json provides a JSON output formatter.
package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cemililik/leakwatch/pkg/finding"
)

// Formatter outputs findings in JSON format.
type Formatter struct {
	// ShowRaw controls whether the Raw field is included in output.
	// When false, the Raw field is stripped for defense in depth.
	ShowRaw bool
}

// Format writes findings as JSON to the given writer.
// When ShowRaw is false, Raw fields are actively cleared.
func (f *Formatter) Format(w io.Writer, findings []finding.Finding) error {
	output := make([]finding.Finding, len(findings))
	copy(output, findings)

	if !f.ShowRaw {
		for i := range output {
			output[i].Raw = ""
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}

// FileExtension returns the JSON file extension.
func (f *Formatter) FileExtension() string {
	return ".json"
}
