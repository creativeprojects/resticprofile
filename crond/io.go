package crond

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

const (
	maxCrontabFileSize      = 16 * 1024 * 1024
	defaultCrontabFilePerms = fs.FileMode(0644)
)

func verifyCrontabFile(file string) error {
	if file == "" {
		return fmt.Errorf("no contrab file was specified")
	}
	return nil
}

func loadCrontabFile(file string) (content, charset string, err error) {
	if err = verifyCrontabFile(file); err != nil {
		return
	}
	var f *os.File
	if f, err = os.Open(file); err == nil {
		defer func() { _ = f.Close() }()

		var bytes []byte
		bytes, err = io.ReadAll(io.LimitReader(f, maxCrontabFileSize))
		if err == nil && len(bytes) == maxCrontabFileSize {
			err = fmt.Errorf("max file size of %d bytes exceeded in %q", maxCrontabFileSize, file)
		}
		if err == nil {
			// TODO: handle charsets
			charset = ""
			content = string(bytes)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	return
}

func saveCrontabFile(file, content, charset string) (err error) {
	if err = verifyCrontabFile(file); err != nil {
		return
	}

	// TODO: handle charsets
	bytes := []byte(content)

	if len(bytes) >= maxCrontabFileSize {
		err = fmt.Errorf("max file size of %d bytes exceeded in new %q", maxCrontabFileSize, file)
	} else {
		err = os.WriteFile(file, bytes, defaultCrontabFilePerms)
	}
	return
}

// CrontabBinary is the name of the crontab binary to use when no file is set
var CrontabBinary = DefaultCrontabBinary

func loadCrontab(file string) (content, charset string, err error) {
	if file == "" && CrontabBinary != "" {
		buffer := new(strings.Builder)
		{
			cmd := exec.Command(CrontabBinary, "-l")
			cmd.Stdout = buffer
			cmd.Stderr = buffer
			err = cmd.Run()
		}
		if err != nil {
			if strings.HasPrefix(buffer.String(), "no crontab for ") {
				err = nil // it's ok to be empty
			} else {
				err = fmt.Errorf("%w: %s", err, buffer.String())
			}
		}
		if err == nil {
			content = buffer.String()
		}
		return
	} else {
		return loadCrontabFile(file)
	}
}

func saveCrontab(file, content, charset string) (err error) {
	if file == "" && CrontabBinary != "" {
		cmd := exec.Command(CrontabBinary, "-")
		cmd.Stdin = strings.NewReader(content)
		cmd.Stderr = os.Stderr
		err = cmd.Run()
	} else {
		err = saveCrontabFile(file, content, charset)
	}
	return
}
