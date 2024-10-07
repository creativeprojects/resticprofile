//go:build !windows

// Simple implementation of a read-only filesystem in memory.
//
// Based on the examples at https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs#pkg-examples
package fuse

import (
	"archive/tar"
	"context"
	iofs "io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type memFS struct {
	fs.Inode

	files []File
}

func newMemFS(files []File) *memFS {
	return &memFS{
		files: files,
	}
}

var _ = (fs.InodeEmbedder)((*memFS)(nil))

// The root populates the tree in its OnAdd method
var _ = (fs.NodeOnAdder)((*memFS)(nil))

// Close erases the data from all the files
func (memfs *memFS) Close() {
	for i := range memfs.files {
		memfs.files[i].Close()
	}
	memfs.files = nil
}

// OnAdd is called once we are attached to an Inode. We can
// then construct a tree.  We construct the entire tree, and
// we don't want parts of the tree to disappear when the
// kernel is short on memory, so we use persistent inodes.
func (memfs *memFS) OnAdd(ctx context.Context) {
	for _, file := range memfs.files {
		dir, base := filepath.Split(filepath.Clean(file.name))

		p := memfs.EmbeddedInode()
		for _, comp := range strings.Split(dir, "/") {
			if len(comp) == 0 {
				continue
			}
			ch := p.GetChild(comp)
			if ch == nil {
				ch = p.NewPersistentInode(ctx,
					&fs.Inode{},
					fs.StableAttr{Mode: syscall.S_IFDIR})
				p.AddChild(comp, ch, false)
			}
			p = ch
		}

		attr := fileInfoToAttr(file.fileInfo)
		switch {
		case file.fileInfo.Mode().Type()&os.ModeSymlink == os.ModeSymlink:
			file.data = nil
			fsfile := &fsFile{
				attr: attr,
				file: file,
			}
			p.AddChild(base, memfs.NewPersistentInode(ctx, fsfile, fs.StableAttr{Mode: syscall.S_IFLNK}), false)

		case file.fileInfo.Mode().IsDir():
			fsdir := &fsFile{
				attr: attr,
				file: file,
			}
			p.AddChild(base, memfs.NewPersistentInode(ctx, fsdir, fs.StableAttr{Mode: syscall.S_IFDIR}), false)

		case file.fileInfo.Mode().IsRegular():
			fsfile := &fsFile{
				attr: attr,
				file: file,
			}
			p.AddChild(base, memfs.NewPersistentInode(ctx, fsfile, fs.StableAttr{}), false)

		default:
			log.Printf("entry %q: unsupported type '%c'", file.name, file.fileInfo.Mode().Type())
		}
	}
}

func fileInfoToAttr(fileInfo iofs.FileInfo) fuse.Attr {
	var out fuse.Attr
	if header, ok := fileInfo.Sys().(*tar.Header); ok {
		out.Mode = uint32(header.Mode)
		out.Size = uint64(header.Size)
		out.Uid = uint32(header.Uid)
		out.Gid = uint32(header.Gid)
		out.SetTimes(&header.AccessTime, &header.ModTime, &header.ChangeTime)
	} else {
		out.Mode = uint32(fileInfo.Mode())
		out.Size = uint64(fileInfo.Size())
		out.Uid = uint32(os.Geteuid())
		out.Gid = uint32(os.Getegid())
		modTime := fileInfo.ModTime()
		out.SetTimes(nil, &modTime, nil)
	}
	out.Nlink = 1
	const bs = 512
	out.Blksize = bs
	out.Blocks = (out.Size + bs - 1) / bs

	return out
}
