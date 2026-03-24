package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/cemililik/leakwatch/pkg/finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_Format_EmptyFindings_WritesEmptyArray(t *testing.T) {
	f := &Formatter{}
	var buf bytes.Buffer

	err := f.Format(&buf, []finding.Finding{})
	require.NoError(t, err)

	var result []finding.Finding
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFormatter_Format_SingleFinding_WritesValidJSON(t *testing.T) {
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
			},
			DetectedAt: time.Now(),
		},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	var result []finding.Finding
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "aws-access-key-id", result[0].DetectorID)
	assert.Equal(t, "AKIA****MPLE", result[0].Redacted)
}

func TestFormatter_Format_OmitsRawWhenEmpty(t *testing.T) {
	f := &Formatter{}
	var buf bytes.Buffer

	findings := []finding.Finding{
		{ID: "test", Redacted: "****"},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	var rawJSON []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &rawJSON)
	require.NoError(t, err)

	_, hasRaw := rawJSON[0]["raw"]
	assert.False(t, hasRaw, "raw boşken JSON çıktısında olmamalı")
}

func TestFormatter_Format_ShowRawFalse_StripsRawFromOutput(t *testing.T) {
	f := &Formatter{ShowRaw: false}
	var buf bytes.Buffer

	findings := []finding.Finding{
		{
			ID:       "test-1",
			Redacted: "AKIA****MPLE",
			Raw:      "AKIAIOSFODNN7EXAMPLE",
		},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	var rawJSON []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &rawJSON)
	require.NoError(t, err)

	_, hasRaw := rawJSON[0]["raw"]
	assert.False(t, hasRaw, "ShowRaw=false iken Raw alanı JSON çıktısında olmamalı")
}

func TestFormatter_Format_ShowRawTrue_IncludesRawInOutput(t *testing.T) {
	f := &Formatter{ShowRaw: true}
	var buf bytes.Buffer

	findings := []finding.Finding{
		{
			ID:       "test-1",
			Redacted: "AKIA****MPLE",
			Raw:      "AKIAIOSFODNN7EXAMPLE",
		},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	var rawJSON []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &rawJSON)
	require.NoError(t, err)

	rawVal, hasRaw := rawJSON[0]["raw"]
	assert.True(t, hasRaw, "ShowRaw=true iken Raw alanı JSON çıktısında olmalı")
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", rawVal)
}

func TestFormatter_Format_ShowRawFalse_DoesNotMutateOriginal(t *testing.T) {
	f := &Formatter{ShowRaw: false}
	var buf bytes.Buffer

	findings := []finding.Finding{
		{
			ID:  "test-1",
			Raw: "AKIAIOSFODNN7EXAMPLE",
		},
	}

	err := f.Format(&buf, findings)
	require.NoError(t, err)

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", findings[0].Raw, "Format orijinal slice'ı değiştirmemeli")
}

func TestFormatter_FileExtension_ReturnsJSON(t *testing.T) {
	f := &Formatter{}
	assert.Equal(t, ".json", f.FileExtension())
}

// errWriter, yazım hatası simüle eden writer.
type errWriter struct{}

func (w *errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestFormatter_Format_WriterError_ReturnsError(t *testing.T) {
	f := &Formatter{}
	findings := []finding.Finding{{ID: "test"}}

	err := f.Format(&errWriter{}, findings)
	assert.Error(t, err)
}
