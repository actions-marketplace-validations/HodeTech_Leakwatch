// Package heroku provides a verifier for Heroku API keys.
// It uses the Heroku API GET /account endpoint to check token validity.
package heroku

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

const detectorID = "heroku-api-key"

// defaultAPIURL is the base URL for the Heroku API.
const defaultAPIURL = "https://api.heroku.com"

// Verifier checks whether a Heroku API key is active by calling the
// Heroku API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Heroku API base URL (for testing).
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

// Verify checks if the detected Heroku API key is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/account", nil)
	if err != nil {
		slog.ErrorContext(ctx, "heroku verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "heroku verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return handleActiveToken(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "heroku verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Heroku API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "heroku verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Heroku API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var account struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(body).Decode(&account); err != nil {
		slog.ErrorContext(ctx, "heroku verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Heroku API key is active (could not parse account info)",
		}
	}

	extra := map[string]string{
		"email": account.Email,
	}

	slog.InfoContext(ctx, "heroku verifier: token is active",
		slog.String("email", account.Email),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Heroku API key is active",
		ExtraData: extra,
	}
}
