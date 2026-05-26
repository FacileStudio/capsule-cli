package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/config"
)

var revokeToken string

func init() {
	revokeCmd.Flags().StringVar(&revokeToken, "token", "", "Delete token (required)")
	revokeCmd.MarkFlagRequired("token")
	rootCmd.AddCommand(revokeCmd)
}

var revokeCmd = &cobra.Command{
	Use:   "revoke <url>",
	Short: "Burn a capsule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]

		id, _, err := parseRevokeURL(rawURL)
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		client := api.NewClient(cfg.ServerURL)
		if err := client.Delete(id, revokeToken); err != nil {
			return fmt.Errorf("revoking: %w", err)
		}

		green := color.New(color.FgGreen)
		green.Println("Capsule burned.")
		return nil
	},
}

func parseRevokeURL(rawURL string) (string, string, error) {
	rawURL = strings.Split(rawURL, "#")[0]

	parts := strings.Split(rawURL, "/")
	if len(parts) == 0 {
		return "", "", fmt.Errorf("no paste ID found in URL")
	}

	id := parts[len(parts)-1]
	if id == "" {
		return "", "", fmt.Errorf("no paste ID found in URL")
	}

	return id, "", nil
}
