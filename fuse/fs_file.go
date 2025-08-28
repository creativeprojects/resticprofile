//go:build !windows

package fuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type fsFile struct {
	fs.Inode
	attr fuse.Attr
	file File
}

var _ = (fs.NodeOpener)((*fsFile)(nil))
var _ = (fs.NodeGetattrer)((*fsFile)(nil))

func (fsf *fsFile) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Attr = fsf.attr
	return 0
}

// Open only needs to send the flags back to the kernel
func (fsf *fsFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	// tell the kernel not to cache the data
	return fsf, fuse.FOPEN_DIRECT_IO, fs.OK
}

// Read simply returns the data from the file
func (fsf *fsFile) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	end := int(off) + len(dest)
	if end > len(fsf.file.data) {
		end = len(fsf.file.data)
	}
	return fuse.ReadResultData(fsf.file.data[off:end]), fs.OK
}
