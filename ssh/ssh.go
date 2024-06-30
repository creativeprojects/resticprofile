package ssh

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/creativeprojects/clog"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const startPort = 10001

type SSH struct {
	config Config
	port   int
	client *ssh.Client
	tunnel net.Listener
	server *http.Server
}

func NewSSH(config Config) *SSH {
	return &SSH{
		config: config,
		port:   startPort,
	}
}

func (s *SSH) Connect() error {
	err := s.config.Validate()
	if err != nil {
		return err
	}
	hostKeyCallback, err := knownhosts.New(s.config.KnownHostsPath)
	if err != nil {
		return fmt.Errorf("cannot load host keys from known_hosts: %w", err)
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

	config := &ssh.ClientConfig{
		User: s.config.Username,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback:   hostKeyCallback,
		HostKeyAlgorithms: []string{ssh.KeyAlgoED25519, ssh.KeyAlgoECDSA256}, // we might need to make this configurable
	}

	// Connect to the remote server and perform the SSH handshake.
	s.client, err = ssh.Dial("tcp", s.config.Host, config)
	if err != nil {
		return fmt.Errorf("unable to connect: %w", err)
	}

	// Request the remote side to open a local port
	s.tunnel, err = s.client.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		log.Fatal("unable to register tcp forward: ", err)
	}

	go func() {
		s.server = &http.Server{
			Handler: s.config.Handler,
		}
		// Serve HTTP with your SSH server acting as a reverse proxy.
		err := s.server.Serve(s.tunnel)
		if err != nil && err != http.ErrServerClosed {
			clog.Warningf("unable to serve http: %s", err)
		}
	}()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := s.client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, we can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	// "/home/gouarfig/go/src/github.com/creativeprojects/resticprofile/resticprofile"
	if err := session.Run(fmt.Sprintf("curl http://localhost:%d | tar -tv", s.port)); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
	return nil
}

func (s *SSH) Close() {
	if s.tunnel != nil {
		err := s.tunnel.Close()
		if err != nil {
			clog.Warningf("unable to close tunnel: %s", err)
		}
	}
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
