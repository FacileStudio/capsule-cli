package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/config"
	"github.com/FacileStudio/capsule-cli/internal/ui"
)

var (
	flagKeyApp     string
	flagKeyPublic  bool
	flagKeyOrigins string
	flagKeyQuota   int
	flagKeyYes     bool
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long:  "Manage secret backend and public browser API keys for apps connecting to Capsule.",
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signalContext()
		defer stop()

		client, err := connectClient()
		if err != nil {
			return err
		}

		keys, err := client.ListKeys(ctx, flagKeyApp)
		if err != nil {
			return err
		}

		if flagJSON {
			return ui.JSON(keys)
		}
		if len(keys) == 0 {
			ui.Plain("No API keys")
			return nil
		}

		rows := make([][]string, 0, len(keys))
		for _, k := range keys {
			status := "active"
			if k.RevokedAt != nil {
				status = "revoked"
			}
			quota := "unlimited"
			if k.DailyQuota > 0 {
				quota = fmt.Sprintf("%d/day (%d used)", k.DailyQuota, k.UsedToday)
			}
			rows = append(rows, []string{
				strconv.FormatInt(k.ID, 10),
				k.App,
				k.Kind,
				k.Prefix,
				status,
				quota,
				k.CreatedAt.Format(time.RFC3339),
			})
		}
		ui.Table([]string{"ID", "APP", "KIND", "PREFIX", "STATUS", "QUOTA", "CREATED"}, rows)
		return nil
	},
}

var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API key",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(flagKeyApp) == "" {
			return errors.New("--app is required")
		}

		ctx, stop := signalContext()
		defer stop()

		client, err := connectClient()
		if err != nil {
			return err
		}

		kind := "secret"
		if flagKeyPublic {
			kind = "public"
		}

		var origins []string
		if flagKeyOrigins != "" {
			for _, o := range strings.Split(flagKeyOrigins, ",") {
				trimmed := strings.TrimSpace(o)
				if trimmed != "" {
					origins = append(origins, trimmed)
				}
			}
		}

		resp, err := client.CreateKey(ctx, api.CreateKeyRequest{
			App:            flagKeyApp,
			Kind:           kind,
			AllowedOrigins: origins,
			DailyQuota:     flagKeyQuota,
		})
		if err != nil {
			return err
		}

		if flagJSON {
			return ui.JSON(resp)
		}

		ui.Success("created %s key for %s (id: %d)", resp.Key.Kind, resp.Key.App, resp.Key.ID)
		ui.Plain("%s", resp.Token)
		ui.Hint("store this token securely, it will not be shown again")
		return nil
	},
}

var keysRevokeCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil || id <= 0 {
			return errors.New("key id must be a positive integer")
		}

		ctx, stop := signalContext()
		defer stop()

		if !flagKeyYes && !flagJSON && term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Fprintf(os.Stderr, "Revoke key %d? [y/N] ", id)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				return errors.New("aborted")
			}
		}

		client, err := connectClient()
		if err != nil {
			return err
		}

		if err := client.RevokeKey(ctx, id); err != nil {
			return err
		}

		if flagJSON {
			return ui.JSON(struct {
				ID      int64 `json:"id"`
				Revoked bool  `json:"revoked"`
			}{ID: id, Revoked: true})
		}

		ui.Success("key %d revoked", id)
		return nil
	},
}

func connectClient() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	serverURL := cfg.Server()
	token := cfg.Token
	if envToken := os.Getenv("CAPSULE_TOKEN"); envToken != "" {
		token = envToken
	}
	return api.New(serverURL, token), nil
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func init() {
	keysListCmd.Flags().StringVar(&flagKeyApp, "app", "", "Filter keys by application name")

	keysCreateCmd.Flags().StringVar(&flagKeyApp, "app", "", "Application name")
	keysCreateCmd.Flags().BoolVar(&flagKeyPublic, "public", false, "Create a public browser key instead of a secret key")
	keysCreateCmd.Flags().StringVar(&flagKeyOrigins, "origins", "", "Comma-separated allowed origins for public keys")
	keysCreateCmd.Flags().IntVar(&flagKeyQuota, "quota", 0, "Daily event quota limit for public keys")
	keysCreateCmd.MarkFlagRequired("app")

	keysRevokeCmd.Flags().BoolVarP(&flagKeyYes, "yes", "y", false, "Confirm revocation without prompting")

	keysCmd.AddCommand(keysListCmd, keysCreateCmd, keysRevokeCmd)
	rootCmd.AddCommand(keysCmd)
}
