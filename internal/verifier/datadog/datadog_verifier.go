// Package datadog provides a verifier for Datadog API keys.
// It uses the Datadog API GET /api/v1/validate endpoint to check key validity.
package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "datadog-api-key"

// defaultAPIURL is the base URL for the Datadog API.
const defaultAPIURL = "https://api.datadoghq.com"

// Verifier checks whether a Datadog API key is active by calling the
// Datadog validation API. It NEVER logs or persists raw key values.
type Verifier struct {
	// apiURL overrides the Datadog API base URL (for testing).
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

// Verify checks if the detected Datadog API key is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/api/v1/validate", nil)
	if err != nil {
		slog.ErrorContext(ctx, "datadog verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("DD-API-KEY", token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "datadog verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return handleOKResponse(ctx, resp)
	case http.StatusForbidden:
		slog.DebugContext(ctx, "datadog verifier: API key is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Datadog API key is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "datadog verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleOKResponse parses the Datadog validation response to determine
// whether the key is valid or invalid.
func handleOKResponse(ctx context.Context, resp *http.Response) finding.VerificationResult {
	var body struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		slog.ErrorContext(ctx, "datadog verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to decode response: %v", err),
		}
	}

	if body.Valid {
		slog.InfoContext(ctx, "datadog verifier: API key is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Datadog API key is active",
		}
	}

	slog.DebugContext(ctx, "datadog verifier: API key is inactive")
	return finding.VerificationResult{
		Status:  finding.StatusVerifiedInactive,
		Message: "Datadog API key is invalid",
	}
}
