package util

import (
	"fmt"
	"io"
	"maps"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// MaxEnvironmentFileContentSize limits the max size of a single env file (to sanitize incorrect user input)
const MaxEnvironmentFileContentSize = 512 * 1024

// EnvironmentFile supports loading dotenv files
type EnvironmentFile struct {
	filename string
	fileInfo os.FileInfo
	env      map[string]string
}

var envFiles = sync.Map{}

// GetEnvironmentFile returns a pooled & immutable EnvironmentFile instance for the specified dotenv filename.
func GetEnvironmentFile(filename string) (env *EnvironmentFile, err error) {
	cached, ok := envFiles.Load(filename)
	if ok {
		env, ok = cached.(*EnvironmentFile)
	}

	if !ok || !env.Valid() {
		env = new(EnvironmentFile)
		if err = env.init(filename); err == nil {
			envFiles.Store(filename, env)
		} else {
			env = nil
		}
	}
	return
}

func (f *EnvironmentFile) init(filename string) (err error) {
	if len(f.filename) == 0 {
		f.filename = filename
	} else {
		return fmt.Errorf("illegal state, file %s already loaded", f.filename)
	}

	if f.fileInfo, err = os.Stat(filename); err != nil {
		return
	}
	if !f.Valid() {
		return fmt.Errorf("%s is not valid", f.filename)
	}

	var file *os.File
	if file, err = os.Open(filename); err == nil {
		defer func() { _ = file.Close() }()
		reader := NewUTF8Reader(file) // accepts UTF8, UTF16 & ISO8859_1
		reader = io.LimitReader(reader, MaxEnvironmentFileContentSize)
		f.env, err = godotenv.Parse(reader)
	}
	return
}

// Name returns the filename of the underlying dotenv file
func (f *EnvironmentFile) Name() string { return f.filename }

// Valid returns true as the loaded values are in-sync with underlying dotenv file
func (f *EnvironmentFile) Valid() bool {
	if s, err := os.Stat(f.filename); err == nil && !s.IsDir() {
		return f.fileInfo.Size() == s.Size() && f.fileInfo.ModTime().Equal(s.ModTime())
	}
	return false
}

// AddTo adds all environment values to the Environment instance
func (f *EnvironmentFile) AddTo(e *Environment) {
	for name, value := range f.env {
		e.Put(e.ResolveName(name), value)
	}
}

// ValuesAsMap returns all environment variables as name & value map
func (f *EnvironmentFile) ValuesAsMap() map[string]string { return maps.Clone(f.env) }
