// Package sendgrid provides a verifier for SendGrid API keys.
// It uses the SendGrid API GET /v3/user/profile endpoint to check key validity.
package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "sendgrid-api-key"

// defaultAPIURL is the base URL for the SendGrid API.
const defaultAPIURL = "https://api.sendgrid.com"

// Verifier checks whether a SendGrid API key is active by calling the
// SendGrid API. It NEVER logs or persists raw key values.
type Verifier struct {
	// apiURL overrides the SendGrid API base URL (for testing).
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

// Verify checks if the detected SendGrid API key is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v3/user/profile", nil)
	if err != nil {
		slog.ErrorContext(ctx, "sendgrid verifier: failed to create request", slog.String("error", err.Error()))
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
		slog.ErrorContext(ctx, "sendgrid verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return handleActiveKey(ctx, resp.Body)
	case http.StatusUnauthorized, http.StatusForbidden:
		slog.DebugContext(ctx, "sendgrid verifier: API key is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "SendGrid API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "sendgrid verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveKey parses the SendGrid API response for a valid key.
func handleActiveKey(ctx context.Context, body io.Reader) finding.VerificationResult {
	var profile struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(body).Decode(&profile); err != nil {
		slog.ErrorContext(ctx, "sendgrid verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "SendGrid API key is active (could not parse user info)",
		}
	}

	extra := map[string]string{
		"username": profile.Username,
	}

	slog.InfoContext(ctx, "sendgrid verifier: API key is active",
		slog.String("username", profile.Username),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "SendGrid API key is active",
		ExtraData: extra,
	}
}
