package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/config"
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
		rawURL := args[0]

		id, fragment, err := parseURL(rawURL)
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		client := api.NewClient(cfg.ServerURL)

		var key []byte
		if crypto.IsPasswordProtected(fragment) {
			fmt.Fprint(os.Stderr, "Password: ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}
			fmt.Fprintln(os.Stderr)

			key, err = crypto.UnwrapKey(fragment, string(pw))
			if err != nil {
				return fmt.Errorf("wrong password or corrupted key")
			}
		} else {
			key, err = crypto.FromBase64URL(fragment)
			if err != nil {
				return fmt.Errorf("decoding key: %w", err)
			}
		}

		paste, err := client.GetContent(id)
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

func parseURL(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("parsing URL: %w", err)
	}

	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return "", "", fmt.Errorf("no paste ID found in URL")
	}

	fragment := u.Fragment
	if fragment == "" {
		return "", "", fmt.Errorf("no key fragment found in URL (missing #)")
	}

	return path, fragment, nil
}
