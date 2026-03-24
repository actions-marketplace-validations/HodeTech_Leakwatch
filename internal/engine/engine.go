// Package engine provides the Leakwatch scan engine.
package engine

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/entropy"
	"github.com/cemililik/leakwatch/internal/matcher"
	"github.com/cemililik/leakwatch/internal/source"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const (
	// defaultEntropyThreshold is the default Shannon entropy threshold.
	defaultEntropyThreshold = 4.0

	// channelBufferMultiplier is the channel buffer size multiplier.
	channelBufferMultiplier = 2

	// hashTruncateLen is the Finding ID hash truncation length in bytes.
	// 16 bytes = 128 bits provides sufficient collision resistance.
	hashTruncateLen = 16
)

// Config holds the scan engine configuration.
type Config struct {
	Concurrency      int
	Detectors        []detector.Detector
	EnableEntropy    bool
	EntropyThreshold float64
	ShowRaw          bool
	Clock            func() time.Time // Optional; defaults to time.Now
	VerifierConfig   verifier.Config
	Verifiers        []verifier.Verifier
	OnlyVerified     bool             // If true, only return verified active findings
	MinSeverity      finding.Severity // Minimum severity to include in results
}

// ScanResult represents the outcome of a scan.
type ScanResult struct {
	Findings      []finding.Finding
	ScannedChunks int
	Duration      time.Duration
	Interrupted   bool
}

// Engine is the scan engine that orchestrates detection and verification.
type Engine struct {
	config   Config
	matcher  *matcher.Matcher
	verifyEn *verifier.Engine
}

// New creates a new scan engine.
// The Aho-Corasick automaton is compiled from detector keywords.
func New(cfg Config) *Engine {
	if cfg.Concurrency < 1 {
		cfg.Concurrency = 1
	}
	if cfg.EntropyThreshold <= 0 {
		cfg.EntropyThreshold = defaultEntropyThreshold
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	return &Engine{
		config:   cfg,
		matcher:  matcher.New(cfg.Detectors),
		verifyEn: verifier.NewEngine(cfg.VerifierConfig, cfg.Verifiers),
	}
}

// Scan scans the given source and returns results.
func (e *Engine) Scan(ctx context.Context, src source.Source) (*ScanResult, error) {
	if err := src.Validate(); err != nil {
		return nil, fmt.Errorf("source validation failed: %w", err)
	}

	start := time.Now()

	slog.Info("scan started",
		"source", src.Type(),
		"concurrency", e.config.Concurrency,
		"detectors", len(e.config.Detectors),
	)

	jobs := make(chan source.Chunk, e.config.Concurrency*channelBufferMultiplier)
	results := make(chan verifier.VerifyPair, e.config.Concurrency*channelBufferMultiplier)

	// Start workers.
	var wg sync.WaitGroup
	for i := 0; i < e.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.worker(ctx, jobs, results)
		}()
	}

	// Collect results (Finding + RawFinding pairs).
	var pairs []verifier.VerifyPair
	var collectWg sync.WaitGroup
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for p := range results {
			pairs = append(pairs, p)
		}
	}()

	// Send chunks to the jobs channel.
	// NOTE: Context cancellation during this loop depends on the source implementation
	// closing its Chunks channel promptly when ctx is cancelled. If a source blocks
	// indefinitely on send, this loop may not exit until the source returns.
	scannedChunks := 0
loop:
	for chunk := range src.Chunks(ctx) {
		select {
		case <-ctx.Done():
			break loop
		case jobs <- chunk:
			scannedChunks++
		}
	}
	close(jobs)

	// Wait for workers to finish
	wg.Wait()
	close(results)

	// Wait for collector to finish
	collectWg.Wait()

	// Run verification on collected pairs.
	findings := e.verifyEn.VerifyAll(ctx, pairs)

	// Apply post-scan filters.
	findings = e.applyFilters(findings)

	result := &ScanResult{
		Findings:      findings,
		ScannedChunks: scannedChunks,
		Duration:      time.Since(start),
		Interrupted:   ctx.Err() != nil,
	}

	slog.Info("scan completed",
		"findings", len(findings),
		"chunks", scannedChunks,
		"duration", result.Duration,
		"interrupted", result.Interrupted,
	)

	if ctx.Err() != nil {
		return result, fmt.Errorf("scan interrupted: %w", ctx.Err())
	}

	return result, nil
}

// worker reads chunks from the jobs channel, runs matched detectors, and sends
// VerifyPair results. It exits cleanly when the channel is closed or context is cancelled.
func (e *Engine) worker(ctx context.Context, jobs <-chan source.Chunk, results chan<- verifier.VerifyPair) {
	for chunk := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Aho-Corasick pre-filtering: only run matched detectors.
		matchedDetectors := e.matcher.Match(chunk.Data)

		for _, det := range matchedDetectors {
			select {
			case <-ctx.Done():
				return
			default:
			}

			rawFindings := det.Scan(ctx, chunk.Data)
			for _, raw := range rawFindings {
				f := e.rawToFinding(raw, chunk, det)
				pair := verifier.VerifyPair{Finding: f, Raw: raw}
				select {
				case <-ctx.Done():
					return
				case results <- pair:
				}
			}
		}
	}
}

// rawToFinding converts a raw detector finding to an enriched Finding.
// Generates a deterministic ID and optionally calculates entropy.
func (e *Engine) rawToFinding(raw detector.RawFinding, chunk source.Chunk, det detector.Detector) finding.Finding {
	f := finding.Finding{
		DetectorID:     det.ID(),
		Severity:       det.Severity(),
		Redacted:       raw.Redacted,
		SourceMetadata: chunk.SourceMetadata,
		Verification: finding.VerificationResult{
			Status: finding.StatusUnverified,
		},
		DetectedAt: e.config.Clock(),
		ExtraData:  raw.ExtraData,
	}

	if e.config.ShowRaw {
		f.Raw = string(raw.Raw)
	}

	if e.config.EnableEntropy && len(raw.Raw) > 0 {
		f.Entropy = entropy.Calculate(raw.Raw)
	}

	// Deterministik ID: detectorID + redacted + filePath
	hash := sha256.Sum256([]byte(det.ID() + raw.Redacted + chunk.SourceMetadata.FilePath))
	f.ID = fmt.Sprintf("%x", hash[:hashTruncateLen])

	return f
}

// applyFilters applies post-scan filters (severity, verification status).
func (e *Engine) applyFilters(findings []finding.Finding) []finding.Finding {
	var result []finding.Finding
	for _, f := range findings {
		if e.config.OnlyVerified && f.Verification.Status != finding.StatusVerifiedActive {
			continue
		}
		if f.Severity < e.config.MinSeverity {
			continue
		}
		result = append(result, f)
	}
	if result == nil {
		return []finding.Finding{}
	}
	return result
}
