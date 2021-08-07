package shell

import (
	"fmt"
	"sort"
)

type Args struct {
	args   map[string][]Arg
	more   []Arg
	legacy bool
}

func NewArgs() *Args {
	return &Args{
		args: make(map[string][]Arg, 10),
		more: make([]Arg, 0, 10),
	}
}

// SetLegacyArg is used to activate the legacy (broken) mode of sending arguments on the restic command line
func (a *Args) SetLegacyArg(legacy bool) *Args {
	a.legacy = legacy
	return a
}

func (a *Args) addLegacy(argType ArgType) ArgType {
	if a.legacy && argType <= ArgCommandLineEscape {
		argType += 3
	}
	return argType
}

// AddFlag adds a value to a flag
func (a *Args) AddFlag(key, value string, argType ArgType) {
	a.args[key] = []Arg{NewArg(value, a.addLegacy(argType))}
}

// AddFlags adds a slice of values for the same flag
func (a *Args) AddFlags(key string, values []string, argType ArgType) {
	args := make([]Arg, len(values))
	for i, value := range values {
		args[i] = NewArg(value, a.addLegacy(argType))
	}
	a.args[key] = args
}

// AddArg adds a single argument with no flag
func (a *Args) AddArg(arg string, argType ArgType) {
	a.more = append(a.more, NewArg(arg, a.addLegacy(argType)))
}

// AddArgs adds multiple arguments not associated with a flag
func (a *Args) AddArgs(args []string, argType ArgType) {
	for _, arg := range args {
		a.more = append(a.more, NewArg(arg, a.addLegacy(argType)))
	}
}

// ToMap converts the arguments to a map.
// It is only used by unit tests.
func (a *Args) ToMap() map[string][]string {
	output := make(map[string][]string, len(a.args))
	for key, values := range a.args {
		strValues := make([]string, len(values))
		for i, value := range values {
			strValues[i] = value.String()
		}
		output[key] = strValues
	}
	return output
}

func (a *Args) Get(name string) ([]Arg, bool) {
	arg, ok := a.args[name]
	return arg, ok
}

// GetAll return a clean list of arguments to send on the command line
func (a *Args) GetAll() []string {
	args := make([]string, 0, len(a.args)+len(a.more)+10)

	if len(a.args) == 0 && len(a.more) == 0 {
		return args
	}

	// we make a list of keys first, so we can loop on the map from an ordered list of keys
	keys := make([]string, 0, len(a.args))
	for key := range a.args {
		keys = append(keys, key)
	}
	// sort the keys in order
	sort.Strings(keys)

	// now we loop from the ordered list of keys
	for _, key := range keys {
		values := a.args[key]
		if values == nil {
			continue
		}
		if len(values) == 0 {
			args = append(args, fmt.Sprintf("--%s", key))
			continue
		}
		for _, value := range values {
			args = append(args, fmt.Sprintf("--%s", key))
			if value.HasValue() {
				args = append(args, value.String())
			}
		}
	}

	// and the list of flat arguments
	for _, arg := range a.more {
		args = append(args, arg.String())
	}
	return args
}
