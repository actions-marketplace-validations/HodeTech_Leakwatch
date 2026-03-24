// Package discord provides a verifier for Discord Bot tokens.
// It uses the Discord API GET /users/@me endpoint to check token validity.
package discord

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

const detectorID = "discord-bot-token"

// defaultAPIURL is the base URL for the Discord API.
const defaultAPIURL = "https://discord.com/api/v10"

// Verifier checks whether a Discord Bot token is active by calling the
// Discord API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Discord API base URL (for testing).
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

// Verify checks if the detected Discord Bot token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/users/@me", nil)
	if err != nil {
		slog.ErrorContext(ctx, "discord verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "discord verifier: request failed", slog.String("error", err.Error()))
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
		slog.DebugContext(ctx, "discord verifier: token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Discord Bot token is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "discord verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Discord API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var user struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(body).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "discord verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Discord Bot token is active (could not parse user info)",
		}
	}

	extra := map[string]string{
		"username": user.Username,
	}

	slog.InfoContext(ctx, "discord verifier: token is active",
		slog.String("username", user.Username),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Discord Bot token is active",
		ExtraData: extra,
	}
}
