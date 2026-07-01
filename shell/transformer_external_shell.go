package shell

import "github.com/creativeprojects/resticprofile/platform"

type ExternalShellTransformer struct{}

func (t ExternalShellTransformer) Transform(name string, arg Arg) ([]Arg, bool) {
	if platform.IsWindows() {
		return []Arg{arg}, false
	}
	value := arg.Value()
	switch arg.argType {
	case ArgConfigKeepGlobQuote:
		if doubleQuotePattern.MatchString(value) {
			value = `"` + escapeString(value, []byte{'"'}) + `"`
		}

	case ArgConfigEscape, ArgCommandLineEscape, ArgConfigBackupSource:
		value = escapeString(value, escapeNoGlobCharacters)

	// legacy mode was a mess: 4 different ways of escaping arguments!
	case ArgLegacyEscape:
		value = quoteArgument(escapeString(value, []byte{' '}))

	case ArgLegacyKeepGlobQuote:
		value = quoteArgument(escapeString(value, []byte{' ', '*', '?'}))

	case ArgLegacyConfigBackupSource:
		value = escapeString(value, []byte{' '})
	}
	if value == arg.Value() {
		// value hasn't changed
		return []Arg{arg}, false
	}
	newArg := arg.Clone()
	newArg.value = value
	return []Arg{newArg}, true
}

var _ ArgTransformer = &ExternalShellTransformer{}
