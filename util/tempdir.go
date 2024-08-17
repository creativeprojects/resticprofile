package util

import (
	"fmt"
	"os"
	"sync"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/util/shutdown"
)

var (
	tempDirInitializer sync.Once
	tempDir            string
	tempDirErr         error
)

const (
	tempDirPattern = "resticprofile*"
	tempDirHookTag = "tempDir-cleanup-hook"
)

// TempDir returns the path to a temporary directory that is stable within the same process and cleaned when shutdown.RunHooks is invoked
func TempDir() (string, error) {
	tempDirInitializer.Do(func() {
		tempDir, tempDirErr = createTempDir("")

		if !shutdown.ContainsHook(tempDirHookTag) {
			shutdown.AddHook(func() {
				removeTempDir(tempDir, tempDirErr)
				tempDir = ""
				tempDirErr = fmt.Errorf("illegal state: temp directory has been removed")
			}, tempDirHookTag)
		}
	})

	return tempDir, tempDirErr
}

// MustGetTempDir returns the dir from TempDir or panics if an error occurred
func MustGetTempDir() string {
	if dir, err := TempDir(); err == nil {
		return dir
	} else {
		panic(err)
	}
}

// ClearTempDir removes the temporary directory (if present) and resets the state.
// This is not safe for concurrent use and is meant for cleanup in unit tests only.
func ClearTempDir() {
	removeTempDir(tempDir, tempDirErr)
	tempDir = ""
	tempDirErr = nil
	tempDirInitializer = sync.Once{}
}

func createTempDir(path string) (tempDir string, tempDirErr error) {
	if temp, tempErr := os.MkdirTemp(path, tempDirPattern); tempErr == nil {
		tempDir = temp
	} else {
		cacheDir, err := os.UserCacheDir()
		if err == nil {
			if temp, err = os.MkdirTemp(cacheDir, tempDirPattern); err == nil {
				tempDir = temp
			}
		}
		if err != nil {
			clog.Errorf("failed creating temp dir in temp: %q and cache: %q", tempErr.Error(), err.Error())
			tempDirErr = err
		}
	}

	if len(tempDir) > 0 {
		clog.Tracef("temporary directory created: %s", tempDir)
	}
	return
}

func removeTempDir(tempDir string, tempDirErr error) {
	if len(tempDir) > 0 && tempDirErr == nil {
		tempDirErr = os.RemoveAll(tempDir)
		if tempDirErr != nil {
			clog.Warningf("failed removing temporary directory %q: %s", tempDir, tempDirErr.Error())
		}
	}
}
