package shell

import (
	"os"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/platform"
)

type ExpandEnvModifier struct {
	env     map[string]string
	noShell bool
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

// NewNoShellExpandEnvModifier expands environment variables for a command that doesn't run through a shell:
// it works on any OS, and leaves undefined variables as written.
func NewNoShellExpandEnvModifier(environment []string) *ExpandEnvModifier {
	modifier := NewExpandEnvModifier(environment)
	modifier.noShell = true
	return modifier
}

// Arg returns either the same of a new argument if the value was expanded.
// A boolean value indicates if the argument was expanded.
func (m ExpandEnvModifier) Arg(name string, arg *Arg) (*Arg, bool) {
	if !arg.HasValue() || arg.Type() != ArgConfigEscape || (platform.IsWindows() && !m.noShell) {
		return arg, false
	}
	expand := m.expandEnv
	if m.noShell {
		expand = m.expandDefinedEnv
	}
	if value, changed := expand(arg.Value()); changed {
		newArg := arg.Clone()
		newArg.value = value
		return &newArg, true
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

var envVariablePattern = regexp.MustCompile(`\$\{(\w+)}|\$(\w+)`)

// expandDefinedEnv replaces "$VAR" and "${VAR}" when VAR is defined, anything else is left as written
func (m ExpandEnvModifier) expandDefinedEnv(value string) (string, bool) {
	changed := false
	value = envVariablePattern.ReplaceAllStringFunc(value, func(variable string) string {
		groups := envVariablePattern.FindStringSubmatch(variable)
		if envValue, found := m.env[groups[1]+groups[2]]; found {
			changed = true
			return envValue
		}
		return variable
	})
	return value, changed
}
