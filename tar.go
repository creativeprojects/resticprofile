package main

import (
	"archive/tar"
	"io"
	"os"

	"github.com/creativeprojects/clog"
)

func sendFiles(w io.Writer, files []string) error {
	tarWriter := tar.NewWriter(w)
	defer tarWriter.Close()

	for _, filename := range files {
		fileInfo, err := os.Stat(filename)
		if err != nil {
			clog.Errorf("unable to stat file %s: %v", filename, err)
			continue
		}
		fileHeader, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			clog.Errorf("unable to create tar header for file %s: %v", filename, err)
			continue
		}
		err = tarWriter.WriteHeader(fileHeader)
		if err != nil {
			clog.Errorf("unable to write tar header for file %s: %v", filename, err)
			break
		}
		file, err := os.Open(filename)
		if err != nil {
			clog.Errorf("unable to open file %s: %v", filename, err)
			continue
		}
		defer file.Close()

		written, err := io.Copy(tarWriter, file)
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
