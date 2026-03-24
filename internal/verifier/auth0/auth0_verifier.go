// Package auth0 provides a verifier for Auth0 Management API tokens.
// It uses the Auth0 Management API GET /api/v2/ endpoint with Bearer auth
// to check token validity.
package auth0

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "auth0-management-token"

// defaultAPIURL is the base URL for the Auth0 Management API.
const defaultAPIURL = "https://login.auth0.com"

// Verifier checks whether an Auth0 Management API token is active by calling
// the Auth0 Management API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Auth0 API base URL (for testing).
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

// Verify checks if the detected Auth0 Management API token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/api/v2/", nil)
	if err != nil {
		slog.ErrorContext(ctx, "auth0 verifier: failed to create request", slog.String("error", err.Error()))
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
		slog.ErrorContext(ctx, "auth0 verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "auth0 verifier: management token is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Auth0 management token is active",
		}
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "auth0 verifier: management token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Auth0 management token is invalid or expired",
		}
	default:
		slog.ErrorContext(ctx, "auth0 verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
