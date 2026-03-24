// Package rabbitmq provides a verifier for RabbitMQ connection strings.
// Live verification requires network access to the RabbitMQ server,
// so this verifier performs URL format validation only.
package rabbitmq

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "rabbitmq-connection-string"

// Verifier checks whether a RabbitMQ connection string has a valid URL format.
// It NEVER logs or persists raw password values.
type Verifier struct{}

func init() {
	verifier.Register(&Verifier{})
}

// Type returns the detector ID this verifier handles.
func (v *Verifier) Type() string {
	return detectorID
}

// Verify checks if the detected RabbitMQ connection string has a valid AMQP URL format.
// Raw contains the full connection string (amqp://user:pass@host:port/vhost).
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	connStr := string(raw.Raw)
	if connStr == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty connection string",
		}
	}

	parsed, err := url.Parse(connStr)
	if err != nil {
		slog.DebugContext(ctx, "rabbitmq verifier: failed to parse URL",
			slog.String("error", err.Error()),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "invalid URL format",
		}
	}

	if parsed.Scheme != "amqp" && parsed.Scheme != "amqps" {
		slog.DebugContext(ctx, "rabbitmq verifier: unexpected scheme",
			slog.String("scheme", parsed.Scheme),
		)
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "URL scheme is not amqp or amqps",
		}
	}

	if parsed.User == nil || parsed.User.Username() == "" {
		slog.DebugContext(ctx, "rabbitmq verifier: missing user credentials")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "missing user credentials in URL",
		}
	}

	if parsed.Host == "" {
		slog.DebugContext(ctx, "rabbitmq verifier: missing host")
		return finding.VerificationResult{
			Status:  finding.StatusVerifiedInactive,
			Message: "missing host in URL",
		}
	}

	extra := map[string]string{
		"host": parsed.Hostname(),
		"user": parsed.User.Username(),
	}

	slog.InfoContext(ctx, "rabbitmq verifier: connection string format validated",
		slog.String("host", parsed.Hostname()),
		slog.String("user", parsed.User.Username()),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "Connection string format validated (live verification requires network access)",
		ExtraData: extra,
	}
}
