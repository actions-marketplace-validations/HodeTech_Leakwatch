package engine

import (
	"context"
	"testing"
	"time"

	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/source"
	"github.com/cemililik/leakwatch/pkg/finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Source ---

type mockSource struct {
	chunks []source.Chunk
}

func (m *mockSource) Type() string    { return "mock" }
func (m *mockSource) Validate() error { return nil }
func (m *mockSource) Chunks(_ context.Context) <-chan source.Chunk {
	ch := make(chan source.Chunk, len(m.chunks))
	for _, c := range m.chunks {
		ch <- c
	}
	close(ch)
	return ch
}

// --- Mock Detector ---

type mockDetector struct {
	id       string
	keywords []string
	findings []detector.RawFinding
	severity finding.Severity
}

func (m *mockDetector) ID() string                                             { return m.id }
func (m *mockDetector) Description() string                                    { return "mock " + m.id }
func (m *mockDetector) Keywords() []string                                     { return m.keywords }
func (m *mockDetector) Severity() finding.Severity                             { return m.severity }
func (m *mockDetector) Scan(_ context.Context, _ []byte) []detector.RawFinding { return m.findings }

var fixedClock = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

// --- Tests ---

func TestScan_SingleChunkSingleDetector_ReturnsOneFinding(t *testing.T) {
	src := &mockSource{
		chunks: []source.Chunk{
			{
				Data:           []byte("AKIAIOSFODNN7TESTKEY1"),
				SourceMetadata: finding.SourceMetadata{SourceType: "filesystem", FilePath: "config.yaml"},
			},
		},
	}

	det := &mockDetector{
		id:       "test-detector",
		severity: finding.SeverityCritical,
		findings: []detector.RawFinding{
			{DetectorID: "test-detector", Raw: []byte("AKIAIOSFODNN7TESTKEY1"), Redacted: "AKIA****KEY1"},
		},
	}

	eng := New(Config{Concurrency: 2, Detectors: []detector.Detector{det}, Clock: fixedClock})

	result, err := eng.Scan(context.Background(), src)
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, "test-detector", result.Findings[0].DetectorID)
	assert.Equal(t, "AKIA****KEY1", result.Findings[0].Redacted)
	assert.Equal(t, finding.SeverityCritical, result.Findings[0].Severity)
	assert.NotEmpty(t, result.Findings[0].ID)
	assert.Empty(t, result.Findings[0].Raw)
	assert.Equal(t, 1, result.ScannedChunks)
	assert.False(t, result.Interrupted)
}

func TestScan_EmptySource_ReturnsNoFindings(t *testing.T) {
	src := &mockSource{chunks: nil}
	det := &mockDetector{id: "det", findings: nil}

	eng := New(Config{Concurrency: 1, Detectors: []detector.Detector{det}, Clock: fixedClock})

	result, err := eng.Scan(context.Background(), src)
	require.NoError(t, err)
	assert.Empty(t, result.Findings)
	assert.Equal(t, 0, result.ScannedChunks)
}

func TestScan_ShowRawEnabled_IncludesRawContent(t *testing.T) {
	src := &mockSource{
		chunks: []source.Chunk{
			{Data: []byte("secret"), SourceMetadata: finding.SourceMetadata{FilePath: "f.txt"}},
		},
	}

	det := &mockDetector{
		id:       "det",
		findings: []detector.RawFinding{{DetectorID: "det", Raw: []byte("secret-value"), Redacted: "sec***"}},
	}

	eng := New(Config{Concurrency: 1, Detectors: []detector.Detector{det}, ShowRaw: true, Clock: fixedClock})

	result, err := eng.Scan(context.Background(), src)
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, "secret-value", result.Findings[0].Raw)
}

func TestScan_EntropyEnabled_CalculatesEntropy(t *testing.T) {
	src := &mockSource{
		chunks: []source.Chunk{
			{Data: []byte("data"), SourceMetadata: finding.SourceMetadata{FilePath: "f.txt"}},
		},
	}

	det := &mockDetector{
		id:       "det",
		findings: []detector.RawFinding{{DetectorID: "det", Raw: []byte("aB3kL9mN2pQ7rT4xYz"), Redacted: "aB3k****T4xYz"}},
	}

	eng := New(Config{
		Concurrency:      1,
		Detectors:        []detector.Detector{det},
		EnableEntropy:    true,
		EntropyThreshold: 4.0,
		Clock:            fixedClock,
	})

	result, err := eng.Scan(context.Background(), src)
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Greater(t, result.Findings[0].Entropy, 0.0)
}

func TestScan_CancelledContext_ReturnsInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	src := &mockSource{
		chunks: []source.Chunk{
			{Data: []byte("data"), SourceMetadata: finding.SourceMetadata{FilePath: "f.txt"}},
		},
	}

	det := &mockDetector{id: "det", findings: nil}
	eng := New(Config{Concurrency: 1, Detectors: []detector.Detector{det}, Clock: fixedClock})

	result, err := eng.Scan(ctx, src)
	assert.Error(t, err)
	assert.True(t, result.Interrupted)
}

func TestScan_MultipleDetectors_ReturnsAllFindings(t *testing.T) {
	src := &mockSource{
		chunks: []source.Chunk{
			{Data: []byte("data"), SourceMetadata: finding.SourceMetadata{FilePath: "f.txt"}},
		},
	}

	det1 := &mockDetector{id: "det-1", findings: []detector.RawFinding{{DetectorID: "det-1", Raw: []byte("s1"), Redacted: "r1"}}}
	det2 := &mockDetector{id: "det-2", findings: []detector.RawFinding{{DetectorID: "det-2", Raw: []byte("s2"), Redacted: "r2"}}}

	eng := New(Config{Concurrency: 2, Detectors: []detector.Detector{det1, det2}, Clock: fixedClock})

	result, err := eng.Scan(context.Background(), src)
	require.NoError(t, err)
	assert.Len(t, result.Findings, 2)
}

func TestScan_SameInput_ReturnsDeterministicID(t *testing.T) {
	src := &mockSource{
		chunks: []source.Chunk{
			{Data: []byte("data"), SourceMetadata: finding.SourceMetadata{FilePath: "f.txt"}},
		},
	}

	det := &mockDetector{id: "det", findings: []detector.RawFinding{{DetectorID: "det", Raw: []byte("val"), Redacted: "v**"}}}
	eng := New(Config{Concurrency: 1, Detectors: []detector.Detector{det}, Clock: fixedClock})

	r1, _ := eng.Scan(context.Background(), src)
	r2, _ := eng.Scan(context.Background(), src)

	require.Len(t, r1.Findings, 1)
	require.Len(t, r2.Findings, 1)
	assert.Equal(t, r1.Findings[0].ID, r2.Findings[0].ID, "same input must produce same ID")
}
