// Package gcp provides a verifier for GCP service account keys.
//
// It validates the JSON structure without making any API calls. Because no
// live verification is performed, the result is always StatusUnverified: a
// valid structure does not prove the key is active, and an invalid structure
// does not prove it is inactive.
package gcp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "gcp-service-account"

// Verifier checks whether a GCP service account key has a valid JSON structure.
// It NEVER logs or persists raw key values.
type Verifier struct{}

func init() {
	verifier.Register(&Verifier{})
}

// Type returns the detector ID this verifier handles.
func (v *Verifier) Type() string {
	return detectorID
}

// serviceAccountKey represents the expected structure of a GCP service account key file.
type serviceAccountKey struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	ClientEmail  string `json:"client_email"`
}

// Verify checks if the detected GCP service account key has a valid JSON structure.
// Raw contains the full JSON key file content.
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	data := string(raw.Raw)
	if data == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty input",
		}
	}

	var key serviceAccountKey
	if err := json.Unmarshal(raw.Raw, &key); err != nil {
		slog.DebugContext(
			ctx, "gcp verifier: failed to parse JSON",
			slog.String("error", err.Error()),
		)
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "format invalid (invalid JSON structure); live verification not supported",
		}
	}

	if key.Type != "service_account" {
		slog.DebugContext(
			ctx, "gcp verifier: unexpected type field",
			slog.String("type", key.Type),
		)
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "format invalid (JSON type field is not service_account); live verification not supported",
		}
	}

	if key.ProjectID == "" || key.PrivateKeyID == "" || key.ClientEmail == "" {
		slog.DebugContext(ctx, "gcp verifier: missing required fields")
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "format invalid (missing required fields in service account key); live verification not supported",
		}
	}

	slog.DebugContext(
		ctx, "gcp verifier: service account key format validated",
		slog.String("project_id", key.ProjectID),
		slog.String("client_email", key.ClientEmail),
	)

	return finding.VerificationResult{
		Status:  finding.StatusUnverified,
		Message: "format valid; live verification not supported (would require GCP OAuth2 token exchange)",
		ExtraData: map[string]string{
			"project_id":   key.ProjectID,
			"client_email": key.ClientEmail,
		},
	}
}
