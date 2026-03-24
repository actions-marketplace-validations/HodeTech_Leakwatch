// Package dockerhub provides a verifier for Docker Hub Personal Access Tokens.
// It uses the Docker Hub API GET /v2/user/ endpoint to check token validity.
package dockerhub

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

const detectorID = "dockerhub-pat"

// defaultAPIURL is the base URL for the Docker Hub API.
const defaultAPIURL = "https://hub.docker.com"

// Verifier checks whether a Docker Hub PAT is active by calling the
// Docker Hub API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the Docker Hub API base URL (for testing).
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

// Verify checks if the detected Docker Hub PAT is valid/active.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v2/user/", nil)
	if err != nil {
		slog.ErrorContext(ctx, "dockerhub verifier: failed to create request", slog.String("error", err.Error()))
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
		slog.ErrorContext(ctx, "dockerhub verifier: request failed", slog.String("error", err.Error()))
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
		slog.DebugContext(ctx, "dockerhub verifier: PAT is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Docker Hub PAT is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "dockerhub verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Docker Hub API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var user struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(body).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "dockerhub verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Docker Hub PAT is active (could not parse user info)",
		}
	}

	extra := map[string]string{
		"username": user.Username,
	}

	slog.InfoContext(ctx, "dockerhub verifier: PAT is active",
		slog.String("username", user.Username),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Docker Hub PAT is active",
		ExtraData: extra,
	}
}
