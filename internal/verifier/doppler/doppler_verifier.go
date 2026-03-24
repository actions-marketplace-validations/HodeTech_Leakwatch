// Package doppler provides a verifier for Doppler tokens.
// It uses the Doppler API GET /v3/me endpoint to check token validity.
package doppler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "doppler-token"

// defaultAPIURL is the base URL for the Doppler API.
const defaultAPIURL = "https://api.doppler.com"

// Verifier checks whether a Doppler token is active by calling the
// Doppler API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Doppler API base URL (for testing).
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

// Verify checks if the detected Doppler token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v3/me", nil)
	if err != nil {
		slog.ErrorContext(ctx, "doppler verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "doppler verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "doppler verifier: token is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Doppler token is active",
		}
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "doppler verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Doppler token is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "doppler verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
