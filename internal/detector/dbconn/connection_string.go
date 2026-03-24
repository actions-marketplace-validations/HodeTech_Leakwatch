// Package dbconn provides a detector for database connection strings.
package dbconn

import (
	"context"
	"net/url"
	"regexp"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/pkg/finding"
)

var connStringPattern = regexp.MustCompile(`(postgres|mysql|mongodb(\+srv)?|redis)://[^\s'"]+@[^\s'"]+`)

// ConnectionString detects database connection strings containing credentials.
type ConnectionString struct{}

// ID returns the unique identifier of the database connection string detector.
func (d *ConnectionString) ID() string { return "database-connection-string" }

// Description returns a human-readable description of the database connection string detector.
func (d *ConnectionString) Description() string { return "Database Connection String" }

// Keywords returns the Aho-Corasick pre-filter keywords for database connection string detection.
func (d *ConnectionString) Keywords() []string {
	return []string{"postgres://", "mysql://", "mongodb://", "mongodb+srv://", "redis://"}
}

// Severity returns the default severity level for database connection string findings.
func (d *ConnectionString) Severity() finding.Severity { return finding.SeverityCritical }

// Scan scans the given data for database connection string patterns.
// The password portion of the URL is redacted in the finding output.
func (d *ConnectionString) Scan(_ context.Context, data []byte) []detector.RawFinding {
	matches := connStringPattern.FindAll(data, -1)
	if len(matches) == 0 {
		return nil
	}

	findings := make([]detector.RawFinding, 0, len(matches))
	for _, match := range matches {
		findings = append(findings, detector.RawFinding{
			DetectorID: d.ID(),
			Raw:        match,
			Redacted:   redactPassword(string(match)),
		})
	}
	return findings
}

// redactPassword masks the password portion in a database connection URL.
// Uses net/url.Parse for proper parsing, then reconstructs with masked password.
func redactPassword(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "****"
	}
	if u.User == nil {
		return raw // No credentials to redact
	}
	username := u.User.Username()
	// Reconstruct manually to avoid URL-encoding of **** characters.
	u.User = nil
	return u.Scheme + "://" + username + ":****@" + u.Host + u.RequestURI()
}

func init() {
	detector.Register(&ConnectionString{})
}
