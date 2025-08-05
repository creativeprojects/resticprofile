package remote

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/spf13/afero"
)

type Tar struct {
	writer *tar.Writer
}

func NewTar(w io.Writer) *Tar {
	return &Tar{
		writer: tar.NewWriter(w),
	}
}

func (t *Tar) SendFiles(fs afero.Fs, files []string) error {
	for _, filename := range files {
		fileInfo, err := fs.Stat(filename)
		if err != nil {
			clog.Errorf("unable to stat file %s: %v", filename, err)
			continue
		}
		fileHeader, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			clog.Errorf("unable to create tar header for file %s: %v", filename, err)
			continue
		}
		err = t.writer.WriteHeader(fileHeader)
		if err != nil {
			clog.Errorf("unable to write tar header for file %s: %v", filename, err)
			break
		}
		file, err := fs.Open(filename)
		if err != nil {
			clog.Errorf("unable to open file %s: %v", filename, err)
			continue
		}
		defer file.Close()

		written, err := io.Copy(t.writer, file)
		if err != nil {
			clog.Errorf("unable to write file %s: %v", filename, err)
			break
		}
		if written != fileInfo.Size() {
			clog.Errorf("file %s: written %d bytes, expected %d", filename, written, fileInfo.Size())
			break
		}
		clog.Debugf("file %s: written %d bytes", filename, written)
	}
	return nil
}

func (t *Tar) SendFile(name string, data []byte) error {
	header := &tar.Header{
		Name:     name,
		Size:     int64(len(data)),
		ModTime:  time.Now(),
		Mode:     0o444,
		Typeflag: tar.TypeReg,
		Uid:      os.Geteuid(),
		Gid:      os.Getegid(),
	}
	if err := t.writer.WriteHeader(header); err != nil {
		return err
	}
	written, err := t.writer.Write(data)
	if err != nil {
		return err
	}
	if written != len(data) {
		return fmt.Errorf("manifest written %d bytes, expected %d", written, len(data))
	}
	clog.Debugf("manifest written %d bytes", written)
	return nil
}

func (t *Tar) Close() error {
	return t.writer.Close()
}
