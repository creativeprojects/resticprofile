package shell

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/platform"
)

var (
	escapeNoGlobCharacters = []byte{'|', '&', ';', '<', '>', '(', ')', '$', '`', '\\', '"', '\'', ' ', '\t', '\r', '\n'}
	doubleQuotePattern     = regexp.MustCompile(`[^\\][|&;<>()$'" \t\r\n*?[]`)
)

type ArgType int

const ArgTypeCount = 4

const (
	ArgConfigEscape        ArgType = iota // escape each special character but don't add quotes
	ArgConfigKeepGlobQuote                // use double quotes around argument when needed so the shell doesn't resolve glob patterns
	ArgCommandLineEscape                  // same as ArgConfigEscape but argument is coming from the command line
	ArgConfigBackupSource                 // same as ArgConfigEscape but represents the folders to add at the end of a backup command
	ArgLegacyEscape
	ArgLegacyKeepGlobQuote
	ArgLegacyCommandLineEscape
	ArgLegacyConfigBackupSource
)

type filterFunc func(string) string

type Arg struct {
	value              string
	argType            ArgType
	empty              bool
	confidentialFilter filterFunc
}

func NewArg(value string, argType ArgType, options ...ArgOption) Arg {
	arg := Arg{
		value:              value,
		argType:            argType,
		empty:              false,
		confidentialFilter: nil,
	}
	for _, option := range options {
		option.setup(&arg)
	}
	return arg
}

func NewArgsSlice(args []string, argType ArgType, options ...ArgOption) []Arg {
	result := make([]Arg, len(args))
	for i, arg := range args {
		result[i] = NewArg(arg, argType, options...)
	}
	return result
}

func NewEmptyValueArg() Arg {
	return NewArg("", ArgConfigKeepGlobQuote, &EmptyArgOption{})
}

func (a Arg) Clone() Arg {
	return Arg{
		value:              a.value,
		argType:            a.argType,
		empty:              a.empty,
		confidentialFilter: a.confidentialFilter,
	}
}

// IsEmptyValue means the flag is specifically empty, not just a flag without a value
// (e.g. --flag="")
func (a Arg) IsEmptyValue() bool {
	return a.empty
}

// HasValue returns true if the argument has a value (e.g. --flag=value).
// The value could be empty (e.g. --flag="").
// If false, the argument is a simple flag (e.g. --verbose).
func (a Arg) HasValue() bool {
	return a.empty || a.value != ""
}

// HasConfidentialFilter returns true if the argument may contain credentials.
func (a Arg) HasConfidentialFilter() bool {
	return a.confidentialFilter != nil
}

func (a Arg) Value() string {
	return a.value
}

func (a Arg) Type() ArgType {
	return a.argType
}

func (a Arg) GetConfidentialValue() string {
	if a.HasConfidentialFilter() {
		return a.confidentialFilter(a.value)
	}
	return a.value
}

// String returns an escaped value to send to the command line
func (a Arg) String() string {
	if !a.HasValue() {
		return ""
	}

	value := a.Value()
	if value == "" {
		return `""`
	}

	if !platform.IsWindows() {
		switch a.argType {
		case ArgConfigKeepGlobQuote:
			if doubleQuotePattern.MatchString(value) {
				return `"` + escapeString(value, []byte{'"'}) + `"`
			}

		case ArgConfigEscape, ArgCommandLineEscape, ArgConfigBackupSource:
			return escapeString(value, escapeNoGlobCharacters)

		// legacy mode was a mess: 4 different ways of escaping arguments!
		case ArgLegacyEscape:
			return quoteArgument(escapeString(value, []byte{' '}))

		case ArgLegacyKeepGlobQuote:
			return quoteArgument(escapeString(value, []byte{' ', '*', '?'}))

		case ArgLegacyConfigBackupSource:
			return escapeString(value, []byte{' '})
		}
	}

	return value
}

// escapeString adds a '\' in front of the characters to escape.
// it checks for the number of '\' characters in front:
// - if even: add one
// - if odd: do nothing, it means the character is already escaped
func escapeString(value string, chars []byte) string {
	output := &strings.Builder{}
	escape := 0
	for i := 0; i < len(value); i++ {
		if value[i] == '\\' {
			escape++
		} else {
			for _, char := range chars {
				if value[i] == char {
					if escape%2 == 0 {
						// even number of escape characters in front, we need to escape this one
						output.WriteByte('\\')
					}
				}
			}
			// reset number of '\'
			escape = 0
		}
		output.WriteByte(value[i])
	}
	return output.String()
}

// quoteArgument is used for the legacy mode only
func quoteArgument(value string) string {
	if strings.Contains(value, " ") {
		// quote the string containing spaces
		value = fmt.Sprintf(`"%s"`, value)
	}
	return value
}
