// Package postmark provides a verifier for Postmark Server Tokens.
// It uses the Postmark API GET /server endpoint to check token validity.
package postmark

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

const detectorID = "postmark-server-token"

// defaultAPIURL is the base URL for the Postmark API.
const defaultAPIURL = "https://api.postmarkapp.com"

// Verifier checks whether a Postmark Server Token is active by calling the
// Postmark API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Postmark API base URL (for testing).
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

// Verify checks if the detected Postmark Server Token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/server", nil)
	if err != nil {
		slog.ErrorContext(ctx, "postmark verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("X-Postmark-Server-Token", token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "postmark verifier: request failed", slog.String("error", err.Error()))
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
		slog.DebugContext(ctx, "postmark verifier: server token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Postmark server token is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "postmark verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Postmark API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var server struct {
		Name string `json:"Name"`
	}

	if err := json.NewDecoder(body).Decode(&server); err != nil {
		slog.ErrorContext(ctx, "postmark verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Postmark server token is active (could not parse server info)",
		}
	}

	extra := map[string]string{
		"server_name": server.Name,
	}

	slog.InfoContext(ctx, "postmark verifier: server token is active",
		slog.String("server_name", server.Name),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Postmark server token is active",
		ExtraData: extra,
	}
}
