// Package twilio provides a verifier for Twilio API keys.
// It uses the Twilio Accounts API GET /2010-04-01/Accounts.json endpoint
// with Basic auth to check key validity.
package twilio

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "twilio-api-key"

// defaultAPIURL is the base URL for the Twilio API.
const defaultAPIURL = "https://api.twilio.com"

// Verifier checks whether a Twilio API key is active by calling the
// Twilio Accounts API. It NEVER logs or persists raw key values.
type Verifier struct {
	// apiURL overrides the Twilio API base URL (for testing).
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

// Verify checks if the detected Twilio API key is valid/active.
// The Account SID must be provided in raw.ExtraData["account_sid"].
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	token := string(raw.Raw)
	if token == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty token",
		}
	}

	accountSID := ""
	if raw.ExtraData != nil {
		accountSID = raw.ExtraData["account_sid"]
	}
	if accountSID == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "Account SID required",
		}
	}

	apiURL := v.apiURL
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/2010-04-01/Accounts.json", nil)
	if err != nil {
		slog.ErrorContext(ctx, "twilio verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.SetBasicAuth(accountSID, token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "twilio verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "twilio verifier: API key is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Twilio API key is active",
		}
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "twilio verifier: API key is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Twilio API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "twilio verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
