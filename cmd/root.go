// Package cmd defines Leakwatch CLI commands.
// This package is a thin wiring layer; it must not contain business logic.
package cmd

import (
	"errors"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	logLevel string

	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

// SetVersionInfo sets build information (called from main.go).
func SetVersionInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

// FindingsExitError indicates that secrets were found (exit code 1).
type FindingsExitError struct {
	Count int
}

func (e *FindingsExitError) Error() string {
	return "secrets found"
}

var rootCmd = &cobra.Command{
	Use:   "leakwatch",
	Short: "Detects leaked secrets in codebases",
	Long: `Leakwatch detects, verifies, and reports leaked secrets (API keys, passwords,
certificates) in codebases, Git histories, and container images.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and returns the exit code.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		var fErr *FindingsExitError
		if errors.As(err, &fErr) {
			return 1
		}
		slog.Error("command failed", "error", err)
		return 2
	}
	return 0
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .leakwatch.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "log level (debug, info, warn, error)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".leakwatch")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
	}

	viper.SetEnvPrefix("LEAKWATCH")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// File not found is acceptable; parse errors are not.
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && viper.ConfigFileUsed() != "" {
			slog.Warn("failed to parse config file", "file", viper.ConfigFileUsed(), "error", err)
		}
	}
}

func initLogger() {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		slog.Warn("unknown log level, falling back to warn", "level", logLevel)
		level = slog.LevelWarn
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
