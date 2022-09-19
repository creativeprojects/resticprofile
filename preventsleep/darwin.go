//go:build darwin

package preventsleep

import (
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/creativeprojects/clog"
)

const (
	signalInterrupt = "signal: interrupt"
	command         = "/usr/bin/caffeinate"
	// Create an assertion to prevent the system from sleeping.
	// This assertion is valid only when system is running on AC power.
	flag = "-s"
)

type Caffeinate struct {
	cmd atomic.Value
	wg  sync.WaitGroup
}

func New() *Caffeinate {
	return &Caffeinate{}
}

// Start caffeinate in the background and returns.
func (c *Caffeinate) Start() error {
	cmd, ok := c.cmd.Load().(*exec.Cmd)
	if ok {
		return ErrAlreadyStarted
	}
	cmd = exec.Command(command, flag)
	c.cmd.Store(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		err := cmd.Wait()
		if err != nil {
			if err.Error() != signalInterrupt {
				clog.Warningf("caffeinate: %s", err)
			}
		}
		cmd = nil
		c.cmd.Store(cmd)
	}()
	return nil
}

// Stop caffeinate (and wait until it's stopped)
func (c *Caffeinate) Stop() error {
	cmd, ok := c.cmd.Load().(*exec.Cmd)
	if !ok || cmd == nil {
		return ErrNotStarted
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	c.wg.Wait()
	return nil
}

func (c *Caffeinate) IsRunning() bool {
	cmd, ok := c.cmd.Load().(*exec.Cmd)
	return ok && cmd != nil
}
