// Package snyk provides a verifier for Snyk API keys.
// It uses the Snyk REST API GET /rest/self endpoint to check key validity.
package snyk

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "snyk-api-key"

// defaultAPIURL is the base URL for the Snyk API.
const defaultAPIURL = "https://api.snyk.io"

// Verifier checks whether a Snyk API key is active by calling the
// Snyk REST API. It NEVER logs or persists raw key values.
type Verifier struct {
	// apiURL overrides the Snyk API base URL (for testing).
	apiURL string
	// httpClient overrides the default HTTP client (for testing).
	httpClient *http.Client
}

func init() {
	verifier.Register(&Verifier{})
}

// Type returns the detector ID this verifier handles.
func (v *Verifier) Type() string {
	return detectorID
}

// Verify checks if the detected Snyk API key is valid/active.
// Raw contains the key value.
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	token := string(raw.Raw)
	if token == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty token",
		}
	}

	apiURL := v.apiURL
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/rest/self", nil)
	if err != nil {
		slog.ErrorContext(ctx, "snyk verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Version", "2024-04-29")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "snyk verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "snyk verifier: API key is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Snyk API key is active",
		}
	case http.StatusUnauthorized, http.StatusForbidden:
		slog.DebugContext(ctx, "snyk verifier: API key is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Snyk API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "snyk verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
