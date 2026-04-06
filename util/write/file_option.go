package write

import "os"

type FileOption func(f *File)

// WithFileKeepOpen toggles whether the file is kept open between writes. Defaults to true for all OS except Windows.
func WithFileKeepOpen(keepOpen bool) FileOption {
	return func(f *File) { f.keepOpen = keepOpen }
}

// WithFilePerm sets file perms to apply when creating the file
func WithFilePerm(perm os.FileMode) FileOption {
	return func(f *File) { f.perm = perm }
}

// WithFileFlag sets file open flags
func WithFileFlag(flag int) FileOption {
	return func(f *File) { f.flag = flag }
}

// WithFileTruncate enables that existing files are truncated
func WithFileTruncate() FileOption {
	return func(f *File) { f.flag |= os.O_TRUNC }
}
