package main

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/spf13/pflag"
)

const (
	configTag = "```"
)

// this small script walks through files and picks all the .md ones (containing documentation)
// then it parses the .md files to see if there's some configuration examples, as they should be starting
// with the tag ``` followed by the types: yaml, json, toml and hcl
// once a configuration snippet has been detected, it tries to load it to see if there's no error in it ;)
func main() {
	exitCode := 0
	clog.SetDefaultLogger(clog.NewFilteredConsoleLogger(clog.LevelDebug))
	clog.Info("checking documentation for configuration examples")

	var root string
	pflag.StringVarP(&root, "root", "r", "", "root directory where to search for documentation files (*.md)")
	pflag.Parse()

	// if there's an error here, wd is going to be empty, and that's ok
	wd, _ := os.Getwd()
	if !filepath.IsAbs(root) {
		root = filepath.Join(wd, root)
	}

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		if ext != ".md" {
			return nil
		}
		simplePath := strings.TrimPrefix(path, wd)
		clog.Infof("* file %s", simplePath)
		if !findConfiguration(path) {
			exitCode = 1
		}
		return nil
	})
	os.Exit(exitCode)
}

func findConfiguration(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		clog.Error("error reading %q: %s", path, err)
		return false
	}
	defer file.Close()

	noError := true
	configType := ""
	configLines := false
	configBuffer := &bytes.Buffer{}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		// wrap it in a func so we can "return" early
		func(line string) {
			if strings.HasPrefix(line, configTag) {
				if configLines {
					configLines = false
					clog.Debugf(" - end of %q block on line %d", configType, lineNum)
					// finished reading a configuration, send the buffer for checking
					cfg, err := config.Load(configBuffer, configType)
					if err != nil {
						clog.Error(err)
						noError = false
					}
					if cfg == nil {
						clog.Errorf("empty %s configuration", configType)
						noError = false
					}
					return
				}
				configType = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), configTag))
				if configType == "toml" || configType == "json" || configType == "yaml" || configType == "hcl" {
					configLines = true
					configBuffer.Reset()
					clog.Debugf(" - start of %q block on line %d", configType, lineNum)
				}
				return
			}
			if configLines {
				configBuffer.WriteString(line)
				configBuffer.Write([]byte("\n"))
			}
		}(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		clog.Errorf("error scanning file %q: %w", path, err)
		return false
	}
	return noError
}
