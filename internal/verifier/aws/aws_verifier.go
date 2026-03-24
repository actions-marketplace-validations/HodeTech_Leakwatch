// Package aws provides a verifier for AWS Access Key credentials.
// It uses AWS STS GetCallerIdentity to check whether a key pair is active.
package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const detectorID = "aws-access-key-id"

// stsClient abstracts the STS API for testing.
type stsClient interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// Verifier checks whether an AWS access key pair is active by calling
// STS GetCallerIdentity. It NEVER logs or persists raw key values.
type Verifier struct {
	client stsClient
}

func init() {
	verifier.Register(&Verifier{})
}

// Type returns the detector ID this verifier handles.
func (v *Verifier) Type() string {
	return detectorID
}

// Verify checks if the detected AWS access key is valid/active.
// Raw contains the Access Key ID and RawV2 contains the Secret Access Key.
// Both are required for verification.
func (v *Verifier) Verify(ctx context.Context, raw detector.RawFinding) finding.VerificationResult {
	if len(raw.RawV2) == 0 {
		slog.DebugContext(ctx, "aws verifier: secret access key not found, skipping verification")
		return finding.VerificationResult{
			Status:  finding.StatusUnverified,
			Message: "secret access key not found",
		}
	}

	client := v.client
	if client == nil {
		client = sts.New(sts.Options{
			Region: "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider(
				string(raw.Raw),
				string(raw.RawV2),
				"",
			),
		})
	}

	output, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		// Check if this is an authentication error (invalid credentials).
		// AWS returns specific error codes for invalid/expired keys.
		if isAuthError(err) {
			slog.DebugContext(ctx, "aws verifier: credentials are inactive")
			return finding.VerificationResult{
				Status:  finding.StatusVerifiedInactive,
				Message: "AWS credentials are invalid or inactive",
			}
		}
		slog.ErrorContext(ctx, "aws verifier: verification failed", slog.String("error", err.Error()))
		return finding.VerificationResult{
			Status:  finding.StatusVerifyError,
			Message: fmt.Sprintf("verification failed: %v", err),
		}
	}

	extra := make(map[string]string)
	if output.Account != nil {
		extra["account"] = aws.ToString(output.Account)
	}
	if output.Arn != nil {
		extra["arn"] = aws.ToString(output.Arn)
	}
	if output.UserId != nil {
		extra["user_id"] = aws.ToString(output.UserId)
	}

	slog.InfoContext(ctx, "aws verifier: credentials are active",
		slog.String("account", extra["account"]),
		slog.String("arn", extra["arn"]),
	)

	return finding.VerificationResult{
		Status:    finding.StatusVerifiedActive,
		Message:   "AWS credentials are active",
		ExtraData: extra,
	}
}

// isAuthError checks whether the error indicates invalid credentials.
func isAuthError(err error) bool {
	// AWS SDK v2 wraps errors; check the error message for common auth failure indicators.
	errMsg := err.Error()
	authErrorPatterns := []string{
		"InvalidClientTokenId",
		"SignatureDoesNotMatch",
		"ExpiredToken",
		"AccessDenied",
	}
	for _, pattern := range authErrorPatterns {
		if containsString(errMsg, pattern) {
			return true
		}
	}
	return false
}

// containsString checks if s contains substr (simple string search).
func containsString(s, substr string) bool {
	return len(substr) <= len(s) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
