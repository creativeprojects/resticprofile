package write

import (
	"errors"
	"os"

	"github.com/creativeprojects/resticprofile/platform"
)

type File struct {
	filename string
	perm     os.FileMode
	flag     int
	keepOpen bool
	handle   *os.File
}

func NewFile(filename string, options ...FileOption) (f *File, err error) {
	f = &File{
		filename: filename,
		perm:     0644,
		flag:     os.O_WRONLY | os.O_APPEND | os.O_CREATE,
		keepOpen: !platform.IsWindows(),
	}

	for _, option := range options {
		option(f)
	}

	err = f.open()
	if !f.keepOpen {
		defer func() {
			err = errors.Join(err, f.Close())
		}()
	}
	return
}

func (f *File) open() error {
	if f.handle != nil {
		return nil
	}
	var err error
	f.handle, err = os.OpenFile(f.filename, f.flag, f.perm)
	return err
}

func (f *File) Close() error {
	if f.handle != nil {
		err := f.handle.Close()
		f.handle = nil
		return err
	}
	return nil
}

func (f *File) Flush() error {
	if f.handle != nil {
		return f.handle.Sync()
	}
	return nil
}

func (f *File) Write(data []byte) (n int, err error) {
	if !f.keepOpen {
		err := f.open()
		if err != nil {
			return 0, err
		}
		defer func() {
			err = errors.Join(err, f.Close())
		}()
	}
	n, err = f.handle.Write(data)
	err = errors.Join(err, f.Flush())
	return
}
