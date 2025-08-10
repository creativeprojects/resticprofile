package fuse

import "io/fs"

type File struct {
	name     string
	fileInfo fs.FileInfo
	data     []byte
}

func NewFile(name string, fileInfo fs.FileInfo, data []byte) *File {
	return &File{
		name:     name,
		fileInfo: fileInfo,
		data:     data,
	}
}

func (f *File) Close() {
	// emptying file data
	for i := range f.data {
		f.data[i] = 0
	}
	f.data = nil
}
