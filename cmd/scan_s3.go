package cmd

import (
	"runtime"

	"github.com/spf13/cobra"

	s3source "github.com/cemililik/leakwatch/internal/source/s3"
)

var scanS3Cmd = &cobra.Command{
	Use:   "s3 <bucket>",
	Short: "Scans an AWS S3 bucket",
	Long: `Scans objects in the specified AWS S3 bucket to detect leaked secrets.
Uses the default AWS credential chain (env vars, shared config, IAM role).`,
	Args: cobra.ExactArgs(1),
	RunE: runScanS3,
}

func init() {
	scanCmd.AddCommand(scanS3Cmd)

	flags := scanS3Cmd.Flags()
	flags.StringP("format", "f", "json", "output format (json, sarif, csv, table)")
	flags.StringP("output", "o", "", "output file (default: stdout)")
	flags.IntP("concurrency", "c", runtime.NumCPU(), "number of concurrent workers")
	flags.Int64("max-file-size", 10*1024*1024, "maximum file size in bytes")
	flags.Bool("show-raw", false, "show raw secret content in output")
	flags.String("prefix", "", "scan only objects with this key prefix")
	flags.String("region", "", "AWS region (default: from AWS config)")

	addVerifyFlags(flags)
	bindScanFlags(flags)
}

func runScanS3(cmd *cobra.Command, args []string) error {
	cfg, err := loadScanConfig(cmd)
	if err != nil {
		return err
	}

	var opts []s3source.Option
	opts = append(opts, s3source.WithMaxFileSize(cfg.maxFileSize))

	if prefix, _ := cmd.Flags().GetString("prefix"); prefix != "" {
		opts = append(opts, s3source.WithPrefix(prefix))
	}

	if region, _ := cmd.Flags().GetString("region"); region != "" {
		opts = append(opts, s3source.WithRegion(region))
	}

	src := s3source.New(args[0], opts...)

	return executeScan(cmd.Context(), cfg, src, nil)
}
