package ssh

import (
	"fmt"
	"net/http"
)

type Config struct {
	Host           string
	Username       string
	PrivateKeyPath string
	KnownHostsPath string
	SSHConfigPath  string // Path to the OpenSSH config file, if any
	Handler        http.Handler
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	return nil
}
