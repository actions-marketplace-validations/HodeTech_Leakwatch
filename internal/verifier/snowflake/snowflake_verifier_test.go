package snowflake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

func TestVerifier_Type_ReturnsCorrectID(t *testing.T) {
	v := &Verifier{}
	assert.Equal(t, "snowflake-credentials", v.Type())
}

func TestVerify_ValidCredentials_ReturnsActive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("MyStr0ngP@ssword!"),
		Redacted:   "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedActive, result.Status)
	assert.Equal(t, "Credentials format validated (live verification requires database connection)", result.Message)
}

func TestVerify_EmptyInput_ReturnsUnverified(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte(""),
		Redacted:   "",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusUnverified, result.Status)
	assert.Equal(t, "empty credentials", result.Message)
}

func TestVerify_NonEmptyCredentials_ReturnsActive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("simple"),
		Redacted:   "****",
		ExtraData: map[string]string{
			"account": "xy12345.us-east-1",
			"user":    "ADMIN",
		},
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedActive, result.Status)
}
