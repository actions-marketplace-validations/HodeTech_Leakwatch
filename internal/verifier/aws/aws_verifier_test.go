package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

// mockSTSClient implements stsClient for testing.
type mockSTSClient struct {
	output *sts.GetCallerIdentityOutput
	err    error
}

func (m *mockSTSClient) GetCallerIdentity(_ context.Context, _ *sts.GetCallerIdentityInput, _ ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return m.output, m.err
}

func TestVerify_NoSecretKey_ReturnsUnverified(t *testing.T) {
	v := &Verifier{}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("AKIAIOSFODNN7EXAMPLE"),
		RawV2:      nil,
		Redacted:   "AKIA****MPLE",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusUnverified, result.Status)
	assert.Equal(t, "secret access key not found", result.Message)
}

func TestVerify_Type_ReturnsCorrectID(t *testing.T) {
	v := &Verifier{}
	assert.Equal(t, "aws-access-key-id", v.Type())
}

func TestVerify_ValidCredentials_ReturnsActive(t *testing.T) {
	mock := &mockSTSClient{
		output: &sts.GetCallerIdentityOutput{
			Account: aws.String("123456789012"),
			Arn:     aws.String("arn:aws:iam::123456789012:user/testuser"),
			UserId:  aws.String("AIDAEXAMPLE"),
		},
	}
	v := &Verifier{client: mock}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("AKIAIOSFODNN7EXAMPLE"),
		RawV2:      []byte("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		Redacted:   "AKIA****MPLE",
	}

	result := v.Verify(context.Background(), raw)

	require.Equal(t, finding.StatusVerifiedActive, result.Status)
	assert.Equal(t, "AWS credentials are active", result.Message)
	assert.Equal(t, "123456789012", result.ExtraData["account"])
	assert.Equal(t, "arn:aws:iam::123456789012:user/testuser", result.ExtraData["arn"])
	assert.Equal(t, "AIDAEXAMPLE", result.ExtraData["user_id"])
}

func TestVerify_InvalidCredentials_ReturnsInactive(t *testing.T) {
	mock := &mockSTSClient{
		err: errors.New("operation error STS: GetCallerIdentity, InvalidClientTokenId: The security token included in the request is invalid"),
	}
	v := &Verifier{client: mock}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("AKIAIOSFODNN7EXAMPLE"),
		RawV2:      []byte("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		Redacted:   "AKIA****MPLE",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifiedInactive, result.Status)
}

func TestVerify_OtherError_ReturnsVerifyError(t *testing.T) {
	mock := &mockSTSClient{
		err: errors.New("network timeout"),
	}
	v := &Verifier{client: mock}

	raw := detector.RawFinding{
		DetectorID: detectorID,
		Raw:        []byte("AKIAIOSFODNN7EXAMPLE"),
		RawV2:      []byte("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		Redacted:   "AKIA****MPLE",
	}

	result := v.Verify(context.Background(), raw)

	assert.Equal(t, finding.StatusVerifyError, result.Status)
	assert.Contains(t, result.Message, "network timeout")
}
