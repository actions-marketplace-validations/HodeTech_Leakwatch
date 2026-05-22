// Package notion provides a verifier for Notion integration tokens.
// It uses the Notion API GET /v1/users/me endpoint to check token validity.
package notion

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

const detectorID = "notion-token"

// defaultAPIURL is the base URL for the Notion API.
const defaultAPIURL = "https://api.notion.com"

// notionVersion is the required Notion-Version header value.
const notionVersion = "2022-06-28"

// Verifier checks whether a Notion integration token is active by calling the
// Notion API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Notion API base URL (for testing).
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

// Verify checks if the detected Notion token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v1/users/me", nil)
	if err != nil {
		slog.ErrorContext(ctx, "notion verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = httpx.Client()
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "notion verifier: request failed", slog.String("error", err.Error()))
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
		return handleActiveToken(ctx, resp.Body)
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "notion verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Notion token is invalid or revoked",
		}
	default:
		slog.ErrorContext(
			ctx, "notion verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Notion API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var user struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(httpx.LimitReader(body)).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "notion verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("200 OK but failed to decode response body: %v", err),
		}
	}

	extra := map[string]string{
		"name": user.Name,
	}

	slog.InfoContext(
		ctx, "notion verifier: token is active",
		slog.String("name", user.Name),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Notion token is active",
		ExtraData: extra,
	}
}
