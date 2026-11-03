package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/FacileStudio/porte/loopback"
	"golang.org/x/term"

	"github.com/FacileStudio/capsule-cli/internal/api"
	"github.com/FacileStudio/capsule-cli/internal/browser"
	"github.com/FacileStudio/capsule-cli/internal/ui"
)

const (
	ssoAppName  = "Capsule"
	porteMount  = "/api"
	loginWindow = 5 * time.Minute
)

func browserAvailable() bool {
	if loginNoBrowser || !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	return browser.Available()
}

func serverOffersSSO(serverURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	discovery, err := api.NewClient(serverURL).AuthConfig(ctx)
	return err == nil && discovery.OIDCEnabled
}

func ssoLogin(serverURL string) (string, error) {
	if loginNoListener {
		return ssoLoginByPaste(serverURL)
	}

	listener, err := loopback.Listen()
	if err != nil {
		return "", fmt.Errorf("cannot open a loopback port to receive the login: %w", err)
	}
	defer listener.Close()
	listener.AppName = ssoAppName

	state, err := loopback.RandomState()
	if err != nil {
		return "", err
	}

	authURL := listener.LoginURL(serverURL, porteMount, state)
	_ = browser.Open(authURL)
	ui.Step("Waiting for the browser (if it did not open, visit %s)", authURL)

	ctx, cancel := context.WithTimeout(context.Background(), loginWindow)
	defer cancel()
	code, err := listener.WaitForCode(ctx, state)
	if err != nil {
		if errors.Is(err, loopback.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("timed out waiting for the browser, run 'capsule login' again")
		}
		return "", err
	}
	return exchangeCode(serverURL, code)
}

func ssoLoginByPaste(serverURL string) (string, error) {
	authURL := strings.TrimRight(serverURL, "/") + porteMount + "/auth/oidc?flow=cli"
	ui.Step("Open this URL in your browser: %s", authURL)
	if browserAvailable() {
		_ = browser.Open(authURL)
	}

	fmt.Fprint(os.Stderr, "Paste the code from that page: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("failed to read the login code: %w", err)
	}
	code := strings.TrimSpace(line)
	if code == "" {
		return "", fmt.Errorf("no login code was pasted, the code is good for sixty seconds so run 'capsule login' again")
	}
	return exchangeCode(serverURL, code)
}

func exchangeCode(serverURL, code string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	token, err := api.NewClient(serverURL).ExchangeCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("sign-in failed: %w", err)
	}
	if token == "" {
		return "", fmt.Errorf("the server returned no token")
	}
	return token, nil
}
