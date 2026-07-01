package shell

type ArgModifier interface {
	// Arg returns either the same or a new Arg if the value was changed.
	// A boolean also indicates if the value was changed.
	Arg(name string, arg *Arg) (*Arg, bool)
}
