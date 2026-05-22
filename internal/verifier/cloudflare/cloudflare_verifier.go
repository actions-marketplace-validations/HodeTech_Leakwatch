// Package cloudflare provides a verifier for Cloudflare API tokens.
// It uses the Cloudflare API GET /client/v4/user/tokens/verify endpoint to check token validity.
package cloudflare

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

const detectorID = "cloudflare-api-token"

// defaultAPIURL is the base URL for the Cloudflare API.
const defaultAPIURL = "https://api.cloudflare.com"

// Verifier checks whether a Cloudflare API token is active by calling the
// Cloudflare API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Cloudflare API base URL (for testing).
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

// Verify checks if the detected Cloudflare API token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/client/v4/user/tokens/verify", nil)
	if err != nil {
		slog.ErrorContext(ctx, "cloudflare verifier: failed to create request", slog.String("error", err.Error()))
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
		slog.ErrorContext(ctx, "cloudflare verifier: request failed", slog.String("error", err.Error()))
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
		return handleResponse(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "cloudflare verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Cloudflare API token is invalid or revoked",
		}
	default:
		slog.ErrorContext(
			ctx, "cloudflare verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleResponse parses the Cloudflare API verification response.
// A 200 response may still indicate an inactive token if success is false.
func handleResponse(ctx context.Context, body io.Reader) finding.VerificationResult {
	var response struct {
		Success bool `json:"success"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(httpx.LimitReader(body)).Decode(&response); err != nil {
		slog.ErrorContext(ctx, "cloudflare verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: "failed to decode Cloudflare API response",
		}
	}

	if !response.Success {
		msg := "Cloudflare API token is inactive"
		if len(response.Errors) > 0 {
			msg = fmt.Sprintf("Cloudflare API token is inactive: %s", response.Errors[0].Message)
		}

		slog.DebugContext(ctx, "cloudflare verifier: token verification failed")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: msg,
		}
	}

	slog.InfoContext(ctx, "cloudflare verifier: token is active")

	return finding.VerificationResult{
		Status:  finding.StatusVerifiedActive,
		Message: "Cloudflare API token is active",
	}
}
