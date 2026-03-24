// Package telegram provides a Telegram Bot Token secret detector.
package telegram

import (
	"context"
	"regexp"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

var telegramTokenPattern = regexp.MustCompile(`[0-9]{7,10}:[A-Za-z0-9_-]{35}`)

// Detector detects Telegram Bot Tokens.
type Detector struct{}

// ID returns the unique identifier of the Telegram Bot Token detector.
func (d *Detector) ID() string { return "telegram-bot-token" }

// Description returns a human-readable description of the Telegram Bot Token detector.
func (d *Detector) Description() string { return "Telegram Bot Token" }

// Keywords returns the Aho-Corasick pre-filter keywords for Telegram Bot Token detection.
func (d *Detector) Keywords() []string {
	return []string{"telegram", "TELEGRAM", "bot_token", "BOT_TOKEN"}
}

// Severity returns the default severity level for Telegram Bot Token findings.
func (d *Detector) Severity() finding.Severity { return finding.SeverityHigh }

// Scan searches the data for Telegram Bot Token patterns.
func (d *Detector) Scan(_ context.Context, data []byte) []detector.RawFinding {
	matches := telegramTokenPattern.FindAll(data, -1)
	if len(matches) == 0 {
		return nil
	}

	findings := make([]detector.RawFinding, 0, len(matches))
	for _, match := range matches {
		findings = append(findings, detector.RawFinding{
			DetectorID: d.ID(),
			Raw:        match,
			Redacted:   string(match[:6]) + "****",
		})
	}
	return findings
}

func init() {
	detector.Register(&Detector{})
}
