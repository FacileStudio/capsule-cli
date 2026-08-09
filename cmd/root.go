package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/ui"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "capsule",
	Short: "Zero-knowledge encrypted paste sharing",
	Long:  "Capsule encrypts your content client-side and shares it via a link. The server never sees plaintext.",
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	cobra.OnInitialize(func() {
		if v, _ := rootCmd.PersistentFlags().GetBool("no-color"); v {
			ui.DisableColor()
		}
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
