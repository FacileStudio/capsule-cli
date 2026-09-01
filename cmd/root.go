package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/ui"
)

var version = "dev"

var (
	flagJSON    bool
	flagNoColor bool
)

var rootCmd = &cobra.Command{
	Use:   "capsule",
	Short: "Zero-knowledge encrypted paste sharing",
	Long:  "Capsule encrypts your content client-side and shares it via a link. The server never sees plaintext.",
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Print one JSON document and nothing else")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	cobra.OnInitialize(func() {
		if flagNoColor || flagJSON {
			ui.DisableColor()
		}
	})
}

// Execute runs the root command tree.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
