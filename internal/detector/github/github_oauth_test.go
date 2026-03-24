package github

import (
	"context"
	"strings"
	"testing"

	"github.com/cemililik/leakwatch/pkg/finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthDetector_Metadata_ReturnsExpectedValues(t *testing.T) {
	d := &OAuthDetector{}
	assert.Equal(t, "github-oauth-token", d.ID())
	assert.Equal(t, "GitHub OAuth2 Token", d.Description())
	assert.Equal(t, finding.SeverityCritical, d.Severity())
	assert.NotEmpty(t, d.Keywords())
}

func TestOAuthDetector_Scan_MatchAndReject(t *testing.T) {
	// Synthetic 40-char suffix (above 36 minimum)
	suffix40 := strings.Repeat("Abc1D678", 5)

	tests := []struct {
		name     string
		input    string
		expected int
		redacted string
	}{
		{
			name:     "valid gho_ token",
			input:    "gho_" + suffix40,
			expected: 1,
			redacted: "gho_Abc1****",
		},
		{
			name:     "valid ghr_ token",
			input:    "ghr_" + suffix40,
			expected: 1,
			redacted: "ghr_Abc1****",
		},
		{
			name:     "valid ghu_ token",
			input:    "ghu_" + suffix40,
			expected: 1,
			redacted: "ghu_Abc1****",
		},
		{
			name:     "valid ghs_ token",
			input:    "ghs_" + suffix40,
			expected: 1,
			redacted: "ghs_Abc1****",
		},
		{
			name:     "token embedded in config",
			input:    `GITHUB_TOKEN=gho_` + suffix40,
			expected: 1,
		},
		{
			name:     "no match - suffix too short",
			input:    "gho_abc123",
			expected: 0,
		},
		{
			name:     "no match - wrong prefix ghp",
			input:    "ghp_" + suffix40,
			expected: 0,
		},
		{
			name:     "no match - wrong prefix ghx",
			input:    "ghx_" + suffix40,
			expected: 0,
		},
		{
			name:     "no match - plain text",
			input:    "this is just normal text",
			expected: 0,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
	}

	d := &OAuthDetector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := d.Scan(context.Background(), []byte(tt.input))
			assert.Len(t, findings, tt.expected)
			if tt.expected > 0 && tt.redacted != "" {
				require.NotEmpty(t, findings)
				assert.Equal(t, tt.redacted, findings[0].Redacted)
			}
		})
	}
}
