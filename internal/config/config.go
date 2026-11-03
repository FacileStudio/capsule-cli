package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultServerURL is the default server base URL.
const DefaultServerURL = "https://capsule.facile.studio"

// Config holds the local configuration.
type Config struct {
	ServerURL string `yaml:"server_url"`
	Token     string `yaml:"token,omitempty"`
}

// Path returns the path to the configuration file.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".capsule.yml"), nil
}

// Load reads and parses the configuration file.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := &Config{ServerURL: DefaultServerURL}
			if writeErr := Save(cfg); writeErr != nil {
				return cfg, nil
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if info, statErr := os.Stat(path); statErr == nil && info.Mode().Perm()&0o077 != 0 {
		_ = os.Chmod(path, 0o600)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.ServerURL == "" {
		cfg.ServerURL = DefaultServerURL
	}

	return &cfg, nil
}

// Server returns the server base URL: CAPSULE_SERVER_URL, then CAPSULE_URL,
// then the stored value, then DefaultServerURL.
func (c *Config) Server() string {
	for _, value := range []string{os.Getenv("CAPSULE_SERVER_URL"), os.Getenv("CAPSULE_URL"), c.ServerURL} {
		if value != "" {
			return strings.TrimRight(value, "/")
		}
	}
	return DefaultServerURL
}

// Save writes the configuration to disk.
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	defer file.Close()
	if err := file.Chmod(0o600); err != nil {
		return fmt.Errorf("securing config: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
