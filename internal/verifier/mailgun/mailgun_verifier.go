// Package mailgun provides a verifier for Mailgun API keys.
// It uses the Mailgun API GET /v3/domains endpoint with Basic auth to check key validity.
package mailgun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "mailgun-api-key"

// defaultAPIURL is the base URL for the Mailgun API.
const defaultAPIURL = "https://api.mailgun.net"

// Verifier checks whether a Mailgun API key is active by calling the
// Mailgun API. It NEVER logs or persists raw key values.
type Verifier struct {
	// apiURL overrides the Mailgun API base URL (for testing).
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

// Verify checks if the detected Mailgun API key is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v3/domains", nil)
	if err != nil {
		slog.ErrorContext(ctx, "mailgun verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.SetBasicAuth("api", token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "mailgun verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "mailgun verifier: API key is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Mailgun API key is active",
		}
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "mailgun verifier: API key is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Mailgun API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "mailgun verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
