package cmd

import (
	"testing"

	"github.com/HodeTech/leakwatch/internal/detector"
	"github.com/HodeTech/leakwatch/internal/meta"
	"github.com/HodeTech/leakwatch/internal/verifier"
	"github.com/stretchr/testify/assert"
)

// detectorsAtInit and verifiersAtInit snapshot the registries right after every
// package blank-imported by imports.go has run its init(), before any test can
// mutate the global registries. Capturing here makes the guard below
// independent of test ordering.
var (
	detectorsAtInit []detector.Detector
	verifiersAtInit []verifier.Verifier
)

func init() {
	detectorsAtInit = detector.All()
	verifiersAtInit = verifier.All()
}

// TestMetaCounts_MatchRuntime guards the published counts in internal/meta
// against what the binary actually registers. Every detector and verifier
// package is blank-imported by imports.go in this package, so both registries
// are fully populated here (the detector-only test in internal/detector cannot
// see verifiers, hence the cross-check lives here).
func TestMetaCounts_MatchRuntime(t *testing.T) {
	assert.Len(t, detectorsAtInit, meta.Detectors,
		"meta.Detectors drifted from detector.All(); update internal/meta then run `go generate ./...`")
	assert.Len(t, verifiersAtInit, meta.Verifiers,
		"meta.Verifiers drifted from verifier.All(); update internal/meta then run `go generate ./...`")
}
