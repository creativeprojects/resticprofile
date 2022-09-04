//go:build !darwin && !windows && !linux

package preventsleep

type Caffeinate struct{}

func New() *Caffeinate {
	return &Caffeinate{}
}

func (c *Caffeinate) Start() error {
	return ErrNotSupported
}

func (c *Caffeinate) Stop() error {
	return ErrNotSupported
}

func (c *Caffeinate) IsRunning() bool {
	return false
}
