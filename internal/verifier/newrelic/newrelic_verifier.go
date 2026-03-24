// Package newrelic provides a verifier for New Relic API keys.
// It uses the New Relic API GET /v2/users.json endpoint to check token validity.
package newrelic

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "newrelic-api-key"

// defaultAPIURL is the base URL for the New Relic API.
const defaultAPIURL = "https://api.newrelic.com"

// Verifier checks whether a New Relic API key is active by calling the
// New Relic API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the New Relic API base URL (for testing).
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

// Verify checks if the detected New Relic API key is valid/active.
// Raw contains the token value.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v2/users.json", nil)
	if err != nil {
		slog.ErrorContext(ctx, "newrelic verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Api-Key", token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "newrelic verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "newrelic verifier: token is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "New Relic API key is active",
		}
	case http.StatusUnauthorized, http.StatusForbidden:
		slog.DebugContext(ctx, "newrelic verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "New Relic API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "newrelic verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
