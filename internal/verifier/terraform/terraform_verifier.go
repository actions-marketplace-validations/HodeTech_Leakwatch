// Package terraform provides a verifier for Terraform Cloud API tokens.
// It uses the Terraform Cloud API GET /api/v2/account/details endpoint to check token validity.
package terraform

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

const detectorID = "terraform-cloud-token"

// defaultAPIURL is the base URL for the Terraform Cloud API.
const defaultAPIURL = "https://app.terraform.io"

// Verifier checks whether a Terraform Cloud API token is active by calling the
// Terraform Cloud API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Terraform Cloud API base URL (for testing).
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

// Verify checks if the detected Terraform Cloud API token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/api/v2/account/details", nil)
	if err != nil {
		slog.ErrorContext(ctx, "terraform verifier: failed to create request", slog.String("error", err.Error()))
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
		slog.ErrorContext(ctx, "terraform verifier: request failed", slog.String("error", err.Error()))
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
		slog.DebugContext(ctx, "terraform verifier: API token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Terraform Cloud API token is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "terraform verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Terraform Cloud API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var account struct {
		Data struct {
			Attributes struct {
				Username string `json:"username"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(body).Decode(&account); err != nil {
		slog.ErrorContext(ctx, "terraform verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Terraform Cloud API token is active (could not parse account details)",
		}
	}

	extra := map[string]string{
		"username": account.Data.Attributes.Username,
	}

	slog.InfoContext(ctx, "terraform verifier: API token is active",
		slog.String("username", account.Data.Attributes.Username),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Terraform Cloud API token is active",
		ExtraData: extra,
	}
}
