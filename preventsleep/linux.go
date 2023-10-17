//go:build linux

package preventsleep

import (
	"fmt"
	"os"

	systemd "github.com/coreos/go-systemd/v22/login1"
	"github.com/creativeprojects/resticprofile/constants"
)

const (
	inhibitWhat      = "idle:sleep:shutdown"
	inhibitWhy       = "Backup and/or restic repository maintenance"
	inhibitMode      = "block"
	permissionDenied = "Permission denied"
)

type Caffeinate struct {
	conn *systemd.Conn
	file *os.File
}

func New() *Caffeinate {
	return &Caffeinate{}
}

func (c *Caffeinate) Start() error {
	var err error

	if c.file != nil {
		return ErrAlreadyStarted
	}

	c.conn, err = systemd.New()
	if err != nil {
		return fmt.Errorf("error connecting to dbus: %w", err)
	}
	c.file, err = c.conn.Inhibit(inhibitWhat, constants.ApplicationName, inhibitWhy, inhibitMode)
	if err != nil {
		if err.Error() == permissionDenied {
			return ErrPermissionDenied
		}
		return err
	}
	return nil
}

func (c *Caffeinate) Stop() error {
	if c.file == nil {
		return ErrNotStarted
	}
	if c.file != nil {
		c.file.Close()
		c.file = nil
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	return nil
}

func (c *Caffeinate) IsRunning() bool {
	return c.conn != nil && c.file != nil
}
