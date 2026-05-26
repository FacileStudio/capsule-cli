package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/api"
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
		parsed, err := parseURL(args[0])
		if err != nil {
			return err
		}

		client := api.NewClient(parsed.ServerURL)
		if err := client.Delete(parsed.ID, revokeToken); err != nil {
			return fmt.Errorf("revoking: %w", err)
		}

		green := color.New(color.FgGreen)
		green.Println("Capsule burned.")
		return nil
	},
}
