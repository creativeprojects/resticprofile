package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type InternalClient struct {
	config     Config
	tunnelPort int
	client     *ssh.Client
	tunnel     net.Listener
	server     *http.Server
	wg         sync.WaitGroup
}

func NewInternalClient(config Config) *InternalClient {
	return &InternalClient{
		config: config,
	}
}

func (s *InternalClient) Name() string {
	return "InternalSSH"
}

// Connect establishes the SSH connection and starts the file server.
// It returns an error if the connection or server setup fails.
// You SHOULD run the Close() method even after a connection failure.
func (s *InternalClient) Connect(_ context.Context) error {
	err := s.config.ValidateInternal()
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

	authMethods := make([]ssh.AuthMethod, 0, len(s.config.PrivateKeyPaths))
	for _, privateKeyPath := range s.config.PrivateKeyPaths {
		key, err := os.ReadFile(privateKeyPath)
		if err != nil {
			clog.Errorf("unable to read private key: %s", err)
			continue
		}

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			clog.Errorf("unable to parse private key: %s", err)
			continue
		}
		clog.Debugf("using ssh key of type %q", signer.PublicKey().Type())
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if len(authMethods) == 0 {
		return errors.New("no ssh key found")
	}

	// The algorithms returned by ssh.SupportedAlgorithms() are different from
	// the default ones and do not include algorithms that are considered
	// insecure, such as those using SHA-1, returned by
	// ssh.InsecureAlgorithms().
	algorithms := ssh.SupportedAlgorithms()

	config := &ssh.ClientConfig{
		User:              s.config.Username,
		Auth:              authMethods,
		HostKeyCallback:   getHostKeyCallback(hostKeyCallback),
		HostKeyAlgorithms: algorithms.HostKeys,
		Config: ssh.Config{
			KeyExchanges: algorithms.KeyExchanges,
			Ciphers:      algorithms.Ciphers,
			MACs:         algorithms.MACs,
		},
	}

	host := s.config.Host
	if s.config.Port > 0 {
		host = net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port))
	}
	// Connect to the remote server and perform the SSH handshake.
	s.client, err = ssh.Dial("tcp", host, config)
	if err != nil {
		return fmt.Errorf("unable to connect: %w", err)
	}

	// Request the remote side to open a local port
	s.tunnel, err = s.client.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("unable to register tcp forward: %w", err)
	}
	// the return type is net.Addr only but we also need the allocated port
	addrWithPort, ok := s.tunnel.Addr().(interface{ AddrPort() netip.AddrPort })
	if !ok {
		return fmt.Errorf("cannot determine remote tunnel port")
	}
	s.tunnelPort = int(addrWithPort.AddrPort().Port())

	s.wg.Go(func() {
		s.server = &http.Server{
			Handler:           s.config.Handler,
			ReadHeaderTimeout: 5 * time.Second,
		}
		// Serve HTTP with your SSH server acting as a reverse proxy.
		err := s.server.Serve(s.tunnel)
		if err != nil && err != http.ErrServerClosed && !errors.Is(err, io.EOF) {
			clog.Warningf("unable to serve http: %s", err)
		}
	})
	// do we need to wait for the server to start?
	return nil
}

func (s *InternalClient) TunnelPeerPort() int {
	return s.tunnelPort
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
	s.wg.Wait()
}

func getHostKeyCallback(next ssh.HostKeyCallback) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		clog.Debugf("initiating SSH connection to %s using internal client", remote.String())
		if next != nil {
			err := next(hostname, remote, key)
			if err != nil {
				clog.Warningf("host key verification failed: %s", err)
			}
			return err
		}
		return nil
	}
}

// verify interface
var _ Client = (*InternalClient)(nil)
