package ssh

import "context"

type Client interface {
	Name() string
	Connect(ctx context.Context) error
	Close(ctx context.Context)
	Run(command string, arguments ...string) error
	TunnelPeerPort() int
}
