package shell

import (
	"fmt"
	"sort"

	"github.com/creativeprojects/resticprofile/util/collect"
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

func (a *Args) Clone() *Args {
	clone := NewArgs()
	for name, args := range a.args {
		if args != nil {
			args = append(make([]Arg, 0, len(args)), args...)
		}
		clone.args[name] = args
	}
	clone.more = append(clone.more, a.more...)
	clone.legacy = a.legacy
	return clone
}

func (a *Args) Walk(callback func(name string, arg *Arg) *Arg) {
	processArgs := func(name string, args []Arg) {
		for i, arg := range args {
			if newArg := callback(name, &arg); newArg != nil && newArg != &arg {
				args[i] = *newArg
			}
		}
	}
	for name, args := range a.args {
		processArgs(name, args)
	}
	processArgs("", a.more)
}

// SetLegacyArg is used to activate the legacy (broken) mode of sending arguments on the restic command line
func (a *Args) SetLegacyArg(legacy bool) *Args {
	a.legacy = legacy
	return a
}

func (a *Args) addLegacy(argType ArgType) ArgType {
	if a.legacy && argType <= ArgConfigBackupSource {
		argType += ArgTypeCount
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

func (a *Args) Remove(name string) ([]Arg, bool) {
	arg, ok := a.Get(name)
	delete(a.args, name)
	return arg, ok
}

func (a *Args) RemoveArg(name string) (removed []Arg) {
	nameMatch := func(t Arg) bool { return t.Value() == name }
	removed = collect.All(a.more, nameMatch)
	a.more = collect.All(a.more, collect.Not(nameMatch))
	return
}

func (a *Args) Rename(oldName, newName string) bool {
	args, ok := a.Remove(oldName)
	if ok {
		for _, arg := range args {
			a.AddFlag(newName, arg.Value(), arg.Type())
		}
	}
	args = a.RemoveArg(oldName)
	ok = ok || len(args) > 0
	for _, arg := range args {
		a.AddArg(newName, arg.Type())
	}
	return ok
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
			if value.HasValue() {
				args = append(args, fmt.Sprintf("--%s=%s", key, value.String())) // must use "=" as some values (e.g. verbose) need this to work correctly
			} else {
				args = append(args, fmt.Sprintf("--%s", key))
			}
		}
	}

	// and the list of flat arguments
	for _, arg := range a.more {
		args = append(args, arg.String())
	}
	return args
}
