// Package sentry provides a verifier for Sentry authentication tokens.
// It uses the auth-required Sentry API GET /api/0/organizations/ endpoint to
// check token validity. The API root (/api/0/) responds 200 without
// authentication, so it cannot distinguish a valid token from an invalid one
// and must not be used for verification.
package sentry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/internal/verifier/internal/httpx"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "sentry-token"

// defaultAPIURL is the base URL for the Sentry API.
const defaultAPIURL = "https://sentry.io"

// Verifier checks whether a Sentry token is active by calling the
// Sentry API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Sentry API base URL (for testing).
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

// Verify checks if the detected Sentry token is valid/active.
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

	// Use an auth-required endpoint: /api/0/ responds 200 without authentication
	// (false positive), whereas /api/0/organizations/ returns 401 for an invalid
	// token.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/api/0/organizations/", nil)
	if err != nil {
		slog.ErrorContext(ctx, "sentry verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = httpx.Client()
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "sentry verifier: request failed", slog.String("error", err.Error()))
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
		slog.InfoContext(ctx, "sentry verifier: token is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Sentry token is active",
		}
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "sentry verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Sentry token is invalid or revoked",
		}
	default:
		slog.ErrorContext(
			ctx, "sentry verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
