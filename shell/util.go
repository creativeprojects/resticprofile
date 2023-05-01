package shell

import (
	"bufio"
	"bytes"
	"io"

	"github.com/creativeprojects/resticprofile/platform"
)

var (
	bogusPrefix = []byte("\r\x1b[2K")
)

func LineOutputFilter(output io.Writer, included func(line []byte) bool) io.WriteCloser {
	eol := []byte("\n")
	if platform.IsWindows() {
		eol = []byte("\r\n")
	}

	reader, writer := io.Pipe()

	go func() {
		var err error
		defer func() {
			_ = reader.CloseWithError(err)
		}()

		scanner := bufio.NewScanner(reader)
		for err == nil && scanner.Scan() {
			line := bytes.TrimPrefix(scanner.Bytes(), bogusPrefix)
			if !included(line) {
				continue
			}
			if err == nil {
				_, err = output.Write(line)
			}
			if err == nil {
				_, err = output.Write(eol)
			}
		}
		if err == nil {
			err = scanner.Err()
		}
	}()

	return writer
}
