package ssh

type Client interface {
	Name() string
	Connect() error
	Close()
	Run(command string, arguments ...string) error
	TunnelPeerPort() int
}
