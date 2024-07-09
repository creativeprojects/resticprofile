package shell

type ConfidentialArgModifier struct {
}

var _ ArgModifier = (*ConfidentialArgModifier)(nil)

func NewConfidentialArgModifier() *ConfidentialArgModifier {
	return &ConfidentialArgModifier{}
}

// Arg returns either the same of a new argument if the type has changed.
// A boolean value indicates if Arg has changed.
func (m ConfidentialArgModifier) Arg(name string, arg *Arg) (*Arg, bool) {
	if arg.HasConfidentialFilter() {
		newArg := arg.Clone()
		newArg.value = arg.GetConfidentialValue()
		newArg.confidentialFilter = nil
		return &newArg, true
	}
	return arg, false
}
