// Package bitbucket provides a verifier for Bitbucket app passwords.
// It uses the Bitbucket User API GET /2.0/user endpoint with Basic auth
// to check app password validity.
package bitbucket

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

const detectorID = "bitbucket-app-password"

// defaultAPIURL is the base URL for the Bitbucket API.
const defaultAPIURL = "https://api.bitbucket.org"

// Verifier checks whether a Bitbucket app password is active by calling the
// Bitbucket User API. It NEVER logs or persists raw password values.
type Verifier struct {
	// apiURL overrides the Bitbucket API base URL (for testing).
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

// Verify checks if the detected Bitbucket app password is valid/active.
// The Bitbucket username must be provided in raw.ExtraData["username"].
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	token := string(raw.Raw)
	if token == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty token",
		}
	}

	username := ""
	if raw.ExtraData != nil {
		username = raw.ExtraData["username"]
	}
	if username == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "Bitbucket username required",
		}
	}

	apiURL := v.apiURL
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/2.0/user", nil)
	if err != nil {
		slog.ErrorContext(ctx, "bitbucket verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.SetBasicAuth(username, token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = httpx.Client()
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "bitbucket verifier: request failed", slog.String("error", err.Error()))
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
		return handleActivePassword(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "bitbucket verifier: app password is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Bitbucket app password is invalid or revoked",
		}
	default:
		slog.ErrorContext(
			ctx, "bitbucket verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActivePassword parses the Bitbucket API response for a valid password.
func handleActivePassword(ctx context.Context, body io.Reader) finding.VerificationResult {
	var user struct {
		DisplayName string `json:"display_name"`
	}

	if err := json.NewDecoder(httpx.LimitReader(body)).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "bitbucket verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("200 OK but failed to decode response body: %v", err),
		}
	}

	extra := map[string]string{
		"display_name": user.DisplayName,
	}

	slog.InfoContext(
		ctx, "bitbucket verifier: app password is active",
		slog.String("display_name", user.DisplayName),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Bitbucket app password is active",
		ExtraData: extra,
	}
}
