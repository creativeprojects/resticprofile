package systemd

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/creativeprojects/clog"
)

var (
	ownedDropInRegex = regexp.MustCompile(".resticprofile.conf$")
)

func CreateDropIns(dir string, files []string) error {
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fileBasenames := make(map[string]struct{})
	for _, file := range files {
		fileBasenames[filepath.Base(file)] = struct{}{}
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
		_, notOrphaned := fileBasenames[f.Name()]
		if createdByUs && !notOrphaned {
			clog.Infof("deleting orphaned drop-in file %v", f.Name())
			if err := fs.Remove(filepath.Join(dir, f.Name())); err != nil {
				return err
			}
		}
	}

	for _, dropInFile := range files {
		// change the extension to prepend `.resticprofile`
		// to signify it wasn't created outside of resticprofile, i.e. we own it
		origExt := filepath.Ext(dropInFile)
		dropInFileOwned := fmt.Sprintf("%s.resticprofile%s", strings.TrimSuffix(dropInFile, origExt), origExt)
		dst, err := fs.Create(filepath.Join(dir, dropInFileOwned))
		if err != nil {
			return err
		}
		src, err := fs.Open(dropInFile)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	return nil
}
