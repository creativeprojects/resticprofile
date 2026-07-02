package shell

import (
	"os"
	"strings"

	"github.com/creativeprojects/clog"
	"mvdan.cc/sh/v3/expand"
)

type Env struct {
	store map[string]expand.Variable
}

func NewEnv() *Env {
	return &Env{
		store: make(map[string]expand.Variable, 100),
	}
}

// AddEnv adds the existing environment variables to the runner.
// The parameter should contain a list of variable names.
func (e *Env) AddEnv(names []string) *Env {
	for _, name := range names {
		if value, found := os.LookupEnv(name); found {
			e.store[name] = newVariable(value)
		}
	}
	return e
}

// AddEnviron adds a list of strings representing the environment, in the form "key=value".
func (e *Env) AddEnviron(environ []string) *Env {
	for _, pair := range environ {
		if key, value, found := strings.Cut(pair, "="); found {
			e.store[key] = newVariable(value)
			continue
		}
		clog.Warningf("invalid environment variable key=value pair: %q", pair)
	}
	return e
}

func (e *Env) Get(name string) expand.Variable {
	if variable, found := e.store[name]; found {
		return variable
	}
	return expand.Variable{}
}

func (e *Env) Each(callback func(name string, vr expand.Variable) bool) {
	for key, value := range e.store {
		next := callback(key, value)
		if !next {
			return
		}
	}
}

func newVariable(value string) expand.Variable {
	return expand.Variable{
		Set:      true,
		Exported: true,
		Kind:     expand.String,
		Str:      value,
	}
}
