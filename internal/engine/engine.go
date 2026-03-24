// Package engine, Leakwatch tarama motorunu sağlar.
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
	"github.com/cemililik/leakwatch/internal/source"
	"github.com/cemililik/leakwatch/pkg/finding"
)

const (
	// defaultEntropyThreshold, varsayılan Shannon entropi eşik değeri.
	defaultEntropyThreshold = 4.0

	// channelBufferMultiplier, kanal buffer boyutu çarpanı.
	channelBufferMultiplier = 2

	// hashTruncateLen, Finding ID hash kesme uzunluğu (byte).
	hashTruncateLen = 8
)

// Config, tarama motoru yapılandırması.
type Config struct {
	Concurrency      int
	Detectors        []detector.Detector
	EnableEntropy    bool
	EntropyThreshold float64
	ShowRaw          bool
	Clock            func() time.Time // Opsiyonel, nil ise time.Now kullanılır
}

// ScanResult, tarama sonucunu temsil eder.
type ScanResult struct {
	Findings      []finding.Finding
	ScannedChunks int
	Duration      time.Duration
	Interrupted   bool
}

// Engine, tarama motorunu temsil eder.
type Engine struct {
	config Config
}

// New, yeni bir tarama motoru oluşturur.
func New(cfg Config) *Engine {
	if cfg.Concurrency < 1 {
		cfg.Concurrency = 1
	}
	if cfg.EntropyThreshold == 0 {
		cfg.EntropyThreshold = defaultEntropyThreshold
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	return &Engine{config: cfg}
}

// Scan, verilen kaynağı tarar ve sonuçları döndürür.
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
	results := make(chan finding.Finding, e.config.Concurrency*channelBufferMultiplier)

	// Worker'ları başlat
	var wg sync.WaitGroup
	for i := 0; i < e.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.worker(ctx, jobs, results)
		}()
	}

	// Sonuçları topla
	var findings []finding.Finding
	var collectWg sync.WaitGroup
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for f := range results {
			findings = append(findings, f)
		}
	}()

	// Chunk'ları jobs kanalına gönder
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

	// Worker'ların bitmesini bekle
	wg.Wait()
	close(results)

	// Toplamanın bitmesini bekle
	collectWg.Wait()

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

// worker, jobs kanalından chunk okuyup dedektörlere gönderen işçi goroutine'idir.
// Channel kapatıldığında veya context iptal edildiğinde temiz bir şekilde çıkar.
func (e *Engine) worker(ctx context.Context, jobs <-chan source.Chunk, results chan<- finding.Finding) {
	for chunk := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, det := range e.config.Detectors {
			// Dedektörler arası context kontrolü
			select {
			case <-ctx.Done():
				return
			default:
			}

			rawFindings := det.Scan(ctx, chunk.Data)
			for _, raw := range rawFindings {
				f := e.rawToFinding(raw, chunk, det)
				select {
				case <-ctx.Done():
					return
				case results <- f:
				}
			}
		}
	}
}

// rawToFinding, ham dedektör bulgusunu zenginleştirilmiş Finding'e dönüştürür.
// Deterministik bir ID oluşturur ve opsiyonel olarak entropi hesaplar.
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
