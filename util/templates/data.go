package templates

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/util"
)

// DefaultData provides default variables for templates
type DefaultData struct {
	Now        time.Time
	CurrentDir string
	StartupDir string
	TempDir    string
	BinaryDir  string
	Hostname   string
	OS         string
	Arch       string
	Env        map[string]string
}

// InitDefaults initializes DefaultData if not yet initialized
func (d *DefaultData) InitDefaults() {
	if d.Now.IsZero() {
		*d = NewDefaultData(nil)
	}
}

// NewDefaultData returns an initialized DefaultData
func NewDefaultData(env map[string]string) (data DefaultData) {
	data = DefaultData{
		Now:        time.Now(),
		TempDir:    filepath.ToSlash(os.TempDir()),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Hostname:   "localhost",
		StartupDir: startupDir,
		CurrentDir: startupDir,
	}

	if logStartupDirError != nil {
		logStartupDirError()
	}

	if dir, errorFunc := internalGetCurrentDir(".CurrentDir"); errorFunc == nil {
		data.CurrentDir = dir
	} else {
		errorFunc()
	}

	if binary, err := os.Executable(); err == nil {
		data.BinaryDir = filepath.ToSlash(filepath.Dir(binary))
	}

	if hostname, err := os.Hostname(); err == nil {
		data.Hostname = hostname
	}

	osEnv := util.NewDefaultEnvironment(os.Environ()...)
	for name, value := range env {
		osEnv.Put(osEnv.ResolveName(name), value)
	}
	data.Env = osEnv.ValuesAsMap()

	// add uppercase env variants to simplify usage in templates
	for name, value := range data.Env {
		if un := strings.ToUpper(name); un != name {
			if _, exists := data.Env[un]; !exists {
				data.Env[un] = value
			}
		}
	}

	return data
}

func internalGetCurrentDir(name string) (startupDir string, logError func()) {
	if dir, err := os.Getwd(); err == nil {
		startupDir = filepath.ToSlash(dir)
	} else {
		startupDir = "."
		logError = sync.OnceFunc(func() {
			clog.Debugf("using %q as fallback for %s ; %s", startupDir, name, err.Error())
		})
	}
	return
}

var startupDir, logStartupDirError = internalGetCurrentDir(".StartupDir")
