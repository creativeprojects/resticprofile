package main

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

const (
	configTag      = "```"
	checkdocIgnore = "<!-- checkdoc-ignore -->"
	goTemplate     = "{{"
	replaceURL     = "http://localhost:1313/resticprofile/jsonschema/config$1.json"
)

var (
	urlPattern = regexp.MustCompile(`{{< [^>}]+config(-\d)?\.json"[^>}]+ >}}`)
	_          = regexp.MustCompile(`{{< [^>}]+config(-\d)?\.json"[^>}]+ >}}`) // Remove this when VS Code fixed the syntax highlighting issues
)

var (
	tempDir string
)

// this small script walks through files and picks all the .md ones (containing documentation)
// then it parses the .md files to see if there's some configuration examples, as they should be starting
// with the tag ``` followed by the types: yaml, json, toml and hcl
// once a configuration snippet has been detected, it tries to detect and load the profiles to see if there's any error
func main() {
	exitCode := 0

	var root string
	var verbose bool
	var ignoreFiles []string
	pflag.StringVarP(&root, "root", "r", "", "root directory where to search for documentation files (*.md)")
	pflag.StringVarP(&tempDir, "temp-dir", "t", "", "temporary directory to store extracted configuration files")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "display more information")
	pflag.StringSliceVarP(&ignoreFiles, "ignore", "i", nil, "ignore files")
	pflag.Parse()

	level := clog.LevelInfo
	if verbose {
		level = clog.LevelDebug
	}
	clog.SetDefaultLogger(clog.NewFilteredConsoleLogger(level))
	clog.Info("checking documentation for configuration examples")

	deleteFunc := setupTempDir()
	defer deleteFunc()
	clog.Debugf("using temporary directory %q", tempDir)

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
		simplePath := strings.TrimPrefix(path, wd)
		for _, ignore := range ignoreFiles {
			if strings.Contains(simplePath, ignore) {
				clog.Infof("* ignoring file %s", simplePath)
				return nil
			}
		}
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		if ext != ".md" {
			return nil
		}
		clog.Infof("* file %s", simplePath)
		if !extractConfigurationSnippets(path) {
			exitCode = 1
		}
		return nil
	})
	os.Exit(exitCode)
}

// extractConfigurationSnippets returns true when the configuration is valid
func extractConfigurationSnippets(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		clog.Error("error reading %q: %s", path, err)
		return false
	}
	defer file.Close()

	hasError := false
	ignoreError := false
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
					// end of configuration snippet
					configLines = false
					// finished reading a configuration, save the buffer for checking
					if !ignoreError {
						filename, err := saveConfiguration(configBuffer.Bytes(), configType)
						defer os.Remove(filename)
						if err != nil {
							clog.Errorf("cannot save configuration: %s", err)
							return
						}
						if !checkConfiguration(filename, configType, lineNum) {
							hasError = true
						}
					}
					ignoreError = false
					clog.Debugf(" - end of %q block on line %d", configType, lineNum)
					return
				}
				configType = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), configTag))
				if configType == "toml" || configType == "json" || configType == "yaml" || configType == "hcl" {
					// start of configuration snippet
					configLines = true
					configBuffer.Reset()
					ignored := ""
					if ignoreError {
						ignored = "(ignored)"
					}
					clog.Debugf(" - start of %q block on line %d %s", configType, lineNum, ignored)
				}
				return
			}
			if configLines {
				// inside a configuration snippet
				// replace hugo tags
				if strings.Contains(line, goTemplate) {
					line = urlPattern.ReplaceAllString(line, replaceURL)
				}
				configBuffer.WriteString(line)
				configBuffer.WriteString("\n")
			}
			if line == checkdocIgnore {
				ignoreError = true
			}
		}(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		clog.Errorf("error scanning file %q: %w", path, err)
		return false
	}
	return !hasError
}

func saveConfiguration(content []byte, configType string) (string, error) {
	filename := filepath.Join(tempDir, "profiles."+configType)
	err := os.WriteFile(filename, content, 0600)
	return filename, err
}

// checkConfiguration returns true when the configuration is valid
func checkConfiguration(filename, configType string, lineNum int) bool {
	cfg, err := config.LoadFile(afero.NewOsFs(), filename, configType)
	if err != nil {
		clog.Errorf("    %q on line %d: %s", configType, lineNum, err)
		return false
	}
	if cfg == nil {
		clog.Errorf("    %q on line %d: configuration is empty", configType, lineNum)
		return false
	}
	if configType != "hcl" && !cfg.IsSet("version") {
		clog.Infof("    %q on line %d: missing 'version' option in configuration", configType, lineNum)
	}
	if configType == "hcl" && cfg.IsSet("version") {
		clog.Infof("    %q on line %d: configuration has the 'version' option specified", configType, lineNum)
	}
	profiles := cfg.GetProfileNames()
	if len(profiles) == 0 && !cfg.IsSet("global") {
		clog.Warningf("    %q on line %d: configuration has no profile", configType, lineNum)
	}
	for _, profileName := range profiles {
		_, err := cfg.GetProfile(profileName)
		if err != nil {
			clog.Errorf("    %q on line %d: profile %s: %s", configType, lineNum, profileName, err)
			return false
		}
	}
	return true
}

// setupTempDir is using the tempDir global variable
func setupTempDir() func() {
	if tempDir != "" {
		_ = os.MkdirAll(tempDir, 0700)
		return func() {
			// nothing to do
		}
	}
	tempDir, _ = os.MkdirTemp("", "checkdoc*")
	return func() {
		_ = os.RemoveAll(tempDir)
	}
}
