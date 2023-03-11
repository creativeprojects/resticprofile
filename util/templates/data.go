package templates

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DefaultData provides default variables for templates
type DefaultData struct {
	Now        time.Time
	CurrentDir string
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
		Now:      time.Now(),
		TempDir:  filepath.ToSlash(os.TempDir()),
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: "localhost",
		Env:      formatEnv(env),
	}

	if cwd, err := os.Getwd(); err == nil {
		data.CurrentDir = filepath.ToSlash(cwd)
	}

	if binary, err := os.Executable(); err == nil {
		data.BinaryDir = filepath.ToSlash(filepath.Dir(binary))
	}

	if hostname, err := os.Hostname(); err == nil {
		data.Hostname = hostname
	}

	for _, envValue := range os.Environ() {
		kv := strings.SplitN(envValue, "=", 2)
		key, value := strings.ToUpper(strings.TrimSpace(kv[0])), kv[1]
		if _, contains := data.Env[key]; !contains && key != "" {
			data.Env[key] = value
		}
	}

	return data
}

func formatEnv(env map[string]string) map[string]string {
	if env == nil {
		env = make(map[string]string)
	} else {
		for name, v := range env {
			if un := strings.ToUpper(name); un != name {
				delete(env, name)
				env[un] = v
			}
		}
	}
	return env
}
