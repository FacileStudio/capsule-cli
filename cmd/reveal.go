package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/crypto"
)

func init() {
	rootCmd.AddCommand(revealCmd)
}

var revealCmd = &cobra.Command{
	Use:   "reveal <url>",
	Short: "Decrypt and display a capsule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parsed, err := parseURL(args[0])
		if err != nil {
			return err
		}

		if parsed.Fragment == "" {
			return fmt.Errorf("no key fragment found in URL (missing #)")
		}

		client := api.NewClient(parsed.ServerURL)

		var key []byte
		if crypto.IsPasswordProtected(parsed.Fragment) {
			fmt.Fprint(os.Stderr, "Password: ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}
			fmt.Fprintln(os.Stderr)

			key, err = crypto.UnwrapKey(parsed.Fragment, string(pw))
			if err != nil {
				return fmt.Errorf("wrong password or corrupted key")
			}
		} else {
			key, err = crypto.FromBase64URL(parsed.Fragment)
			if err != nil {
				return fmt.Errorf("decoding key: %w", err)
			}
		}

		paste, err := client.GetContent(parsed.ID)
		if err != nil {
			return fmt.Errorf("fetching content: %w", err)
		}

		plaintext, err := crypto.Decrypt(paste.Content, key)
		if err != nil {
			return fmt.Errorf("decrypting: %w", err)
		}

		fmt.Print(string(plaintext))
		return nil
	},
}
