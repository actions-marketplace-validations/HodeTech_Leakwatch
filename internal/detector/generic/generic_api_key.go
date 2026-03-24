// Package generic, genel sır dedektörlerini sağlar.
package generic

import (
	"context"
	"regexp"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/entropy"
	"github.com/cemililik/leakwatch/pkg/finding"
)

var apiKeyPattern = regexp.MustCompile(`(?i)(api[_\-]?key|api[_\-]?secret|secret[_\-]?key)\s*[:=]\s*['"]?([a-zA-Z0-9/+=\-_]{16,64})['"]?`)

// APIKeyDetector, genel API anahtarı dedektörü.
type APIKeyDetector struct{}

func (d *APIKeyDetector) ID() string          { return "generic-api-key" }
func (d *APIKeyDetector) Description() string  { return "Generic API Key" }
func (d *APIKeyDetector) Keywords() []string {
	return []string{
		"api_key", "api-key", "apikey",
		"api_secret", "api-secret", "apisecret",
		"secret_key", "secret-key", "secretkey",
	}
}
func (d *APIKeyDetector) Severity() finding.Severity { return finding.SeverityMedium }

// Scan, verilen veriyi genel API anahtarı kalıplarına karşı tarar.
// Regex eşleşmesinden sonra Shannon entropi kontrolü uygular;
// entropi değeri 3.0'ın altındaki eşleşmeler gerçek sır olmadığı kabul edilerek atlanır.
func (d *APIKeyDetector) Scan(_ context.Context, data []byte) []detector.RawFinding {
	matches := apiKeyPattern.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		return nil
	}

	findings := make([]detector.RawFinding, 0, len(matches))
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		value := match[2]

		// Düşük entropili değerleri atla — gerçek sır olma olasılığı düşük
		if entropy.Calculate(value) < 3.0 {
			continue
		}

		redacted := redactValue(value)
		findings = append(findings, detector.RawFinding{
			DetectorID: d.ID(),
			Raw:        value,
			Redacted:   redacted,
			ExtraData: map[string]string{
				"key_name": string(match[1]),
			},
		})
	}
	return findings
}

func redactValue(value []byte) string {
	if len(value) <= 8 {
		return "****"
	}
	return string(value[:4]) + "****" + string(value[len(value)-4:])
}

func init() {
	detector.Register(&APIKeyDetector{})
}
