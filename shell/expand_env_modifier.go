package shell

import (
	"os"
	"strings"

	"github.com/creativeprojects/resticprofile/platform"
)

type ExpandEnvModifier struct {
	env map[string]string
}

var _ ArgModifier = (*ExpandEnvModifier)(nil)

func NewExpandEnvModifier(environment []string) *ExpandEnvModifier {
	envMap := make(map[string]string, len(environment))
	for _, env := range environment {
		envKey, envValue, found := strings.Cut(env, "=")
		if !found {
			continue
		}
		envMap[strings.TrimSpace(envKey)] = strings.TrimSpace(envValue)
	}
	return &ExpandEnvModifier{
		env: envMap,
	}
}

// Arg returns either the same of a new argument if the value was expanded.
// A boolean value indicates if the argument was expanded.
func (m ExpandEnvModifier) Arg(name string, arg *Arg) (*Arg, bool) {
	if arg.HasValue() && arg.Type() == ArgConfigEscape && !platform.IsWindows() {
		if value, changed := m.expandEnv(arg.Value()); changed {
			newArg := arg.Clone()
			newArg.value = value
			return &newArg, true
		}
	}
	return arg, false
}

func (m ExpandEnvModifier) expandEnv(value string) (string, bool) {
	changed := false
	return os.Expand(value, func(key string) string {
		if envValue, found := m.env[key]; found {
			changed = true
			return envValue
		}
		// we don't want to replace the original value
		return "${" + key + "}"
	}), changed
}
