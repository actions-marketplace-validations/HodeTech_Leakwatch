// Package slack provides a verifier for Slack Bot/User tokens.
// It calls the Slack auth.test API endpoint to check token validity.
package slack

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

const detectorID = "slack-token"

// defaultAPIURL is the base URL for the Slack API.
const defaultAPIURL = "https://slack.com/api"

// Verifier checks whether a Slack token is active by calling the
// Slack auth.test endpoint. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Slack API base URL (for testing).
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

// Verify checks if the detected Slack token is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/auth.test", nil)
	if err != nil {
		slog.ErrorContext(ctx, "slack verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "slack verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		Team  string `json:"team"`
		User  string `json:"user"`
		URL   string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.ErrorContext(ctx, "slack verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to decode response: %v", err),
		}
	}

	if !result.OK {
		slog.DebugContext(ctx, "slack verifier: token is inactive", slog.String("error", result.Error))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: fmt.Sprintf("Slack token is invalid: %s", result.Error),
		}
	}

	extra := map[string]string{}
	if result.Team != "" {
		extra["team"] = result.Team
	}
	if result.User != "" {
		extra["user"] = result.User
	}

	slog.InfoContext(ctx, "slack verifier: token is active",
		slog.String("team", result.Team),
		slog.String("user", result.User),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Slack token is active",
		ExtraData: extra,
	}
}
