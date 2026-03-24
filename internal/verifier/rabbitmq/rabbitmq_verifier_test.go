package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

func TestVerifier_Type_ReturnsCorrectID(t *testing.T) {
	v := &Verifier{}
	assert.Equal(t, "rabbitmq-connection-string", v.Type())
}

func TestVerify_ValidAMQPURL_ReturnsActive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("amqp://admin:s3cret@rabbitmq.example.com:5672/"),
		Redacted:   "amqp://admin:****@rabbitmq.example.com:5672/",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedActive, result.Status)
	assert.Equal(t, "Connection string format validated (live verification requires network access)", result.Message)
	assert.Equal(t, "rabbitmq.example.com", result.ExtraData["host"])
	assert.Equal(t, "admin", result.ExtraData["user"])
}

func TestVerify_ValidAMQPSURL_ReturnsActive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("amqps://myuser:mypass@broker.cloud.io:5671/production"),
		Redacted:   "amqps://myuser:****@broker.cloud.io:5671/production",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedActive, result.Status)
	assert.Equal(t, "broker.cloud.io", result.ExtraData["host"])
	assert.Equal(t, "myuser", result.ExtraData["user"])
}

func TestVerify_WrongScheme_ReturnsInactive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("http://admin:pass@rabbitmq.example.com:5672/"),
		Redacted:   "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
	assert.Equal(t, "URL scheme is not amqp or amqps", result.Message)
}

func TestVerify_MissingCredentials_ReturnsInactive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("amqp://rabbitmq.example.com:5672/"),
		Redacted:   "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
	assert.Equal(t, "missing user credentials in URL", result.Message)
}

func TestVerify_MissingHost_ReturnsInactive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("amqp://user:pass@"),
		Redacted:   "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
	assert.Equal(t, "missing host in URL", result.Message)
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
	assert.Equal(t, "empty connection string", result.Message)
}

func TestVerify_InvalidURL_ReturnsInactive(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("://not-a-valid-url"),
		Redacted:   "****",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
}
