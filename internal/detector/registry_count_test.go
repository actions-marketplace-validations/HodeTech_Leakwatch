package detector_test

// This golden test pins the number of compile-time registered detectors so that
// accidentally dropping a detector (or a duplicate ID silently shadowing one)
// is caught immediately. It lives in the external detector_test package so it
// can blank-import every detector subpackage without creating an import cycle
// (each subpackage imports the detector package under test).
//
// Counts measured from the codebase:
//   - 63 detectors registered at compile time via init() (detector.Register).
//   - 59 packages register statically; azure, github, slack and stripe each
//     register two detectors (59 + 4 = 63).
//   - 60 detector subpackages exist in total; the 60th, "custom", registers its
//     rules at runtime (detector.RegisterIfAbsent) and is therefore not part of
//     the compile-time count.
//
// If you add or remove a detector, update wantDetectorCount below and keep the
// blank-import block in sync with cmd/imports.go.

import (
	"testing"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/stretchr/testify/assert"

	_ "github.com/cemililik/leakwatch/internal/detector/airtable"
	_ "github.com/cemililik/leakwatch/internal/detector/anthropic"
	_ "github.com/cemililik/leakwatch/internal/detector/auth0"
	_ "github.com/cemililik/leakwatch/internal/detector/aws"
	_ "github.com/cemililik/leakwatch/internal/detector/azure"
	_ "github.com/cemililik/leakwatch/internal/detector/bitbucket"
	_ "github.com/cemililik/leakwatch/internal/detector/circleci"
	_ "github.com/cemililik/leakwatch/internal/detector/cloudflare"
	_ "github.com/cemililik/leakwatch/internal/detector/coinbase"
	_ "github.com/cemililik/leakwatch/internal/detector/databricks"
	_ "github.com/cemililik/leakwatch/internal/detector/datadog"
	_ "github.com/cemililik/leakwatch/internal/detector/dbconn"
	_ "github.com/cemililik/leakwatch/internal/detector/deepseek"
	_ "github.com/cemililik/leakwatch/internal/detector/digitalocean"
	_ "github.com/cemililik/leakwatch/internal/detector/discord"
	_ "github.com/cemililik/leakwatch/internal/detector/dockerhub"
	_ "github.com/cemililik/leakwatch/internal/detector/doppler"
	_ "github.com/cemililik/leakwatch/internal/detector/figma"
	_ "github.com/cemililik/leakwatch/internal/detector/ftp"
	_ "github.com/cemililik/leakwatch/internal/detector/gcp"
	_ "github.com/cemililik/leakwatch/internal/detector/generic"
	_ "github.com/cemililik/leakwatch/internal/detector/github"
	_ "github.com/cemililik/leakwatch/internal/detector/gitlab"
	_ "github.com/cemililik/leakwatch/internal/detector/grafana"
	_ "github.com/cemililik/leakwatch/internal/detector/heroku"
	_ "github.com/cemililik/leakwatch/internal/detector/huggingface"
	_ "github.com/cemililik/leakwatch/internal/detector/infura"
	_ "github.com/cemililik/leakwatch/internal/detector/jwt"
	_ "github.com/cemililik/leakwatch/internal/detector/launchdarkly"
	_ "github.com/cemililik/leakwatch/internal/detector/ldap"
	_ "github.com/cemililik/leakwatch/internal/detector/linear"
	_ "github.com/cemililik/leakwatch/internal/detector/mailgun"
	_ "github.com/cemililik/leakwatch/internal/detector/newrelic"
	_ "github.com/cemililik/leakwatch/internal/detector/notion"
	_ "github.com/cemililik/leakwatch/internal/detector/npm"
	_ "github.com/cemililik/leakwatch/internal/detector/okta"
	_ "github.com/cemililik/leakwatch/internal/detector/openai"
	_ "github.com/cemililik/leakwatch/internal/detector/pagerduty"
	_ "github.com/cemililik/leakwatch/internal/detector/postmark"
	_ "github.com/cemililik/leakwatch/internal/detector/privatekey"
	_ "github.com/cemililik/leakwatch/internal/detector/pypi"
	_ "github.com/cemililik/leakwatch/internal/detector/rabbitmq"
	_ "github.com/cemililik/leakwatch/internal/detector/redis"
	_ "github.com/cemililik/leakwatch/internal/detector/rubygems"
	_ "github.com/cemililik/leakwatch/internal/detector/sendgrid"
	_ "github.com/cemililik/leakwatch/internal/detector/sentry"
	_ "github.com/cemililik/leakwatch/internal/detector/shopify"
	_ "github.com/cemililik/leakwatch/internal/detector/slack"
	_ "github.com/cemililik/leakwatch/internal/detector/snowflake"
	_ "github.com/cemililik/leakwatch/internal/detector/snyk"
	_ "github.com/cemililik/leakwatch/internal/detector/sonarcloud"
	_ "github.com/cemililik/leakwatch/internal/detector/stripe"
	_ "github.com/cemililik/leakwatch/internal/detector/supabase"
	_ "github.com/cemililik/leakwatch/internal/detector/teams"
	_ "github.com/cemililik/leakwatch/internal/detector/telegram"
	_ "github.com/cemililik/leakwatch/internal/detector/terraform"
	_ "github.com/cemililik/leakwatch/internal/detector/twilio"
	_ "github.com/cemililik/leakwatch/internal/detector/vault"
	_ "github.com/cemililik/leakwatch/internal/detector/vercel"
)

// wantDetectorCount is the expected number of compile-time registered detectors.
const wantDetectorCount = 63

// registeredAtInit snapshots the registry right after every blank-imported
// detector package has run its init(), but before any test can mutate the
// global registry (the in-package registry_test.go calls detector.Reset()).
// Capturing here makes the golden assertion independent of test ordering.
var registeredAtInit []detector.Detector

func init() {
	registeredAtInit = detector.All()
}

func TestAll_RegisteredDetectorCount_MatchesGolden(t *testing.T) {
	assert.Len(t, registeredAtInit, wantDetectorCount,
		"compile-time registered detector count drifted; update wantDetectorCount and cmd/imports.go together")

	// Every registered detector must have a unique, non-empty ID.
	ids := make(map[string]bool, len(registeredAtInit))
	for _, d := range registeredAtInit {
		assert.NotEmpty(t, d.ID())
		assert.False(t, ids[d.ID()], "duplicate detector ID: %s", d.ID())
		ids[d.ID()] = true
	}
}
