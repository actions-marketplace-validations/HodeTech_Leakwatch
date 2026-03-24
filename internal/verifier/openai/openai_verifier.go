// Package openai provides a verifier for OpenAI API keys.
// It uses the OpenAI API GET /v1/models endpoint to check key validity.
package openai

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

const detectorID = "openai-api-key"

// defaultAPIURL is the base URL for the OpenAI API.
const defaultAPIURL = "https://api.openai.com"

// Verifier checks whether an OpenAI API key is active by calling the
// OpenAI API. It NEVER logs or persists raw key values.
type Verifier struct {
	// apiURL overrides the OpenAI API base URL (for testing).
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

// Verify checks if the detected OpenAI API key is valid/active.
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
		slog.ErrorContext(ctx, "openai verifier: failed to create request", slog.String("error", err.Error()))
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
		slog.ErrorContext(ctx, "openai verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return handleActiveKey(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "openai verifier: API key is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "OpenAI API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "openai verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveKey parses the OpenAI API response for a valid key.
func handleActiveKey(ctx context.Context, body io.Reader) finding.VerificationResult {
	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(body).Decode(&models); err != nil {
		slog.ErrorContext(ctx, "openai verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "OpenAI API key is active (could not parse model list)",
		}
	}

	modelCount := fmt.Sprintf("%d", len(models.Data))
	extra := map[string]string{
		"model_count": modelCount,
	}

	slog.InfoContext(ctx, "openai verifier: API key is active",
		slog.String("model_count", modelCount),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "OpenAI API key is active",
		ExtraData: extra,
	}
}
