package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/config"
	"github.com/FacileStudio/capsule-cli/internal/ui"
)

var (
	loginToken      string
	loginTokenStdin bool
	loginNoBrowser  bool
	loginNoListener bool
)

var loginCmd = &cobra.Command{
	Use:   "login [url]",
	Short: "Sign in to a Capsule server",
	Long: `Sign in to a Capsule server through your browser.

The server defaults to https://capsule.facile.studio. A session already open with
another Facile tool completes the login without a prompt. Alternatives:

  capsule login <url> --token <token>     use an existing token
  capsule login <url> --token-stdin       read the token from stdin
  capsule login <url> --no-listener       sign in from a browser on another machine

--no-listener is for the terminal whose browser is somewhere else: the server
shows a code on the page and this command reads it from stdin.

The URL may be omitted once CAPSULE_SERVER_URL, CAPSULE_URL or a previous login
has set one. Without any of them the default server is used.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVar(&loginToken, "token", "", "Authenticate with an existing token")
	loginCmd.Flags().BoolVar(&loginTokenStdin, "token-stdin", false, "Read the token from stdin")
	loginCmd.Flags().BoolVar(&loginNoBrowser, "no-browser", false, "Print the sign-in URL instead of opening a browser")
	loginCmd.Flags().BoolVar(&loginNoListener, "no-listener", false, "Sign in from a browser on another machine, pasting the code back")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	serverURL := cfg.Server()
	if len(args) == 1 {
		serverURL = strings.TrimRight(args[0], "/")
	}

	token, err := obtainToken(serverURL)
	if err != nil {
		return err
	}

	cfg.ServerURL = serverURL
	cfg.Token = token
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	ui.Success("Logged in to %s", serverURL)
	return nil
}

func obtainToken(serverURL string) (string, error) {
	if loginToken != "" || loginTokenStdin {
		return tokenFromFlags(serverURL)
	}
	if !loginNoListener && !browserAvailable() {
		return "", fmt.Errorf("no browser available here: run 'capsule login --no-listener', or pass --token or --token-stdin")
	}
	if !serverOffersSSO(serverURL) {
		return "", fmt.Errorf("%s offers no browser sign-in: pass --token or --token-stdin", serverURL)
	}
	return ssoLogin(serverURL)
}

func tokenFromFlags(serverURL string) (string, error) {
	token := loginToken
	if loginTokenStdin {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read token: %w", err)
		}
		token = strings.TrimSpace(string(raw))
	}
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	if _, err := api.New(serverURL, token).Me(context.Background()); err != nil {
		return "", fmt.Errorf("token rejected by %s: %w", serverURL, err)
	}
	return token, nil
}
