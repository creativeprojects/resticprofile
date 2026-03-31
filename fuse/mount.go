//go:build !windows

package fuse

import (
	"fmt"

	"github.com/creativeprojects/clog"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func MountFS(mountpoint string, files []File) (func(), error) {
	memFS := newMemFS(files)

	clog.Debugf("mounting filesystem at %s", mountpoint)

	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug:         false, // generates a LOT of logs
			FsName:        "resticprofile",
			DisableXAttrs: true,
			EnableLocks:   false,
		},
	}
	server, err := fs.Mount(mountpoint, memFS, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to mount filesystem: %w", err)
	}
	closeFS := func() {
		clog.Debug("unmounting filesystem")
		err := server.Unmount() // don't need to call Wait after Unmount
		if err != nil {
			clog.Errorf("failed to unmount filesystem: %v", err)
		}

		memFS.Close()
	}
	return closeFS, nil
}
