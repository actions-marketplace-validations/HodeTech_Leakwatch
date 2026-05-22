// Package anthropic provides a verifier for Anthropic API keys.
// It uses the Anthropic API GET /v1/models endpoint to check key validity.
package anthropic

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

const detectorID = "anthropic-api-key"

// defaultAPIURL is the base URL for the Anthropic API.
const defaultAPIURL = "https://api.anthropic.com"

// Verifier checks whether an Anthropic API key is active by calling the
// Anthropic API. It NEVER logs or persists raw key values.
type Verifier struct {
	// apiURL overrides the Anthropic API base URL (for testing).
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

// Verify checks if the detected Anthropic API key is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v1/models", nil)
	if err != nil {
		slog.ErrorContext(ctx, "anthropic verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("x-api-key", token)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = httpx.Client()
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "anthropic verifier: request failed", slog.String("error", err.Error()))
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
		return handleActiveKey(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "anthropic verifier: API key is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Anthropic API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(
			ctx, "anthropic verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveKey parses the Anthropic API response for a valid key.
func handleActiveKey(ctx context.Context, body io.Reader) finding.VerificationResult {
	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(httpx.LimitReader(body)).Decode(&models); err != nil {
		slog.ErrorContext(ctx, "anthropic verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("200 OK but failed to decode response body: %v", err),
		}
	}

	modelCount := fmt.Sprintf("%d", len(models.Data))
	extra := map[string]string{
		"model_count": modelCount,
	}

	slog.InfoContext(
		ctx, "anthropic verifier: API key is active",
		slog.String("model_count", modelCount),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Anthropic API key is active",
		ExtraData: extra,
	}
}
