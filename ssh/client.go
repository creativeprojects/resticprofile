package ssh

type Client interface {
	Connect() error
	Close()
	Run(command string, arguments ...string) error
	TunnelPeerPort() int
}
