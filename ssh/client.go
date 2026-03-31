package ssh

import "context"

type Client interface {
	Name() string
	// Connect establishes the SSH connection and starts the file server.
	// It returns an error if the connection or server setup fails.
	// You SHOULD run the Close() method even after a connection failure.
	Connect(ctx context.Context) error
	Close(ctx context.Context)
	Run(ctx context.Context, command string, arguments ...string) error
	TunnelPeerPort() int
}
