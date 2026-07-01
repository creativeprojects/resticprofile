package shell

type ArgTransformer interface {
	// Arg returns either the same or new Args if the value was changed.
	// A boolean also indicates if the value was changed.
	Transform(name string, arg Arg) ([]Arg, bool)
}
