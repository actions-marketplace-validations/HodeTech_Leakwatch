// Package gcp provides a GCP Service Account Key secret detector.
package gcp

import (
	"context"
	"regexp"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

var (
	gcpServiceAccountPattern = regexp.MustCompile(`"type"\s*:\s*"service_account"`)
	privateKeyIDPattern      = regexp.MustCompile(`"private_key_id"\s*:\s*"([^"]+)"`)
	clientEmailPattern       = regexp.MustCompile(`"client_email"\s*:\s*"([^"]+)"`)
)

// Detector detects GCP Service Account Key files.
type Detector struct{}

// ID returns the unique identifier of the GCP Service Account detector.
func (d *Detector) ID() string { return "gcp-service-account" }

// Description returns a human-readable description of the GCP Service Account detector.
func (d *Detector) Description() string { return "GCP Service Account Key" }

// Keywords returns the Aho-Corasick pre-filter keywords for GCP Service Account detection.
func (d *Detector) Keywords() []string {
	return []string{"service_account", "private_key_id", "client_email"}
}

// Severity returns the default severity level for GCP Service Account findings.
func (d *Detector) Severity() finding.Severity { return finding.SeverityCritical }

// Scan searches the data for GCP Service Account Key JSON markers. When found,
// it extracts the private_key_id as Raw and uses the client_email for redaction.
func (d *Detector) Scan(_ context.Context, data []byte) []detector.RawFinding {
	matches := gcpServiceAccountPattern.FindAllIndex(data, -1)
	if len(matches) == 0 {
		return nil
	}

	findings := make([]detector.RawFinding, 0, len(matches))
	for range matches {
		raw := extractSubmatch(privateKeyIDPattern, data)
		email := extractSubmatch(clientEmailPattern, data)

		redacted := "GCP Service Account Key ****"
		if len(email) > 0 {
			redacted = "****@*.iam.gserviceaccount.com"
		}

		extra := make(map[string]string)
		if len(email) > 0 {
			extra["client_email"] = string(email)
		}
		if len(raw) > 0 {
			extra["private_key_id"] = string(raw)
		}

		f := detector.RawFinding{
			DetectorID: d.ID(),
			Raw:        raw,
			RawV2:      data,
			Redacted:   redacted,
			ExtraData:  extra,
		}
		if len(raw) == 0 {
			f.Raw = []byte("service_account")
		}

		findings = append(findings, f)
	}
	return findings
}

// extractSubmatch returns the first capture group from the given pattern, or nil.
func extractSubmatch(pattern *regexp.Regexp, data []byte) []byte {
	groups := pattern.FindSubmatch(data)
	if len(groups) < 2 {
		return nil
	}
	return groups[1]
}

func init() {
	detector.Register(&Detector{})
}
