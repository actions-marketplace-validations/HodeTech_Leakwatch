package gcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

func TestVerifier_Type_ReturnsCorrectID(t *testing.T) {
	v := &Verifier{}
	assert.Equal(t, "gcp-service-account", v.Type())
}

func TestVerify_ValidServiceAccountKey_ReturnsActive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw: []byte(`{
			"type": "service_account",
			"project_id": "my-project-123",
			"private_key_id": "key-id-abc",
			"private_key": "-----BEGIN RSA PRIVATE KEY-----\nREDACTED\n-----END RSA PRIVATE KEY-----\n",
			"client_email": "sa@my-project-123.iam.gserviceaccount.com",
			"client_id": "123456789",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token"
		}`),
		Redacted: "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedActive, result.Status)
	assert.Equal(t, "Service account key format validated", result.Message)
	assert.Equal(t, "my-project-123", result.ExtraData["project_id"])
	assert.Equal(t, "sa@my-project-123.iam.gserviceaccount.com", result.ExtraData["client_email"])
}

func TestVerify_InvalidJSON_ReturnsInactive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(`{not valid json`),
		Redacted:   "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
	assert.Equal(t, "invalid JSON structure", result.Message)
}

func TestVerify_WrongType_ReturnsInactive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw: []byte(`{
			"type": "authorized_user",
			"project_id": "my-project",
			"private_key_id": "key-id",
			"client_email": "user@example.com"
		}`),
		Redacted: "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
	assert.Equal(t, "JSON type field is not service_account", result.Message)
}

func TestVerify_MissingFields_ReturnsInactive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw: []byte(`{
			"type": "service_account",
			"project_id": "my-project"
		}`),
		Redacted: "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
	assert.Equal(t, "missing required fields in service account key", result.Message)
}

func TestVerify_EmptyInput_ReturnsUnverified(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(""),
		Redacted:   "",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusUnverified, result.Status)
	assert.Equal(t, "empty input", result.Message)
}
