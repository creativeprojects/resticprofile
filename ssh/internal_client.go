package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const startPort = 10001

type InternalClient struct {
	config Config
	port   int
	client *ssh.Client
	tunnel net.Listener
	server *http.Server
}

func NewInternalClient(config Config) *InternalClient {
	return &InternalClient{
		config: config,
		port:   startPort,
	}
}

func (s *InternalClient) Name() string {
	return "InternalSSH"
}

func (s *InternalClient) Connect(_ context.Context) error {
	err := s.config.Validate()
	if err != nil {
		return err
	}
	var hostKeyCallback ssh.HostKeyCallback
	if s.config.KnownHostsPath != "" && s.config.KnownHostsPath != "none" && s.config.KnownHostsPath != "/dev/null" {
		hostKeyCallback, err = knownhosts.New(s.config.KnownHostsPath)
		if err != nil {
			return fmt.Errorf("cannot load host keys from known_hosts: %w", err)
		}
	}
	key, err := os.ReadFile(s.config.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %w", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %w", err)
	}
	clog.Debugf("using ssh keys of type %q", signer.PublicKey().Type())

	// The algorithms returned by ssh.SupportedAlgorithms() are different from
	// the default ones and do not include algorithms that are considered
	// insecure, such as those using SHA-1, returned by
	// ssh.InsecureAlgorithms().
	algorithms := ssh.SupportedAlgorithms()

	config := &ssh.ClientConfig{
		User: s.config.Username,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback:   getHostKeyCallback(hostKeyCallback),
		HostKeyAlgorithms: algorithms.HostKeys,
		Config: ssh.Config{
			KeyExchanges: algorithms.KeyExchanges,
			Ciphers:      algorithms.Ciphers,
			MACs:         algorithms.MACs,
		},
	}

	// Connect to the remote server and perform the SSH handshake.
	s.client, err = ssh.Dial("tcp", s.config.Host, config)
	if err != nil {
		return fmt.Errorf("unable to connect: %w", err)
	}

	// Request the remote side to open a local port
	s.tunnel, err = s.client.Listen("tcp", fmt.Sprintf("localhost:%d", s.port)) // increment the port in a loop in case of an error
	if err != nil {
		return fmt.Errorf("unable to register tcp forward: %w", err)
	}

	go func() {
		s.server = &http.Server{
			Handler:           s.config.Handler,
			ReadHeaderTimeout: 5 * time.Second,
		}
		// Serve HTTP with your SSH server acting as a reverse proxy.
		err := s.server.Serve(s.tunnel)
		if err != nil && err != http.ErrServerClosed && !errors.Is(err, io.EOF) {
			clog.Warningf("unable to serve http: %s", err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // wait for the server to start
	return nil
}

func (s *InternalClient) TunnelPeerPort() int {
	return s.port
}

func (s *InternalClient) Run(_ context.Context, command string, arguments ...string) error {
	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// request a pseudo terminal to display colors
	if termType := os.Getenv("TERM"); termType != "" {
		modes := ssh.TerminalModes{
			ssh.ECHO: 0, // disable echoing
		}
		if err := session.RequestPty(termType, 40, 80, modes); err != nil {
			clog.Warningf("request for pseudo terminal failed: %s", err)
		}
	}

	// Once a Session is created, we can execute a single command on
	// the remote side using the Run method.
	cmdline := command + " " + strings.Join(arguments, " ")
	clog.Debugf("running command: %s", cmdline)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if err := session.Run(cmdline); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}

func (s *InternalClient) Close(ctx context.Context) {
	// close the tunnel first otherwise it fails with error: "ssh: cancel-tcpip-forward failed"
	if s.tunnel != nil {
		err := s.tunnel.Close()
		if err != nil {
			clog.Warningf("unable to close tunnel: %s", err)
		}
	}
	if s.server != nil {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		err := s.server.Shutdown(ctx)
		if err != nil {
			clog.Warningf("unable to close http server: %s", err)
		}
	}
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			clog.Warningf("unable to close ssh connection: %s", err)
		}
	}
}

func getHostKeyCallback(next ssh.HostKeyCallback) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		clog.Debugf("initiating SSH connection to %s using internal client", remote.String())
		if next != nil {
			err := next(hostname, remote, key)
			if err != nil {
				clog.Warningf("host key verification failed: %w", err)
			}
			return err
		}
		return nil
	}
}

// verify interface
var _ Client = (*InternalClient)(nil)
