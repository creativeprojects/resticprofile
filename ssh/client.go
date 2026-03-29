package ssh

import "context"

type Client interface {
	Name() string
	Connect(ctx context.Context) error
	Close(ctx context.Context)
	Run(ctx context.Context, command string, arguments ...string) error
	TunnelPeerPort() int
}
