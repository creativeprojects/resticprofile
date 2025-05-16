//go:build !darwin && !windows

package systemd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/creativeprojects/clog"
)

var (
	ownedDropInRegex = regexp.MustCompile(".resticprofile.conf$")
	timerDropInRegex = regexp.MustCompile(`(?i)^\s*\[TIMER]\s*$`)
)

func getOwnedName(basename string) string {
	ext := filepath.Ext(basename)
	return fmt.Sprintf("%s.resticprofile.conf", strings.TrimSuffix(basename, ext))
}

func (u Unit) DropInFileExists(file string) bool {
	if _, err := u.fs.Stat(file); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			clog.Errorf("drop-in file %q not found", file)
			return false
		}
		clog.Error(err)
		return false
	}
	return true
}

func (u Unit) IsTimerDropIn(file string) bool {
	if f, err := u.fs.Open(file); err == nil {
		defer func() { _ = f.Close() }()
		for reader := bufio.NewScanner(f); reader.Scan(); {
			if timerDropInRegex.Match(reader.Bytes()) {
				return true
			}
		}
	} else {
		clog.Warningf("failed reading %q, cannot determine type", file)
	}
	return false
}

func (u Unit) createDropIns(dir string, files []string) error {
	if err := u.fs.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fileBasenamesOwned := make(map[string]struct{})
	for _, file := range files {
		fileBasenamesOwned[getOwnedName(filepath.Base(file))] = struct{}{}
	}

	d, err := u.fs.Open(dir)
	if err != nil {
		return err
	}

	entries, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	for _, f := range entries {
		if f.IsDir() {
			continue
		}
		createdByUs := ownedDropInRegex.MatchString(f.Name())
		_, notOrphaned := fileBasenamesOwned[f.Name()]
		if createdByUs && !notOrphaned {
			orphanPath := filepath.Join(dir, f.Name())
			clog.Debugf("deleting orphaned drop-in file %v", orphanPath)
			if err := u.fs.Remove(orphanPath); err != nil {
				return err
			}
		}
	}

	for _, dropInFilePath := range files {
		dropInFileBase := filepath.Base(dropInFilePath)
		// change the extension to `.resticprofile.conf`
		// to signify it wasn't created outside of resticprofile, i.e. we own it
		dropInFileOwned := getOwnedName(dropInFileBase)
		dstPath := filepath.Join(dir, dropInFileOwned)
		clog.Debugf("writing %v", dstPath)
		dst, err := u.fs.Create(dstPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		src, err := u.fs.Open(dropInFilePath)
		if err != nil {
			return err
		}
		defer src.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	return nil
}
