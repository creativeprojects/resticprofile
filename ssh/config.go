package ssh

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Host           string
	Username       string
	PrivateKeyPath string
	KnownHostsPath string
	Handler        http.Handler
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.PrivateKeyPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to get current user home directory: %w", err)
		}
		c.PrivateKeyPath = filepath.Join(home, ".ssh/id_rsa") // we can go through all the default name for each key type
	}
	if c.KnownHostsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to get current user home directory: %w", err)
		}
		c.KnownHostsPath = filepath.Join(home, ".ssh/known_hosts")
	}
	if !strings.Contains(c.Host, ":") {
		c.Host = c.Host + ":22"
	}
	return nil
}
