package crond

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/afero"
)

const (
	maxCrontabFileSize      = 16 * 1024 * 1024
	defaultCrontabFilePerms = fs.FileMode(0644)
)

var (
	ErrNoCrontabFile = errors.New("no crontab file was specified")
)

func loadCrontab(fs afero.Fs, file, crontabBinary string) (content, charset string, err error) {
	if file == "" && crontabBinary != "" {
		buffer := new(strings.Builder)
		{
			cmd := exec.Command(crontabBinary, "-l")
			cmd.Stdout = buffer
			cmd.Stderr = buffer
			err = cmd.Run()
		}
		if err != nil {
			if strings.HasPrefix(buffer.String(), "no crontab for ") {
				err = nil // it's ok to be empty
				buffer.Reset()
			} else {
				err = fmt.Errorf("%w: %s", err, buffer.String())
			}
		}
		if err == nil {
			content = buffer.String()
		}
		return
	} else {
		return loadCrontabFile(fs, file)
	}
}

func saveCrontab(fs afero.Fs, file, content, charset, crontabBinary string) (err error) {
	if file == "" && crontabBinary != "" {
		cmd := exec.Command(crontabBinary, "-")
		cmd.Stdin = strings.NewReader(content)
		cmd.Stderr = os.Stderr
		err = cmd.Run()
	} else {
		err = saveCrontabFile(fs, file, content, charset)
	}
	return
}

func verifyCrontabFile(file string) error {
	if file == "" {
		return ErrNoCrontabFile
	}
	return nil
}

func loadCrontabFile(fs afero.Fs, file string) (content, charset string, err error) {
	if err = verifyCrontabFile(file); err != nil {
		return
	}
	var f afero.File
	if f, err = fs.Open(file); err == nil {
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

//nolint:unparam
func saveCrontabFile(fs afero.Fs, file, content, charset string) (err error) {
	if err = verifyCrontabFile(file); err != nil {
		return
	}

	// TODO: handle charsets
	bytes := []byte(content)

	if len(bytes) >= maxCrontabFileSize {
		err = fmt.Errorf("max file size of %d bytes exceeded in new %q", maxCrontabFileSize, file)
	} else {
		err = afero.WriteFile(fs, file, bytes, defaultCrontabFilePerms)
	}
	return
}
