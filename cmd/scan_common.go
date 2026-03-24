package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cemililik/leakwatch/internal/config"
	"github.com/cemililik/leakwatch/internal/detector"
	"github.com/cemililik/leakwatch/internal/engine"
	"github.com/cemililik/leakwatch/internal/output"
	csvout "github.com/cemililik/leakwatch/internal/output/csv"
	jsonout "github.com/cemililik/leakwatch/internal/output/json"
	sarifout "github.com/cemililik/leakwatch/internal/output/sarif"
	tableout "github.com/cemililik/leakwatch/internal/output/table"
	"github.com/cemililik/leakwatch/internal/source"
	"github.com/cemililik/leakwatch/internal/verifier"
	"github.com/cemililik/leakwatch/pkg/finding"
)

// closeable is implemented by sources that hold resources (e.g. cloned repos).
type closeable interface {
	Close() error
}

// scanConfig holds the resolved configuration for a scan command.
type scanConfig struct {
	concurrency      int
	maxFileSize      int64
	excludePaths     []string
	enableEntropy    bool
	entropyThreshold float64
	showRaw          bool
	outputFile       string
	format           string
	noVerify         bool
	onlyVerified     bool
	minSeverity      finding.Severity
}

// bindScanFlags binds common scan flags to Viper.
func bindScanFlags(flags *pflag.FlagSet) {
	_ = viper.BindPFlag("scan.concurrency", flags.Lookup("concurrency"))
	_ = viper.BindPFlag("scan.max-file-size", flags.Lookup("max-file-size"))
	_ = viper.BindPFlag("output.format", flags.Lookup("format"))
	_ = viper.BindPFlag("output.file", flags.Lookup("output"))
	_ = viper.BindPFlag("output.show-raw", flags.Lookup("show-raw"))
}

// addVerifyFlags adds --no-verify and --only-verified flags to a command.
func addVerifyFlags(flags *pflag.FlagSet) {
	flags.Bool("no-verify", false, "disable secret verification")
	flags.Bool("only-verified", false, "only show verified active findings")
	flags.String("min-severity", "low", "minimum severity to report (low, medium, high, critical)")
}

// loadScanConfig loads and validates configuration from Viper.
func loadScanConfig(cmd *cobra.Command) (*scanConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	noVerify, _ := cmd.Flags().GetBool("no-verify")
	onlyVerified, _ := cmd.Flags().GetBool("only-verified")
	minSevStr, _ := cmd.Flags().GetString("min-severity")
	minSev := parseSeverity(minSevStr)

	// Read format and output directly from flags (overrides config file).
	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		format = cfg.Output.Format
	}
	outputFile, _ := cmd.Flags().GetString("output")
	if outputFile == "" {
		outputFile = cfg.Output.File
	}
	showRaw, _ := cmd.Flags().GetBool("show-raw")

	return &scanConfig{
		concurrency:      cfg.Scan.Concurrency,
		maxFileSize:      cfg.Scan.MaxFileSize,
		excludePaths:     cfg.Filter.ExcludePaths,
		enableEntropy:    cfg.Detection.Entropy.Enabled,
		entropyThreshold: cfg.Detection.Entropy.Threshold,
		showRaw:          showRaw || cfg.Output.ShowRaw,
		outputFile:       outputFile,
		format:           format,
		noVerify:         noVerify,
		onlyVerified:     onlyVerified,
		minSeverity:      minSev,
	}, nil
}

// selectFormatter returns the appropriate formatter based on the format string.
func selectFormatter(format string, showRaw bool) output.Formatter {
	switch format {
	case "sarif":
		return &sarifout.Formatter{ShowRaw: showRaw}
	case "csv":
		return &csvout.Formatter{ShowRaw: showRaw}
	case "table":
		return &tableout.Formatter{ShowRaw: showRaw}
	default:
		return &jsonout.Formatter{ShowRaw: showRaw}
	}
}

// executeScan runs the scan pipeline: detect, verify, filter, format, output.
// If cl is non-nil, Close() is called when the scan completes.
func executeScan(parent context.Context, cfg *scanConfig, src source.Source, cl closeable) error {
	if cl != nil {
		defer func() {
			if err := cl.Close(); err != nil {
				slog.Warn("failed to clean up source", "error", err)
			}
		}()
	}

	detectors := detector.All()
	if len(detectors) == 0 {
		return fmt.Errorf("no registered detectors found")
	}
	slog.Debug("detectors loaded", "count", len(detectors))

	// Configure verification.
	verifierCfg := verifier.DefaultConfig()
	if cfg.noVerify {
		verifierCfg.Enabled = false
	}

	eng := engine.New(engine.Config{
		Concurrency:      cfg.concurrency,
		Detectors:        detectors,
		EnableEntropy:    cfg.enableEntropy,
		EntropyThreshold: cfg.entropyThreshold,
		ShowRaw:          cfg.showRaw,
		VerifierConfig:   verifierCfg,
		Verifiers:        verifier.All(),
		OnlyVerified:     cfg.onlyVerified,
	})

	ctx, cancel := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	result, err := eng.Scan(ctx, src)
	if err != nil && result == nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Apply severity filter.
	if cfg.minSeverity > finding.SeverityLow {
		var filtered []finding.Finding
		for _, f := range result.Findings {
			if f.Severity >= cfg.minSeverity {
				filtered = append(filtered, f)
			}
		}
		result.Findings = filtered
	}

	// Write output.
	formatter := selectFormatter(cfg.format, cfg.showRaw)

	var w io.WriteCloser
	if cfg.outputFile != "" {
		cleanPath := filepath.Clean(cfg.outputFile)
		w, err = os.Create(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer w.Close()
	} else {
		w = os.Stdout
	}

	if err := formatter.Format(w, result.Findings); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if len(result.Findings) > 0 {
		slog.Info("secrets found", "count", len(result.Findings))
		return &FindingsExitError{Count: len(result.Findings)}
	}

	return nil
}

func parseSeverity(s string) finding.Severity {
	switch s {
	case "critical":
		return finding.SeverityCritical
	case "high":
		return finding.SeverityHigh
	case "medium":
		return finding.SeverityMedium
	default:
		return finding.SeverityLow
	}
}
