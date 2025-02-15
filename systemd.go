//go:build !darwin && !windows

package main

import (
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/creativeprojects/clog"
)

func notifyStart() {
	ok, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	if err != nil {
		clog.Errorf("cannot notify systemd: %w", err)
	}
	if ok {
		clog.Debug("running as a systemd unit: sending 'ready' status")
	}
}

func notifyStop() {
	ok, err := daemon.SdNotify(false, daemon.SdNotifyStopping)
	if err != nil {
		clog.Errorf("cannot notify systemd: %w", err)
	}
	if ok {
		clog.Debug("running as a systemd unit: sending 'stopping' status")
	}
}
