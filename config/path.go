package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/clog"
)

type pathFix func(string) string

// fixPath applies all the path fixing callbacks one by one
func fixPath(value string, callbacks ...pathFix) string {
	if len(callbacks) == 0 {
		// nothing to do
		return value
	}
	for _, callback := range callbacks {
		value = callback(value)
	}
	return value
}

// fixPaths runs fixPath over a slice of paths
func fixPaths(sources []string, callbacks ...pathFix) []string {
	fixed := make([]string, len(sources))
	for index, source := range sources {
		fixed[index] = fixPath(source, callbacks...)
	}
	return fixed
}

// resolveGlob evaluates glob expressions in a slice of paths and returns a resolved slice
func resolveGlob(sources []string) []string {
	resolved := make([]string, 0, len(sources))
	for _, source := range sources {
		if strings.ContainsAny(source, "?*") {
			if matches, err := filepath.Glob(source); err == nil {
				resolved = append(resolved, matches...)
			} else {
				clog.Warningf("cannot evaluate glob expression '%s' : %v", source, err)
			}
		} else {
			resolved = append(resolved, source)
		}
	}
	return resolved
}

func expandEnv(value string) string {
	if strings.Contains(value, "$") || strings.Contains(value, "%") {
		value = os.ExpandEnv(value)
	}
	return value
}

func absolutePrefix(prefix string) pathFix {
	return func(value string) string {
		if value == "" ||
			filepath.IsAbs(value) ||
			strings.HasPrefix(value, "~") ||
			strings.HasPrefix(value, "$") ||
			strings.HasPrefix(value, "%") {
			return value
		}
		return filepath.Join(prefix, value)
	}
}

func absolutePath(value string) string {
	if value == "" ||
		filepath.IsAbs(value) ||
		strings.HasPrefix(value, "~") ||
		strings.HasPrefix(value, "$") ||
		strings.HasPrefix(value, "%") {
		return value
	}
	if absolute, err := filepath.Abs(value); err == nil {
		return absolute
	}
	// looks like we can't get an absolute version...
	clog.Errorf("cannot determine absolute path for '%s'", value)
	return value
}
