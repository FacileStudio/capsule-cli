package cmd

import (
	"fmt"
	"net/url"
	"strings"
)

type ParsedURL struct {
	ServerURL string
	ID        string
	Fragment  string
}

func parseURL(rawURL string) (*ParsedURL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid URL: missing scheme or host")
	}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")
	id := parts[len(parts)-1]
	if id == "" {
		return nil, fmt.Errorf("no paste ID found in URL")
	}

	return &ParsedURL{
		ServerURL: u.Scheme + "://" + u.Host,
		ID:        id,
		Fragment:  u.Fragment,
	}, nil
}
