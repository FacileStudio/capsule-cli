package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/config"
)

func init() {
	configSetCmd.AddCommand(configSetServerCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or modify configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		dim := color.New(color.Faint)
		dim.Print("server_url: ")
		fmt.Println(cfg.ServerURL)

		path, err := config.Path()
		if err == nil {
			dim.Printf("\nConfig file: %s\n", path)
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
}

var configSetServerCmd = &cobra.Command{
	Use:   "server <url>",
	Short: "Set the server URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		cfg.ServerURL = args[0]
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		green := color.New(color.FgGreen)
		green.Printf("Server URL set to %s\n", cfg.ServerURL)
		return nil
	},
}
