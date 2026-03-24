// Package shopify provides a verifier for Shopify access tokens.
// It uses the Shopify Admin API GET /admin/api/2024-01/shop.json endpoint
// to check token validity. A store domain is required via ExtraData.
package shopify

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

const detectorID = "shopify-access-token"

// shopAPIPath is the Shopify Admin API endpoint for shop details.
const shopAPIPath = "/admin/api/2024-01/shop.json"

// Verifier checks whether a Shopify access token is active by calling the
// Shopify Admin API. It NEVER logs or persists raw token values.
type Verifier struct {
	// apiURL overrides the full base URL (for testing). When set, store_domain
	// from ExtraData is ignored and apiURL is used directly.
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

// Verify checks if the detected Shopify access token is valid/active.
// Raw contains the token value. ExtraData["store_domain"] must contain the
// Shopify store domain (e.g., "mystore.myshopify.com").
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	token := string(raw.Raw)
	if token == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty token",
		}
	}

	baseURL := v.apiURL
	if baseURL == "" {
		domain := raw.ExtraData["store_domain"]
		if domain == "" {
			slog.DebugContext(ctx, "shopify verifier: store domain not available, cannot verify")
			return finding.VerificationResult{
				Status:  finding.StatusUnverified,
				Message: "store domain required for verification",
			}
		}
		baseURL = "https://" + domain
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+shopAPIPath, nil)
	if err != nil {
		slog.ErrorContext(ctx, "shopify verifier: failed to create request", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("X-Shopify-Access-Token", token)
	req.Header.Set("User-Agent", "leakwatch-verifier")

	client := v.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "shopify verifier: request failed", slog.String("error", err.Error()))
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
		slog.DebugContext(ctx, "shopify verifier: access token is inactive")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "Shopify access token is invalid or revoked",
		}
	default:
		slog.ErrorContext(ctx, "shopify verifier: unexpected status code",
			slog.Int("status_code", resp.StatusCode),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// handleActiveToken parses the Shopify API response for a valid token.
func handleActiveToken(ctx context.Context, body io.Reader) finding.VerificationResult {
	var shopResp struct {
		Shop struct {
			Name string `json:"name"`
		} `json:"shop"`
	}

	if err := json.NewDecoder(body).Decode(&shopResp); err != nil {
		slog.ErrorContext(ctx, "shopify verifier: failed to decode response", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedActive,
			Message: "Shopify access token is active (could not parse shop details)",
		}
	}

	extra := map[string]string{}
	if shopResp.Shop.Name != "" {
		extra["shop_name"] = shopResp.Shop.Name
	}

	slog.InfoContext(ctx, "shopify verifier: access token is active",
		slog.String("shop_name", shopResp.Shop.Name),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Shopify access token is active",
		ExtraData: extra,
	}
}
