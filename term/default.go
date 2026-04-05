package term

var defaultTerminal *Terminal

// Get returns the default terminal. It will be initialized on first use with NewTerminal() if not set before.
func Get() *Terminal {
	if defaultTerminal == nil {
		defaultTerminal = NewTerminal()
	}
	return defaultTerminal
}

// Set stores the default terminal, and returns the terminal reference for chaining.
func Set(t *Terminal) *Terminal {
	defaultTerminal = t
	return defaultTerminal
}
