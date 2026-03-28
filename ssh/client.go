package ssh

import "context"

type Client interface {
	Name() string
	Connect(context.Context) error
	Close(context.Context)
	Run(command string, arguments ...string) error
	TunnelPeerPort() int
}
