package util

import (
	"strings"

	"github.com/creativeprojects/resticprofile/platform"
	"golang.org/x/exp/maps"
)

// Environment manages a set of environment variables
type Environment struct {
	env          map[string]string
	preserveCase bool
}

// NewDefaultEnvironment creates Environment with OS defaults for preserveCase and the specified initial values
func NewDefaultEnvironment(values ...string) *Environment {
	return NewEnvironment(EnvironmentPreservesCase(), values...)
}

// EnvironmentPreservesCase returns true if environment variables are case sensitive (all OS except Windows)
func EnvironmentPreservesCase() bool { return !platform.IsWindows() }

// NewFoldingEnvironment creates an Environment that folds the case of variable names
func NewFoldingEnvironment(values ...string) *Environment {
	return NewEnvironment(false, values...)
}

// NewEnvironment creates Environment with optional preserveCase and the specified initial values
func NewEnvironment(preserveCase bool, values ...string) *Environment {
	env := &Environment{
		env:          make(map[string]string),
		preserveCase: preserveCase,
	}
	env.SetValues(values...)
	return env
}

func splitEnvironmentValue(keyValue string) (key, value string) {
	if index := strings.Index(keyValue, "="); index > 0 {
		key = strings.TrimSpace(keyValue[:index])
		value = keyValue[index+1:]
	}
	return
}

// SetValues sets one or more values of the format NAME=VALUE
func (e *Environment) SetValues(values ...string) {
	for _, kv := range values {
		if key, _ := splitEnvironmentValue(kv); key != "" {
			if e.preserveCase {
				e.env[key] = kv
			} else {
				e.env[strings.ToUpper(key)] = kv
			}
		}
	}
}

// Values returns all environment variables as NAME=VALUE lines (case is preserved as inserted)
func (e *Environment) Values() (values []string) { return maps.Values(e.env) }

// Names returns all environment variables names (case is preserved as inserted)
func (e *Environment) Names() (names []string) { return maps.Keys(e.ValuesAsMap()) }

// FoldedNames returns all environment variables names (case depends on preserveCase)
func (e *Environment) FoldedNames() (names []string) { return maps.Keys(e.env) }

// ValuesAsMap returns all environment variables as name & value map
func (e *Environment) ValuesAsMap() (m map[string]string) {
	m = make(map[string]string)
	for _, kv := range e.env {
		k, v := splitEnvironmentValue(kv)
		m[k] = v
	}
	return
}

// Put sets a single name and value pair
func (e *Environment) Put(name, value string) {
	if strings.Contains(name, "=") {
		return
	}
	if value == "" {
		e.Remove(name)
	} else {
		e.SetValues(name + "=" + value)
	}
}

// Get returns the variable value, possibly an empty string when the variable is unset or empty.
func (e *Environment) Get(name string) (value string) {
	_, value, _ = e.Find(name)
	return
}

// Has returns true as the variable with name is set.
func (e *Environment) Has(name string) (found bool) {
	_, _, found = e.Find(name)
	return
}

// Find returns the variable's original name, its value and ok when the variable exists
func (e *Environment) Find(name string) (originalName, value string, ok bool) {
	if !e.preserveCase {
		name = strings.ToUpper(name)
	}
	if value, ok = e.env[name]; ok {
		originalName, value = splitEnvironmentValue(value)
	}
	return
}

// Remove deletes the named env variable
func (e *Environment) Remove(name string) {
	if !e.preserveCase {
		name = strings.ToUpper(name)
	}
	delete(e.env, name)
}

// ResolveName resolves the specified name to the actual variable name if case folding applies (preserveCase is false)
func (e *Environment) ResolveName(name string) string {
	if !e.preserveCase {
		if actualName, _, found := e.Find(name); found {
			return actualName
		}
	}
	return name
}
