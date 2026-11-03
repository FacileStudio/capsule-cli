package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/config"
	"github.com/FacileStudio/capsule-cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(logoutCmd, whoamiCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Forget the stored token and end its session",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfigLeniently()
		if cfg.Token != "" {
			ctx, stop := signalContext()
			defer stop()
			if err := api.New(cfg.Server(), cfg.Token).Logout(ctx); err != nil {
				ui.Warn("the server did not end the session: %v", err)
			}
		}
		cfg.Token = ""
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		ui.Success("Logged out")
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the signed-in user",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := connectClient()
		if err != nil {
			return err
		}
		if client.Token == "" {
			return fmt.Errorf("not logged in: run 'capsule login'")
		}
		ctx, stop := signalContext()
		defer stop()
		me, err := client.Me(ctx)
		if err != nil {
			return err
		}
		if flagJSON {
			return ui.JSON(me)
		}
		ui.Plain("%s", me.Email)
		return nil
	},
}

func loadConfigLeniently() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		return &config.Config{ServerURL: config.DefaultServerURL}
	}
	return cfg
}
