package config

import (
	"runtime"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestViper() *viper.Viper {
	return viper.New()
}

func TestLoadFrom_NoOverrides_ReturnsDefaults(t *testing.T) {
	v := newTestViper()

	cfg, err := LoadFrom(v)
	require.NoError(t, err)

	assert.Equal(t, runtime.NumCPU(), cfg.Scan.Concurrency)
	assert.Equal(t, int64(10*1024*1024), cfg.Scan.MaxFileSize)
	assert.True(t, cfg.Detection.Entropy.Enabled)
	assert.Equal(t, 4.0, cfg.Detection.Entropy.Threshold)
	assert.Equal(t, "json", cfg.Output.Format)
	assert.False(t, cfg.Output.ShowRaw)
}

func TestLoadFrom_CustomValues_OverridesDefaults(t *testing.T) {
	v := newTestViper()
	v.Set("scan.concurrency", 4)
	v.Set("scan.max-file-size", 5*1024*1024)

	cfg, err := LoadFrom(v)
	require.NoError(t, err)

	assert.Equal(t, 4, cfg.Scan.Concurrency)
	assert.Equal(t, int64(5*1024*1024), cfg.Scan.MaxFileSize)
}

func TestLoadFrom_InvalidConcurrency_ReturnsError(t *testing.T) {
	v := newTestViper()
	v.Set("scan.concurrency", 0)

	_, err := LoadFrom(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "concurrency")
}

func TestLoadFrom_InvalidMaxFileSize_ReturnsError(t *testing.T) {
	v := newTestViper()
	v.Set("scan.max-file-size", -1)

	_, err := LoadFrom(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max-file-size")
}

func TestLoadFrom_UnsupportedFormat_ReturnsError(t *testing.T) {
	v := newTestViper()
	v.Set("output.format", "xml")

	_, err := LoadFrom(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format")
}

func TestLoadFrom_InvalidEntropyThreshold_ReturnsError(t *testing.T) {
	v := newTestViper()
	v.Set("detection.entropy.threshold", 9.0)

	_, err := LoadFrom(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entropy threshold")
}

func TestLoadFrom_NegativeEntropyThreshold_ReturnsError(t *testing.T) {
	v := newTestViper()
	v.Set("detection.entropy.threshold", -1.0)

	_, err := LoadFrom(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entropy threshold")
}
