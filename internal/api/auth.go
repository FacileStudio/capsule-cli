package api

import (
	"context"
	"net/http"
)

// AuthConfig is what the server advertises about sign-in.
type AuthConfig struct {
	OIDCEnabled bool `json:"oidc_enabled"`
	SSOOnly     bool `json:"sso_only"`
}

// Me is the signed-in user.
type Me struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// AuthConfig reads the server's sign-in modes.
func (c *Client) AuthConfig(ctx context.Context) (AuthConfig, error) {
	var out AuthConfig
	err := c.do(ctx, http.MethodGet, "/api/auth/config", nil, &out)
	return out, err
}

// ExchangeCode trades a one-time login code for a session token.
func (c *Client) ExchangeCode(ctx context.Context, code string) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/auth/oidc/exchange", map[string]string{"code": code}, &out); err != nil {
		return "", err
	}
	return out.Token, nil
}

// Me returns the user the client's token belongs to.
func (c *Client) Me(ctx context.Context) (Me, error) {
	var out Me
	err := c.do(ctx, http.MethodGet, "/api/auth/me", nil, &out)
	return out, err
}

// Logout ends the session the client's token belongs to.
func (c *Client) Logout(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/api/auth/logout", nil, nil)
}
