package term

import (
	"os"
	"sync/atomic"

	"golang.org/x/term"
)

var defaultTerminal atomic.Pointer[Terminal]

// Get returns the default terminal. It will be initialized on first use with NewTerminal() if not set before.
// Ideally we should pass the terminal where we need it, but Get() can be safely used until the refactoring is finished.
func Get() *Terminal {
	defaultTerminal.CompareAndSwap(nil, NewTerminal())
	return defaultTerminal.Load()
}

// Set stores the default terminal, and returns the terminal reference for chaining.
func Set(t *Terminal) *Terminal {
	defaultTerminal.Store(t)
	return t
}

// Size returns the width and height of the terminal session
func Size() (width, height int) {
	fd := fdToInt(os.Stdout.Fd())
	var err error
	width, height, err = term.GetSize(fd)
	if err != nil {
		width, height = 0, 0
	}
	return
}
