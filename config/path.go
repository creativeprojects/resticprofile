package config

import (
	"os"
	"path/filepath"
	"runtime"
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

func expandEnv(value string) string {
	if strings.Contains(value, "$") || strings.Contains(value, "%") {
		value = os.ExpandEnv(value)
	}
	return value
}

// escapeSpaces escapes ' ' characters (unix only)
func escapeSpaces(value string) string {
	if runtime.GOOS != "windows" {
		value = escapeString(value, []byte{' '})
	}
	return value
}

// escapeShellString escapes ' ', '*' and '?' characters (unix only)
func escapeShellString(value string) string {
	if runtime.GOOS != "windows" {
		value = escapeString(value, []byte{' ', '*', '?'})
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

// escapeString adds a '\' in front of the characters to escape.
// it checks for the number of '\' characters in front:
// - if even: add one
// - if odd: do nothing, it means the character is already escaped
func escapeString(value string, chars []byte) string {
	output := &strings.Builder{}
	escape := 0
	for i := 0; i < len(value); i++ {
		if value[i] == '\\' {
			escape++
		} else {
			for _, char := range chars {
				if value[i] == char {
					if escape%2 == 0 {
						// even number of escape characters in front, we need to escape this one
						output.WriteByte('\\')
					}
				}
			}
			// reset number of '\'
			escape = 0
		}
		output.WriteByte(value[i])
	}
	return output.String()
}
