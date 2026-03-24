package sarif

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/cemililik/leakwatch/pkg/finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_Format_EmptyFindings_WritesValidSARIFWithEmptyResults(t *testing.T) {
	f := &Formatter{}
	var buf bytes.Buffer

	err := f.Format(&buf, []finding.Finding{})
	require.NoError(t, err)

	var doc sarifDocument
	err = json.Unmarshal(buf.Bytes(), &doc)
	require.NoError(t, err)

	assert.Equal(t, sarifSchema, doc.Schema)
	assert.Equal(t, sarifVersion, doc.Version)
	require.Len(t, doc.Runs, 1)
	assert.Equal(t, toolName, doc.Runs[0].Tool.Driver.Name)
	assert.Empty(t, doc.Runs[0].Tool.Driver.Rules)
	assert.Empty(t, doc.Runs[0].Results)
}

func TestFormatter_Format_SingleFinding_WritesCorrectRuleAndResult(t *testing.T) {
	f := &Formatter{}
	var buf bytes.Buffer

	findings := []finding.Finding{
		{
			ID:         "abc123",
			DetectorID: "aws-access-key-id",
			Severity:   finding.SeverityCritical,
			Redacted:   "AKIA****MPLE",
			SourceMetadata: finding.SourceMetadata{
				SourceType: "filesystem",
				FilePath:   "config.yaml",
				Line:       42,
			},
			DetectedAt: time.Now(),
		},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	var doc sarifDocument
	err = json.Unmarshal(buf.Bytes(), &doc)
	require.NoError(t, err)

	// Verify rule.
	require.Len(t, doc.Runs[0].Tool.Driver.Rules, 1)
	rule := doc.Runs[0].Tool.Driver.Rules[0]
	assert.Equal(t, "aws-access-key-id", rule.ID)
	assert.Equal(t, "error", rule.DefaultConfig.Level)

	// Verify result.
	require.Len(t, doc.Runs[0].Results, 1)
	result := doc.Runs[0].Results[0]
	assert.Equal(t, "aws-access-key-id", result.RuleID)
	assert.Equal(t, 0, result.RuleIndex)
	assert.Equal(t, "error", result.Level)
	assert.Contains(t, result.Message.Text, "AKIA****MPLE")

	// Verify location.
	require.Len(t, result.Locations, 1)
	loc := result.Locations[0].PhysicalLocation
	assert.Equal(t, "config.yaml", loc.ArtifactLocation.URI)
	require.NotNil(t, loc.Region)
	assert.Equal(t, 42, loc.Region.StartLine)
}

func TestFormatter_Format_ShowRawFalse_RawNotInOutput(t *testing.T) {
	f := &Formatter{ShowRaw: false}
	var buf bytes.Buffer

	findings := []finding.Finding{
		{
			ID:         "test-1",
			DetectorID: "generic-secret",
			Severity:   finding.SeverityHigh,
			Redacted:   "sk_****abcd",
			Raw:        "sk_live_supersecretvalue",
			SourceMetadata: finding.SourceMetadata{
				FilePath: "app.go",
			},
		},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "sk_live_supersecretvalue",
		"ShowRaw=false must strip raw secret from SARIF output")
}

func TestFormatter_Format_SeverityMapping_MapsCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		severity finding.Severity
		expected string
	}{
		{"Critical maps to error", finding.SeverityCritical, "error"},
		{"High maps to warning", finding.SeverityHigh, "warning"},
		{"Medium maps to note", finding.SeverityMedium, "note"},
		{"Low maps to note", finding.SeverityLow, "note"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Formatter{}
			var buf bytes.Buffer

			findings := []finding.Finding{
				{
					ID:         "sev-test",
					DetectorID: "test-detector",
					Severity:   tt.severity,
					Redacted:   "****",
				},
			}

			err := f.Format(&buf, findings)
			require.NoError(t, err)

			var doc sarifDocument
			err = json.Unmarshal(buf.Bytes(), &doc)
			require.NoError(t, err)

			require.Len(t, doc.Runs[0].Results, 1)
			assert.Equal(t, tt.expected, doc.Runs[0].Results[0].Level)
		})
	}
}

func TestFormatter_Format_ShowRawFalse_DoesNotMutateOriginal(t *testing.T) {
	f := &Formatter{ShowRaw: false}
	var buf bytes.Buffer

	findings := []finding.Finding{
		{
			ID:         "test-1",
			DetectorID: "generic",
			Raw:        "AKIAIOSFODNN7EXAMPLE",
		},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", findings[0].Raw,
		"Format must not mutate the original slice")
}

func TestFormatter_FileExtension_ReturnsSARIF(t *testing.T) {
	f := &Formatter{}
	assert.Equal(t, ".sarif", f.FileExtension())
}
