// Package github also provides a verifier for GitHub OAuth tokens.
// It uses the GitHub API GET /user endpoint to check token validity.
package github

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

const oauthDetectorID = "github-oauth-token"

// OAuthVerifier checks whether a GitHub OAuth token is active by calling the
// GitHub API. It NEVER logs or persists raw token values.
type OAuthVerifier struct {
	// apiURL overrides the GitHub API base URL (for testing).
	apiURL string
	// httpClient overrides the default HTTP client (for testing).
	httpClient *http.Client
}

func init() {
	verifier.Register(&OAuthVerifier{})
}

// Type returns the detector ID this verifier handles.
func (v *OAuthVerifier) Type() string {
	return oauthDetectorID
}

// Verify checks if the detected GitHub OAuth token is valid/active.
// Raw contains the token value.
func (v *OAuthVerifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/user", nil)
	if err != nil {
		slog.ErrorContext(ctx, "github oauth verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "github oauth verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return handleActiveOAuthToken(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "github oauth verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "GitHub OAuth token is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "github oauth verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveOAuthToken parses the GitHub API response for a valid OAuth token.
func handleActiveOAuthToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var user struct {
		Login string `json:"login"`
	}

	if err := json.NewDecoder(body).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "github oauth verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "GitHub OAuth token is active (could not parse user info)",
		}
	}

	extra := map[string]string{
		"login": user.Login,
	}

	slog.InfoContext(ctx, "github oauth verifier: token is active",
		slog.String("login", user.Login),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "GitHub OAuth token is active",
		ExtraData: extra,
	}
}
