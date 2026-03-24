// Package privatekey, özel anahtar dedektörlerini sağlar.
package privatekey

import (
	"context"
	"regexp"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

var privateKeyPattern = regexp.MustCompile(`-----BEGIN\s+(RSA |OPENSSH |DSA |EC |PGP )?PRIVATE KEY( BLOCK)?-----`)

// Detector, özel anahtar (private key) dedektörü.
type Detector struct{}

func (d *Detector) ID() string          { return "private-key" }
func (d *Detector) Description() string  { return "Private Key (RSA, SSH, DSA, EC, PGP)" }
func (d *Detector) Keywords() []string {
	return []string{
		"-----BEGIN RSA PRIVATE",
		"-----BEGIN OPENSSH PRIVATE",
		"-----BEGIN DSA PRIVATE",
		"-----BEGIN EC PRIVATE",
		"-----BEGIN PGP PRIVATE",
		"-----BEGIN PRIVATE KEY",
	}
}
func (d *Detector) Severity() finding.Severity { return finding.SeverityCritical }

// Scan, verilen veriyi PEM formatındaki özel anahtar başlıklarına karşı tarar.
func (d *Detector) Scan(_ context.Context, data []byte) []detector.RawFinding {
	matches := privateKeyPattern.FindAll(data, -1)
	if len(matches) == 0 {
		return nil
	}

	findings := make([]detector.RawFinding, 0, len(matches))
	for _, match := range matches {
		// Private key header'ını redact et
		redacted := "-----BEGIN [REDACTED] PRIVATE KEY-----"
		findings = append(findings, detector.RawFinding{
			DetectorID: d.ID(),
			Raw:        match,
			Redacted:   redacted,
		})
	}
	return findings
}

func init() {
	detector.Register(&Detector{})
}
