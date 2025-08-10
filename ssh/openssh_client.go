package ssh

import (
	"bytes"
	"context"
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
func (c *OpenSSHClient) Connect() error {
	c.sshHost, c.sshPort = parseHost(c.config.Host)
	c.sshUserHost = c.sshHost
	if c.config.Username != "" {
		c.sshUserHost = fmt.Sprintf("%s@%s", c.config.Username, c.sshHost)
	}
	err := c.startFileServer()
	if err != nil {
		return err
	}
	err = c.startSSH(context.Background())
	if err != nil {
		return fmt.Errorf("error while starting SSH connection: %w", err)
	}
	err = c.startTunnel(context.Background())
	if err != nil {
		return fmt.Errorf("error while starting SSH tunnel: %w", err)
	}
	return nil
}

func (c *OpenSSHClient) startFileServer() error {
	var err error
	c.listener, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	c.server = &http.Server{
		Handler:           c.config.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer c.listener.Close()

		clog.Debugf("file server listening locally on %s", c.listener.Addr().String())
		err := c.server.Serve(c.listener)
		if err != nil && err != http.ErrServerClosed {
			clog.Error("error while serving HTTP:", err)
		}
	}()
	return nil
}

func (c *OpenSSHClient) startSSH(ctx context.Context) error {
	c.socket = filepath.Join(os.TempDir(), fmt.Sprintf("ssh-%d.sock", os.Getpid()))
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
	if c.config.PrivateKeyPath != "" {
		args = append(args, "-i", c.config.PrivateKeyPath)
	}
	args = append(args, c.sshUserHost)
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

func (c *OpenSSHClient) stopSSH(ctx context.Context) error {
	if c.socket == "" {
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
		return nil
	}
	args := []string{
		"-S", c.socket, // Specifies the location of the control socket
		"-O", "forward", // Requests the master to exit
		fmt.Sprintf("-R 0:localhost:%d", c.listener.Addr().(*net.TCPAddr).Port), // Forward random remote port to local port
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

func (c *OpenSSHClient) Close() {
	ctx := context.Background()
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
}

func (c *OpenSSHClient) Run(command string, arguments ...string) error {
	if c.socket == "" {
		return nil
	}
	args := append([]string{
		"-t",           // Force pseudo-terminal allocation
		"-t",           // Even when stdin is not attached
		"-S", c.socket, // Specifies the location of the control socket
		c.sshUserHost, // Not used in this case, but required by ssh
		command,
	}, arguments...)
	cmd := exec.CommandContext(context.Background(), "ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
