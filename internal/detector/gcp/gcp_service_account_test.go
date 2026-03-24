package gcp

import (
	"context"
	"testing"

	"github.com/cemililik/leakwatch/pkg/finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetector_Metadata_ReturnsExpectedValues(t *testing.T) {
	d := &Detector{}
	assert.Equal(t, "gcp-service-account", d.ID())
	assert.Equal(t, "GCP Service Account Key", d.Description())
	assert.Equal(t, finding.SeverityCritical, d.Severity())
	assert.NotEmpty(t, d.Keywords())
}

func TestDetector_Scan_MatchesValidServiceAccount(t *testing.T) {
	fullJSON := `{
  "type": "service_account",
  "project_id": "my-project-123",
  "private_key_id": "abcdef1234567890abcdef1234567890abcdef12",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\nREDACTED\n-----END RSA PRIVATE KEY-----\n",
  "client_email": "my-service@my-project-123.iam.gserviceaccount.com",
  "client_id": "123456789012345678901"
}`

	minimalJSON := `{
  "type": "service_account",
  "project_id": "test-project"
}`

	spacedJSON := `{
  "type"  :  "service_account",
  "private_key_id": "def0123456789abcdef0123456789abcdef01234",
  "client_email": "svc@proj.iam.gserviceaccount.com"
}`

	tests := []struct {
		name     string
		input    string
		expected int
		redacted string
		keyID    string
		email    string
	}{
		{
			name:     "full service account JSON",
			input:    fullJSON,
			expected: 1,
			redacted: "****@*.iam.gserviceaccount.com",
			keyID:    "abcdef1234567890abcdef1234567890abcdef12",
			email:    "my-service@my-project-123.iam.gserviceaccount.com",
		},
		{
			name:     "minimal JSON without private key id",
			input:    minimalJSON,
			expected: 1,
			redacted: "GCP Service Account Key ****",
		},
		{
			name:     "JSON with extra spaces around colon",
			input:    spacedJSON,
			expected: 1,
			redacted: "****@*.iam.gserviceaccount.com",
			keyID:    "def0123456789abcdef0123456789abcdef01234",
			email:    "svc@proj.iam.gserviceaccount.com",
		},
	}

	d := &Detector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := d.Scan(context.Background(), []byte(tt.input))
			assert.Len(t, findings, tt.expected)
			if tt.expected > 0 {
				require.NotEmpty(t, findings)
				assert.Equal(t, tt.redacted, findings[0].Redacted)
				if tt.keyID != "" {
					assert.Equal(t, tt.keyID, string(findings[0].Raw))
					assert.Equal(t, tt.keyID, findings[0].ExtraData["private_key_id"])
				}
				if tt.email != "" {
					assert.Equal(t, tt.email, findings[0].ExtraData["client_email"])
				}
			}
		})
	}
}

func TestDetector_Scan_RejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "type user_account instead of service_account",
			input: `{"type": "user_account", "client_email": "user@example.com"}`,
		},
		{
			name:  "plain text with service_account word",
			input: "Please create a service_account for this project",
		},
		{
			name:  "empty JSON object",
			input: "{}",
		},
		{
			name:  "plain text",
			input: "this is just normal text",
		},
		{
			name:  "empty input",
			input: "",
		},
	}

	d := &Detector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := d.Scan(context.Background(), []byte(tt.input))
			assert.Empty(t, findings)
		})
	}
}
