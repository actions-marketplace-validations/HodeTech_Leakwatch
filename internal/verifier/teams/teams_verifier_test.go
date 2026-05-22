package teams

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

// TestVerify_ProbeIsNonDestructive asserts that the verifier never POSTs a
// renderable message: the probe body must be an empty JSON object with no
// "text" or "summary" field, so Teams cannot deliver a card.
func TestVerify_ProbeIsNonDestructive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "{}", string(body))
		assert.NotContains(t, string(body), "text")
		assert.NotContains(t, string(body), "summary")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`Bad payload`))
	}))
	defer server.Close()

	v := &Verifier{httpClient: server.Client()}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(server.URL),
		Redacted:   "https://outlook.office.com/webhook/****",
	}

	result := v.Verify(context.Background(), raw)

	require.Equal(t, finding.StatusVerifiedActive, result.Status)
}

func TestVerify_BadPayloadRejected_ReturnsActive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`Bad payload`))
	}))
	defer server.Close()

	v := &Verifier{httpClient: server.Client()}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(server.URL),
		Redacted:   "https://outlook.office.com/webhook/****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedActive, result.Status)
	assert.Contains(t, result.Message, "rejected non-destructive empty payload")
}

func TestVerify_NotFound_ReturnsInactive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	v := &Verifier{httpClient: server.Client()}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(server.URL),
		Redacted:   "https://outlook.office.com/webhook/****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
	assert.Equal(t, "Teams webhook URL is not found or disabled", result.Message)
}

// TestVerify_OKStatus_ReturnsUnverified verifies that a 2xx response (which a
// genuine Teams webhook never returns for an empty payload) is treated as
// inconclusive rather than active.
func TestVerify_OKStatus_ReturnsUnverified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`1`))
	}))
	defer server.Close()

	v := &Verifier{httpClient: server.Client()}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(server.URL),
		Redacted:   "https://outlook.office.com/webhook/****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusUnverified, result.Status)
	assert.Contains(t, result.Message, "inconclusive")
}

func TestVerify_ServerError_ReturnsUnverified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	v := &Verifier{httpClient: server.Client()}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(server.URL),
		Redacted:   "https://outlook.office.com/webhook/****",
	}

	result := v.Verify(context.Background(), raw)

	// A 5xx is inconclusive for a non-destructive probe.
	assert.Equal(t, finding.StatusUnverified, result.Status)
	assert.Contains(t, result.Message, "inconclusive")
}

func TestVerify_Type_ReturnsCorrectID(t *testing.T) {
	v := &Verifier{}
	assert.Equal(t, "teams-webhook", v.Type())
}

func TestVerify_EmptyURL_ReturnsUnverified(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(""),
		Redacted:   "",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusUnverified, result.Status)
	assert.Equal(t, "empty webhook URL", result.Message)
}
