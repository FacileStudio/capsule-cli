package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "capsule",
	Short: "Zero-knowledge encrypted paste sharing",
	Long:  "Capsule encrypts your content client-side and shares it via a link. The server never sees plaintext.",
}

func init() {
	rootCmd.Version = version
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
