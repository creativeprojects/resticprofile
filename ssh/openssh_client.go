package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
)

type OpenSSHClient struct {
	config         Config
	sshHost        string
	sshUserHost    string
	sshPort        int
	listener       net.Listener
	server         *http.Server
	wg             sync.WaitGroup
	tempDir        string // Temporary directory for SSH socket
	socket         string
	peerTunnelPort int
}

func NewOpenSSHClient(config Config) *OpenSSHClient {
	return &OpenSSHClient{
		config: config,
	}
}

func (c *OpenSSHClient) Name() string {
	return "OpenSSH"
}

// Connect establishes the SSH connection and starts the file server.
// It returns an error if the connection or server setup fails.
// You SHOULD run the Close() method even after a connection failure.
func (c *OpenSSHClient) Connect(ctx context.Context) error {
	err := c.config.ValidateOpenSSH()
	if err != nil {
		return err
	}
	c.sshHost, c.sshPort, c.sshUserHost = c.config.Host, c.config.Port, c.config.Host
	if c.config.Username != "" {
		c.sshUserHost = fmt.Sprintf("%s@%s", c.config.Username, c.sshHost)
	}
	err = c.startSSH(ctx)
	if err != nil {
		return fmt.Errorf("error while starting SSH connection: %w", err)
	}
	err = c.startFileServer(ctx)
	if err != nil {
		return err
	}
	err = c.startTunnel(ctx)
	if err != nil {
		return fmt.Errorf("error while starting SSH tunnel: %w", err)
	}
	return nil
}

func (c *OpenSSHClient) startFileServer(ctx context.Context) error {
	var err error
	listenConfig := net.ListenConfig{}
	c.listener, err = listenConfig.Listen(ctx, "tcp", "localhost:0")
	if err != nil {
		return err
	}
	c.server = &http.Server{
		Handler:           c.config.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	c.wg.Go(func() {
		defer c.listener.Close()

		clog.Debugf("file server listening locally on %s", c.listener.Addr().String())
		err := c.server.Serve(c.listener)
		if err != nil && err != http.ErrServerClosed {
			clog.Error("error while serving HTTP:", err)
		}
	})
	return nil
}

func (c *OpenSSHClient) startSSH(ctx context.Context) error {
	var err error
	c.tempDir, err = os.MkdirTemp("", "rp-ssh")
	if err != nil {
		return fmt.Errorf("error creating temporary directory for SSH socket: %w", err)
	}
	c.socket = filepath.Join(c.tempDir, "ssh.sock")
	args := make([]string, 0, 10)
	args = append(args,
		"-f",           // Requests ssh to go to background just before command execution
		"-M",           // Places the ssh client into “master” mode for connection sharing
		"-N",           // Do not execute a remote command
		"-S", c.socket, // Specifies the location of the control socket
	)
	if c.config.SSHConfigPath != "" {
		args = append(args, "-F", c.config.SSHConfigPath)
	}
	if c.sshPort > 0 {
		args = append(args, "-p", strconv.Itoa(c.sshPort))
	}
	if c.config.KnownHostsPath != "" {
		args = append(args, "-o", fmt.Sprintf("UserKnownHostsFile=%s", c.config.KnownHostsPath))
	}
	for _, privateKeyPath := range c.config.PrivateKeyPaths {
		args = append(args, "-i", privateKeyPath)
	}

	args = append(args, c.sshUserHost)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	clog.Debugf("running command: %s", cmd.String())
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error while running ssh command: %w", err)
	}
	return nil
}

func (c *OpenSSHClient) stopSSH(ctx context.Context) error {
	if c.socket == "" {
		// connection not established
		return nil
	}
	args := []string{
		"-S", c.socket, // Specifies the location of the control socket
		"-O", "exit", // Requests the master to exit
		c.sshUserHost, // Not used in this case, but required by ssh
	}
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	clog.Debugf("running command: %s", cmd.String())
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error while running ssh command: %w", err)
	}
	return nil
}

func (c *OpenSSHClient) startTunnel(ctx context.Context) error {
	if c.socket == "" {
		return errors.New("SSH connection not established")
	}
	args := []string{
		"-S", c.socket, // Specifies the location of the control socket
		"-O", "forward", // Requests the master to do a port forward
		"-R", fmt.Sprintf("0:localhost:%d", c.listener.Addr().(*net.TCPAddr).Port), // Forward random remote port to local port
		c.sshUserHost, // Not used in this case, but required by ssh
	}
	cmd := exec.CommandContext(ctx, "ssh", args...)
	clog.Debugf("running command: %s", cmd.String())
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error while running ssh command: %w", err)
	}
	if len(output) == 0 {
		return fmt.Errorf("no output from SSH tunnel command")
	}
	output = bytes.TrimSpace(output)
	port, err := strconv.Atoi(string(output))
	if err != nil {
		return fmt.Errorf("error parsing SSH tunnel output: %w", err)
	}
	c.peerTunnelPort = port
	clog.Debugf("port %d opened in tunnel", c.peerTunnelPort)
	return nil
}

func (c *OpenSSHClient) Close(ctx context.Context) {
	if c.server != nil {
		err := c.server.Shutdown(ctx)
		if err != nil {
			clog.Warningf("unable to shutdown server: %s", err)
		}
		c.server = nil
	}
	err := c.stopSSH(ctx)
	if err != nil {
		clog.Warningf("unable to stop SSH connection: %s", err)
	}
	c.wg.Wait()
	if c.tempDir != "" {
		err := os.RemoveAll(c.tempDir)
		if err != nil {
			clog.Warningf("unable to remove temporary directory: %s", err)
		}
		c.tempDir = ""
	}
}

func (c *OpenSSHClient) Run(ctx context.Context, command string, arguments ...string) error {
	if c.socket == "" {
		return errors.New("SSH connection not established")
	}
	args := append([]string{
		"-t",           // Force pseudo-terminal allocation
		"-t",           // Even when stdin is not attached
		"-S", c.socket, // Specifies the location of the control socket
		c.sshUserHost, // Not used in this case, but required by ssh
		command,
	}, arguments...)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 10 * time.Second

	clog.Debugf("running command: %s", cmd.String())
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error while running ssh command: %w", err)
	}
	return nil
}

func (c *OpenSSHClient) TunnelPeerPort() int {
	return c.peerTunnelPort
}

// verify interface
var _ Client = (*OpenSSHClient)(nil)
