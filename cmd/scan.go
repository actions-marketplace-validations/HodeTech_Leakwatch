package cmd

import "github.com/spf13/cobra"

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Starts a secret scan",
	Long:  `Sub-commands for scanning filesystem, Git repository, or container image.`,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
