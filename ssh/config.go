package ssh

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// Config holds the configuration to connect to the SSH server
type Config struct {
	Host            string
	Port            int
	Username        string
	PrivateKeyPaths []string
	KnownHostsPath  string
	SSHConfigPath   string // Path to the OpenSSH config file, if any
	Handler         http.Handler
}

func (c *Config) ValidateOpenSSH() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	return nil
}

func (c *Config) ValidateInternal() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(c.PrivateKeyPaths) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to get current user home directory: %w", err)
		}
		c.PrivateKeyPaths, _ = filepath.Glob(filepath.Join(home, ".ssh/id_*[^.pub]"))
	}
	if c.KnownHostsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to get current user home directory: %w", err)
		}
		c.KnownHostsPath = filepath.Join(home, ".ssh/known_hosts")
	}
	return nil
}
