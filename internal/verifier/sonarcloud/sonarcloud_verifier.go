// Package sonarcloud provides a verifier for SonarCloud tokens.
// It uses the SonarCloud API GET /api/authentication/validate endpoint to check token validity.
package sonarcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/internal/verifier/internal/httpx"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "sonarcloud-token"

// defaultAPIURL is the base URL for the SonarCloud API.
const defaultAPIURL = "https://sonarcloud.io"

// Verifier checks whether a SonarCloud token is active by calling the
// SonarCloud API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the SonarCloud API base URL (for testing).
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

// Verify checks if the detected SonarCloud token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/api/authentication/validate", nil)
	if err != nil {
		slog.ErrorContext(ctx, "sonarcloud verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.SetBasicAuth(token, "")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = httpx.Client()
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "sonarcloud verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// A redirect from an API endpoint means the credential context is wrong
	// (for example a login redirect or a moved host). The shared client does
	// not follow redirects so the credential is never re-sent to the redirect
	// target; treat it as a verification error rather than an active secret.
	if httpx.IsRedirect(resp.StatusCode) {
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected redirect (status %d)", resp.StatusCode),
		}
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return handleValidateResponse(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "sonarcloud verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "SonarCloud token is invalid or revoked",
		}
	default:
		slog.ErrorContext(
			ctx, "sonarcloud verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleValidateResponse parses the SonarCloud validation response.
func handleValidateResponse(ctx context.Context, body io.Reader) finding.VerificationResult {
	var validation struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(httpx.LimitReader(body)).Decode(&validation); err != nil {
		slog.ErrorContext(ctx, "sonarcloud verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to decode response: %v", err),
		}
	}

	if validation.Valid {
		slog.InfoContext(ctx, "sonarcloud verifier: token is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "SonarCloud token is active",
		}
	}

	slog.DebugContext(ctx, "sonarcloud verifier: token is inactive (valid=false)")
	return finding.VerificationResult{
		Status:  finding.StatusVerifiedInactive,
		Message: "SonarCloud token is invalid or revoked",
	}
}
