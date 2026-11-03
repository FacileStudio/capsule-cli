package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/FacileStudio/porte/devicelogin"
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

// ssoLogin signs in through single sign-on and returns the session token. The
// device grant is the default; a server without the device exchange, or
// --listener, gets the loopback flow, and --no-listener the pasted code.
func ssoLogin(serverURL string) (string, error) {
	if loginNoListener {
		return ssoLoginByPaste(serverURL)
	}
	if !loginListener {
		token, fallback, err := deviceLogin(serverURL)
		if !fallback {
			return token, err
		}
	}
	return ssoLoginByListener(serverURL)
}

// deviceLogin runs the device grant and trades its token for a Capsule
// session. fallback is true when the server or the identity provider has no
// device login, so the caller switches to the loopback flow.
func deviceLogin(serverURL string) (string, bool, error) {
	session, err := devicelogin.Login(context.Background(), devicelogin.Options{
		AppOrigin:   serverURL,
		Mount:       porteMount,
		OpenBrowser: func(link string) bool { return browserAvailable() && browser.Open(link) == nil },
	})
	switch {
	case errors.Is(err, devicelogin.ErrExchangeNotMounted), errors.Is(err, devicelogin.ErrNoDeviceGrant):
		ui.Warn("%s has no device login yet, using the browser listener", serverURL)
		return "", true, nil
	case errors.Is(err, devicelogin.ErrAccountNotLinked):
		return "", false, fmt.Errorf("sign in once at %s in a browser, then run 'capsule login' again", serverURL)
	case err != nil:
		return "", false, fmt.Errorf("sign-in failed: %w", err)
	}
	return session.Token, false, nil
}

// ssoLoginByListener is the loopback flow: the browser lands on a local port
// with a one-time code. A machine with no browser to land there pastes it.
func ssoLoginByListener(serverURL string) (string, error) {
	if !browserAvailable() && !loginListener {
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
