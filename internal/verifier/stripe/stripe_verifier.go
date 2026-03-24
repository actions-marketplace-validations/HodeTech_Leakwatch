// Package stripe provides verifiers for Stripe API keys (live and test).
// It uses the Stripe Balance API GET /v1/balance endpoint with Basic auth
// to check key validity.
package stripe

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const (
	liveDetectorID = "stripe-api-key-live"
	testDetectorID = "stripe-api-key-test"
)

// defaultAPIURL is the base URL for the Stripe API.
const defaultAPIURL = "https://api.stripe.com"

// LiveKeyVerifier checks whether a Stripe live API key is active by calling
// the Stripe Balance API. It NEVER logs or persists raw key values.
type LiveKeyVerifier struct {
	// apiURL overrides the Stripe API base URL (for testing).
	apiURL string
	// httpClient overrides the default HTTP client (for testing).
	httpClient *http.Client
}

// TestKeyVerifier checks whether a Stripe test API key is active.
// It uses the same endpoint and logic as LiveKeyVerifier.
type TestKeyVerifier struct {
	// apiURL overrides the Stripe API base URL (for testing).
	apiURL string
	// httpClient overrides the default HTTP client (for testing).
	httpClient *http.Client
}

func init() {
	verifier.Register(&LiveKeyVerifier{})
	verifier.Register(&TestKeyVerifier{})
}

// Type returns the detector ID this verifier handles.
func (v *LiveKeyVerifier) Type() string {
	return liveDetectorID
}

// Type returns the detector ID this verifier handles.
func (v *TestKeyVerifier) Type() string {
	return testDetectorID
}

// Verify checks if the detected Stripe live API key is valid/active.
func (v *LiveKeyVerifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	return verifyStripeKey(ctx, v.apiURL, v.httpClient, raw, "live")
}

// Verify checks if the detected Stripe test API key is valid/active.
func (v *TestKeyVerifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	return verifyStripeKey(ctx, v.apiURL, v.httpClient, raw, "test")
}

// verifyStripeKey performs the actual Stripe API key verification.
func verifyStripeKey(ctx context.Context, apiURL string, httpClient *http.Client, raw detector.RawFinding, keyType string) finding.VerificationResult {
	token := string(raw.Raw)
	if token == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty token",
		}
	}

	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/v1/balance", nil)
	if err != nil {
		slog.ErrorContext(ctx, "stripe verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.SetBasicAuth(token, "")
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "stripe verifier: request failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.InfoContext(ctx, "stripe verifier: API key is active",
			slog.String("key_type", keyType),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: fmt.Sprintf("Stripe %s API key is active", keyType),
			ExtraData: map[string]string{
				"key_type": keyType,
			},
		}
	case http.StatusUnauthorized:
		slog.DebugContext(ctx, "stripe verifier: API key is inactive",
			slog.String("key_type", keyType),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: fmt.Sprintf("Stripe %s API key is invalid or revoked", keyType),
		}
	default:
		slog.ErrorContext(ctx, "stripe verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}
