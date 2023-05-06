package config

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/creativeprojects/clog"
)

type pathFix func(string) string

// fixPath applies all the path fixing callbacks one by one
func fixPath(value string, callbacks ...pathFix) string {
	for _, callback := range callbacks {
		if callback != nil {
			value = callback(value)
		}
	}
	return value
}

// fixPaths runs fixPath over a slice of paths
func fixPaths(sources []string, callbacks ...pathFix) (fixed []string) {
	if len(sources) > 0 {
		fixed = make([]string, len(sources))
		for index, source := range sources {
			fixed[index] = fixPath(source, callbacks...)
		}
	}
	return
}

func expandEnv(value string) string {
	if strings.Contains(value, "$") || strings.Contains(value, "%") {
		value = os.Expand(value, func(name string) string {
			if name == "$" {
				return "$" // allow to escape "$" as "$$"
			}
			return os.Getenv(name)
		})
	}
	return value
}

func isUserHomePath(value string) bool {
	return value != expandUserHome(value)
}

func expandUserHome(value string) string {
	// "~", "~/path" and "~unix-user/" but not "~somefile"
	if strings.HasPrefix(value, "~") {
		delimiter := strings.IndexAny(value, `/\`)
		if delimiter < 0 {
			delimiter = len(value)
		}

		if delimiter == 1 {
			if path, err := os.UserHomeDir(); err == nil {
				value = path + value[1:]
			} else {
				clog.Warningf("cannot resolve user home dir for path %s : %v", value, err)
			}
		} else if runtime.GOOS != "windows" { // Windows uses "domain\user", skipping User lookup for this OS
			username := value[1:delimiter]

			if u, err := user.Lookup(username); err == nil {
				value = u.HomeDir + value[delimiter:]
			} else if !errors.Is(err, user.UnknownUserError(username)) {
				clog.Warningf("cannot resolve user home dir for user %s in path %s : %v", username, value, err)
			}
		}
	}

	return value
}

func absolutePrefix(prefix string) pathFix {
	return func(value string) string {
		if value == "" ||
			filepath.IsAbs(value) ||
			isUserHomePath(value) ||
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
		isUserHomePath(value) ||
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

// resolveGlob evaluates glob expressions in a slice of paths and returns a resolved slice
func resolveGlob(sources []string) (resolved []string) {
	if len(sources) > 0 {
		resolved = make([]string, 0, len(sources))
		for _, source := range sources {
			if strings.ContainsAny(source, "?*[") {
				if matches, err := filepath.Glob(source); err == nil {
					resolved = append(resolved, matches...)
				} else {
					clog.Warningf("cannot evaluate glob expression '%s' : %v", source, err)
				}
			} else {
				resolved = append(resolved, source)
			}
		}
	}
	return
}
