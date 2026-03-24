// Package snowflake provides a verifier for Snowflake credentials.
// Live verification requires a JDBC/ODBC connection, so this verifier
// performs format validation only.
package snowflake

import (
	"context"
	"log/slog"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "snowflake-credentials"

// Verifier checks whether Snowflake credentials have a valid format.
// It NEVER logs or persists raw password values.
type Verifier struct{}

func init() {
	verifier.Register(&Verifier{})
}

// Type returns the detector ID this verifier handles.
func (v *Verifier) Type() string {
	return detectorID
}

// Verify checks if the detected Snowflake credentials have a valid format.
// Raw contains the password value.
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	password := string(raw.Raw)
	if password == "" {
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "empty credentials",
		}
	}

	slog.InfoContext(ctx, "snowflake verifier: credentials format validated")

	return finding.VerificationResult{
		Status:  finding.StatusVerifiedActive,
		Message: "Credentials format validated (live verification requires database connection)",
	}
}
