// Package pagerduty provides a PagerDuty API Key secret detector.
package pagerduty

import (
	"context"
	"regexp"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

var pagerdutyKeyPattern = regexp.MustCompile(`u\+[A-Za-z0-9_-]{17,}`)

// Detector detects PagerDuty API Keys.
type Detector struct{}

// ID returns the unique identifier of the PagerDuty API Key detector.
func (d *Detector) ID() string { return "pagerduty-api-key" }

// Description returns a human-readable description of the PagerDuty API Key detector.
func (d *Detector) Description() string { return "PagerDuty API Key" }

// Keywords returns the Aho-Corasick pre-filter keywords for PagerDuty API Key detection.
func (d *Detector) Keywords() []string { return []string{"pagerduty", "PAGERDUTY", "u+"} }

// Severity returns the default severity level for PagerDuty API Key findings.
func (d *Detector) Severity() finding.Severity { return finding.SeverityHigh }

// Scan scans the given data for PagerDuty API Key patterns.
func (d *Detector) Scan(_ context.Context, data []byte) []detector.RawFinding {
	matches := pagerdutyKeyPattern.FindAll(data, -1)
	if len(matches) == 0 {
		return nil
	}

	findings := make([]detector.RawFinding, 0, len(matches))
	for _, match := range matches {
		findings = append(findings, detector.RawFinding{
			DetectorID: d.ID(),
			Raw:        match,
			Redacted:   "u+****" + string(match[len(match)-4:]),
		})
	}
	return findings
}

func init() {
	detector.Register(&Detector{})
}
