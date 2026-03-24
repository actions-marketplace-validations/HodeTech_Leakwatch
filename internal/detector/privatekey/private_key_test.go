package privatekey

import (
	"context"
	"testing"

	"github.com/cemililik/leakwatch/pkg/finding"
	"github.com/stretchr/testify/assert"
)

func TestDetector_Metadata(t *testing.T) {
	d := &Detector{}
	assert.Equal(t, "private-key", d.ID())
	assert.Equal(t, finding.SeverityCritical, d.Severity())
	assert.NotEmpty(t, d.Keywords())
}

func TestDetector_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "RSA private key",
			input:    "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
			expected: 1,
		},
		{
			name:     "OpenSSH private key",
			input:    "-----BEGIN OPENSSH PRIVATE KEY-----\nb3Blb...\n-----END OPENSSH PRIVATE KEY-----",
			expected: 1,
		},
		{
			name:     "EC private key",
			input:    "-----BEGIN EC PRIVATE KEY-----\nMHQC...\n-----END EC PRIVATE KEY-----",
			expected: 1,
		},
		{
			name:     "DSA private key",
			input:    "-----BEGIN DSA PRIVATE KEY-----\nMIIB...\n-----END DSA PRIVATE KEY-----",
			expected: 1,
		},
		{
			name:     "generic private key (PKCS8)",
			input:    "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
			expected: 1,
		},
		{
			name:     "PGP private key block",
			input:    "-----BEGIN PGP PRIVATE KEY BLOCK-----\nxcLY...",
			expected: 1,
		},
		{
			name:     "public key - no match",
			input:    "-----BEGIN PUBLIC KEY-----\nMIIB...\n-----END PUBLIC KEY-----",
			expected: 0,
		},
		{
			name:     "certificate - no match",
			input:    "-----BEGIN CERTIFICATE-----\nMIIE...\n-----END CERTIFICATE-----",
			expected: 0,
		},
		{
			name:     "plain text - no match",
			input:    "just some normal text",
			expected: 0,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
		{
			name:     "multiple keys",
			input:    "-----BEGIN RSA PRIVATE KEY-----\n...\n-----BEGIN EC PRIVATE KEY-----",
			expected: 2,
		},
	}

	d := &Detector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := d.Scan(context.Background(), []byte(tt.input))
			assert.Len(t, findings, tt.expected)
		})
	}
}
