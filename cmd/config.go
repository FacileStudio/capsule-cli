package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/ui"

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

		fmt.Print(ui.Dim("server_url: "))
		fmt.Println(cfg.ServerURL)

		path, err := config.Path()
		if err == nil {
			fmt.Printf("\n%s\n", ui.Dim(fmt.Sprintf("Config file: %s", path)))
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

		ui.Success("Server URL set to %s", cfg.ServerURL)
		return nil
	},
}
