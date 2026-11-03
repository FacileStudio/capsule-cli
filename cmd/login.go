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
	loginListener   bool
)

var loginCmd = &cobra.Command{
	Use:   "login [url]",
	Short: "Sign in to a Capsule server",
	Long: `Sign in to a Capsule server by approving a link in any browser.

The server defaults to https://capsule.facile.studio. The command prints an
approval link and waits until you approve it, on this machine or another one.
Nothing listens on a port and nothing is typed, so it works over SSH.
Alternatives:

  capsule login <url> --token <token>     use an existing token
  capsule login <url> --token-stdin       read the token from stdin
  capsule login <url> --listener          use a browser on this machine and a local port
  capsule login <url> --no-listener       paste back a code shown by the server

A server without device login falls back to the --listener flow on its own,
or to --no-listener when no browser is available here.

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
	loginCmd.Flags().BoolVar(&loginListener, "listener", false, "Use the loopback browser flow instead of the approval link")
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
