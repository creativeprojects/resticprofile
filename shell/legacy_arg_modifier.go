package shell

// LegacyArgModifier changes the type of arguments to a legacy type.
type LegacyArgModifier struct {
	legacy bool
}

var _ ArgModifier = (*LegacyArgModifier)(nil)

func NewLegacyArgModifier(legacy bool) *LegacyArgModifier {
	return &LegacyArgModifier{
		legacy: legacy,
	}
}

// Arg returns either the same of a new argument if the type has changed.
// A boolean value indicates if Arg has changed.
func (m LegacyArgModifier) Arg(name string, arg *Arg) (*Arg, bool) {
	if m.legacy {
		if newArgType, changed := addLegacy(arg.Type()); changed {
			newArg := arg.Clone()
			newArg.argType = newArgType
			return &newArg, true
		}
	}
	return arg, false
}

func addLegacy(argType ArgType) (ArgType, bool) {
	if argType <= ArgConfigBackupSource {
		argType += ArgTypeCount
		return argType, true
	}
	return argType, false
}
