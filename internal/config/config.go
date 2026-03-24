// Package config provides Leakwatch configuration management.
package config

import (
	"fmt"
	"runtime"

	"github.com/spf13/viper"
)

// Supported output formats.
var validFormats = map[string]bool{
	"json": true,
}

// Config represents the complete application configuration.
type Config struct {
	Scan      ScanConfig      `mapstructure:"scan"`
	Detection DetectionConfig `mapstructure:"detection"`
	Filter    FilterConfig    `mapstructure:"filter"`
	Output    OutputConfig    `mapstructure:"output"`
}

// ScanConfig holds scan engine configuration.
type ScanConfig struct {
	Concurrency int   `mapstructure:"concurrency"`
	MaxFileSize int64 `mapstructure:"max-file-size"`
}

// DetectionConfig holds detection configuration.
type DetectionConfig struct {
	Entropy EntropyConfig `mapstructure:"entropy"`
}

// EntropyConfig holds entropy analysis configuration.
type EntropyConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	Threshold float64 `mapstructure:"threshold"`
}

// FilterConfig holds filtering configuration.
type FilterConfig struct {
	ExcludePaths []string `mapstructure:"exclude-paths"`
}

// OutputConfig holds output configuration.
type OutputConfig struct {
	Format  string `mapstructure:"format"`
	File    string `mapstructure:"file"`
	ShowRaw bool   `mapstructure:"show-raw"`
}

// setDefaults configures default values on the given Viper instance.
func setDefaults(v *viper.Viper) {
	v.SetDefault("scan.concurrency", runtime.NumCPU())
	v.SetDefault("scan.max-file-size", 10*1024*1024) // 10MB
	v.SetDefault("detection.entropy.enabled", true)
	v.SetDefault("detection.entropy.threshold", 4.0)
	v.SetDefault("output.format", "json")
	v.SetDefault("output.show-raw", false)
}

// Load reads configuration from the global Viper instance and returns a Config.
func Load() (*Config, error) {
	setDefaults(viper.GetViper())
	return unmarshalAndValidate(viper.GetViper())
}

// LoadFrom reads configuration from a specific Viper instance.
// This is useful for testing with isolated Viper state.
func LoadFrom(v *viper.Viper) (*Config, error) {
	setDefaults(v)
	return unmarshalAndValidate(v)
}

func unmarshalAndValidate(v *viper.Viper) (*Config, error) {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Scan.Concurrency < 1 {
		return fmt.Errorf("invalid concurrency value: %d", c.Scan.Concurrency)
	}
	if c.Scan.MaxFileSize < 1 {
		return fmt.Errorf("invalid max-file-size value: %d", c.Scan.MaxFileSize)
	}
	if !validFormats[c.Output.Format] {
		return fmt.Errorf("unsupported output format: %s", c.Output.Format)
	}
	if c.Detection.Entropy.Threshold < 0 || c.Detection.Entropy.Threshold > 8.0 {
		return fmt.Errorf("invalid entropy threshold: %.2f (must be 0-8)", c.Detection.Entropy.Threshold)
	}
	return nil
}
