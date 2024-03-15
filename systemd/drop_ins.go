//go:build !darwin && !windows

package systemd

import (
	"bufio"
	"fmt"
	"io"
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
	return fmt.Sprintf("%s.resticprofile%s", strings.TrimSuffix(basename, ext), ext)
}

func IsTimerDropIn(file string) bool {
	if f, err := fs.Open(file); err == nil {
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

func CreateDropIns(dir string, files []string) error {
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fileBasenamesOwned := make(map[string]struct{})
	for _, file := range files {
		fileBasenamesOwned[getOwnedName(filepath.Base(file))] = struct{}{}
	}

	d, err := fs.Open(dir)
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
			clog.Infof("deleting orphaned drop-in file %v", orphanPath)
			if err := fs.Remove(orphanPath); err != nil {
				return err
			}
		}
	}

	for _, dropInFilePath := range files {
		dropInFileBase := filepath.Base(dropInFilePath)
		// change the extension to prepend `.resticprofile`
		// to signify it wasn't created outside of resticprofile, i.e. we own it
		dropInFileOwned := getOwnedName(dropInFileBase)
		dstPath := filepath.Join(dir, dropInFileOwned)
		dst, err := fs.Create(dstPath)
		if err != nil {
			return err
		}
		src, err := fs.Open(dropInFilePath)
		if err != nil {
			return err
		}
		clog.Infof("writing %v", dstPath)
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	return nil
}
