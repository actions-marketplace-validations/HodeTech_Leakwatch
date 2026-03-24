// Package okta provides a verifier for Okta API tokens.
// It uses the Okta Users API GET /api/v1/users/me endpoint to check token validity.
// Okta uses the SSWS authorization scheme, not Bearer.
package okta

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

const detectorID = "okta-api-token"

// Verifier checks whether an Okta API token is active by calling the
// Okta Users API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Okta API base URL (for testing).
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

// Verify checks if the detected Okta API token is valid/active.
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
		if domain, ok := raw.ExtraData["domain"]; ok && domain != "" {
			apiURL = "https://" + domain
		} else {
			return finding.VerificationResult{
				Status:  finding.StatusVerifyError,
				Message: "Okta domain required",
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/api/v1/users/me", nil)
	if err != nil {
		slog.ErrorContext(ctx, "okta verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "SSWS "+token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "okta verifier: request failed", slog.String("error", err.Error()))
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
		slog.DebugContext(ctx, "okta verifier: API token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Okta API token is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "okta verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Okta API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var user struct {
		Profile struct {
			Login string `json:"login"`
		} `json:"profile"`
	}

	if err := json.NewDecoder(body).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "okta verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Okta API token is active (could not parse user info)",
		}
	}

	extra := map[string]string{
		"login": user.Profile.Login,
	}

	slog.InfoContext(ctx, "okta verifier: API token is active",
		slog.String("login", user.Profile.Login),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Okta API token is active",
		ExtraData: extra,
	}
}
