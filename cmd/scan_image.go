package cmd

import (
	"runtime"

	"github.com/spf13/cobra"

	"github.com/cemililik/leakwatch/internal/source/container"
)

var scanImageCmd = &cobra.Command{
	Use:   "image <image:tag>",
	Short: "Scans a container image",
	Long: `Scans a container image layer by layer to detect leaked secrets.
Supports Docker Hub, GHCR, ECR, GCR and other OCI-compatible registries.
Does not require a running Docker daemon.`,
	Args: cobra.ExactArgs(1),
	RunE: runScanImage,
}

func init() {
	scanCmd.AddCommand(scanImageCmd)

	flags := scanImageCmd.Flags()
	flags.StringP("format", "f", "json", "output format (json, sarif, csv, table)")
	flags.StringP("output", "o", "", "output file (default: stdout)")
	flags.IntP("concurrency", "c", runtime.NumCPU(), "number of concurrent workers")
	flags.Int64("max-file-size", 10*1024*1024, "maximum file size in bytes")
	flags.Bool("show-raw", false, "show raw secret content in output")

	addVerifyFlags(flags)
	bindScanFlags(flags)
}

func runScanImage(cmd *cobra.Command, args []string) error {
	cfg, err := loadScanConfig(cmd)
	if err != nil {
		return err
	}

	src := container.New(args[0],
		container.WithMaxFileSize(cfg.maxFileSize),
	)

	return executeScan(cmd.Context(), cfg, src, nil)
}
