// Package teams provides a verifier for Microsoft Teams webhook URLs.
// It sends a minimal POST request to the webhook endpoint to check validity.
package teams

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "teams-webhook"

// Verifier checks whether a Microsoft Teams webhook URL is active by sending
// a minimal POST request. It NEVER logs or persists raw webhook URLs.
type Verifier struct {
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

// Verify checks if the detected Teams webhook URL is valid/active.
// Raw contains the full webhook URL.
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	webhookURL := string(raw.Raw)
	if webhookURL == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty webhook URL",
		}
	}

	// Send a minimal message payload to verify the webhook is active.
	payload := []byte(`{"type":"message","text":""}`)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		slog.ErrorContext(ctx, "teams verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "teams verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		slog.InfoContext(ctx, "teams verifier: webhook is active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Teams webhook is active",
		}
	case http.StatusBadRequest:
		// 400 with a valid endpoint means the webhook exists but rejected our
		// empty payload — treat as active since the URL is reachable.
		slog.InfoContext(ctx, "teams verifier: webhook responded with bad request, treating as active")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Teams webhook is active (rejected empty payload)",
		}
	case http.StatusNotFound:
		slog.DebugContext(ctx, "teams verifier: webhook is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Teams webhook URL is not found or disabled",
		}
	default:
		slog.ErrorContext(ctx, "teams verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
