package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/config"
	"github.com/FacileStudio/capsule-cli/internal/crypto"
)

var (
	sealBurn     bool
	sealNoBurn   bool
	sealExpires  string
	sealPassword bool
	sealSyntax   string
)

func init() {
	sealCmd.Flags().BoolVar(&sealBurn, "burn", true, "Burn after read")
	sealCmd.Flags().BoolVar(&sealNoBurn, "no-burn", false, "Do not burn after read")
	sealCmd.Flags().StringVarP(&sealExpires, "expires", "e", "24h", "Expiration (1h, 24h, 7d, 30d)")
	sealCmd.Flags().BoolVarP(&sealPassword, "password", "p", false, "Protect with a password")
	sealCmd.Flags().StringVarP(&sealSyntax, "syntax", "s", "", "Syntax highlighting hint")
	rootCmd.AddCommand(sealCmd)
}

var sealCmd = &cobra.Command{
	Use:   "seal [content]",
	Short: "Encrypt and share content",
	Long:  "Encrypt content client-side and upload to the Capsule server. Content can be passed as an argument or piped via stdin.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sealNoBurn {
			sealBurn = false
		}

		var content string
		if len(args) > 0 {
			content = strings.Join(args, " ")
		} else if !term.IsTerminal(int(os.Stdin.Fd())) {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			content = string(data)
		} else {
			return fmt.Errorf("provide content as an argument or pipe via stdin")
		}

		if content == "" {
			return fmt.Errorf("content cannot be empty")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		key, err := crypto.GenerateKey()
		if err != nil {
			return err
		}

		encrypted, err := crypto.Encrypt([]byte(content), key)
		if err != nil {
			return fmt.Errorf("encrypting: %w", err)
		}

		var fragment string
		hasPassword := false

		if sealPassword {
			fmt.Fprint(os.Stderr, "Password: ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}
			fmt.Fprintln(os.Stderr)

			if len(pw) == 0 {
				return fmt.Errorf("password cannot be empty")
			}

			fragment, err = crypto.WrapKey(key, string(pw))
			if err != nil {
				return fmt.Errorf("wrapping key: %w", err)
			}
			hasPassword = true
		} else {
			fragment = crypto.ToBase64URL(key)
		}

		client := api.NewClient(cfg.ServerURL)
		resp, err := client.CreatePaste(&api.CreatePasteRequest{
			Content:       encrypted,
			BurnAfterRead: sealBurn,
			ExpiresIn:     sealExpires,
			HasPassword:   hasPassword,
			Syntax:        sealSyntax,
		})
		if err != nil {
			return fmt.Errorf("uploading: %w", err)
		}

		url := fmt.Sprintf("%s/%s#%s", cfg.ServerURL, resp.ID, fragment)
		green := color.New(color.FgGreen)
		green.Println(url)

		dim := color.New(color.Faint)
		dim.Fprintf(os.Stderr, "\nDelete token: %s\n", resp.DeleteToken)

		return nil
	},
}
