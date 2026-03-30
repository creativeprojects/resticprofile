package remote

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/spf13/afero"
)

type Tar struct {
	writer        *tar.Writer
	fs            afero.Fs
	preparedFiles map[string]os.FileInfo
}

func NewTar(w io.Writer) *Tar {
	return &Tar{
		writer:        tar.NewWriter(w),
		fs:            afero.NewOsFs(),
		preparedFiles: make(map[string]os.FileInfo),
	}
}

func (t *Tar) WithFs(fs afero.Fs) *Tar {
	t.fs = fs
	return t
}

// PrepareFiles stats the given files and stores their FileInfo for later use in SendFiles.
// This way, if a file is missing, we can return an error before starting to write the tar stream.
func (t *Tar) PrepareFiles(files []string) error {
	for _, filename := range files {
		fileInfo, err := t.fs.Stat(filename)
		if err != nil {
			return fmt.Errorf("unable to stat file %s: %w", filename, err)
		}
		t.preparedFiles[filename] = fileInfo
	}
	return nil
}

// SendFiles sends all prepared files to the tar writer
func (t *Tar) SendFiles() error {
	for filename := range t.preparedFiles {
		err := t.sendFile(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

// sendFile sends a single file to the tar writer
func (t *Tar) sendFile(filename string) error {
	var err error
	fileInfo, ok := t.preparedFiles[filename]
	if !ok {
		fileInfo, err = t.fs.Stat(filename)
		if err != nil {
			return fmt.Errorf("unable to stat file %s: %w", filename, err)
		}
	}
	fileHeader, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return fmt.Errorf("unable to create tar header for file %s: %w", filename, err)
	}

	err = t.writer.WriteHeader(fileHeader)
	if err != nil {
		return fmt.Errorf("unable to write tar header for file %s: %w", filename, err)
	}
	file, err := t.fs.Open(filename)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", filename, err)
	}
	defer file.Close()

	written, err := io.Copy(t.writer, file)
	if err != nil {
		return fmt.Errorf("unable to write file %s: %w", filename, err)
	}
	if written != fileInfo.Size() {
		return fmt.Errorf("file %s: written %d bytes, expected %d", filename, written, fileInfo.Size())
	}
	clog.Debugf("file %s: written %d bytes", filename, written)
	return nil
}

func (t *Tar) SendFile(name string, data []byte) error {
	header := &tar.Header{
		Name:     name,
		Size:     int64(len(data)),
		ModTime:  time.Now(),
		Mode:     0o444,
		Typeflag: tar.TypeReg,
	}
	if !platform.IsWindows() {
		header.Uid = os.Geteuid()
		header.Gid = os.Getegid()
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
