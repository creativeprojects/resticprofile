package shell

import (
	"regexp"
	"runtime"
	"strings"
)

var (
	escapeNoGlobCharacters = []byte{'|', '&', ';', '<', '>', '(', ')', '$', '`', '\\', '"', '\'', ' ', '\t', '\r', '\n'}
	// escapeGlobCharacters   = []byte{'*', '?', '['}
	doubleQuotePattern = regexp.MustCompile(`[^\\][|&;<>()$'" \t\r\n*?[]`)
)

type ArgType int

const (
	ArgConfigEscape      ArgType = iota // escape each special character but don't add quotes
	ArgConfigNoGlobQuote                // use double quotes around argument when needed
	ArgCommandLineEscape                // same as ArgConfigEscape but argument coming from the command line
	ArgLegacyEscape
	ArgLegacyNoGlobQuote
	ArgLegacyCommandLineEscape
)

type Arg struct {
	raw     string
	argType ArgType
}

func NewArg(raw string, argType ArgType) Arg {
	return Arg{
		raw:     raw,
		argType: argType,
	}
}

func (a Arg) HasValue() bool {
	return a.raw != ""
}

func (a Arg) String() string {
	if runtime.GOOS == "windows" {
		return a.raw
	}
	if !a.HasValue() {
		return ""
	}
	switch a.argType {
	case ArgConfigNoGlobQuote:
		if doubleQuotePattern.MatchString(a.raw) {
			return `"` + escapeString(a.raw, []byte{'"'}) + `"`
		}

	case ArgConfigEscape, ArgCommandLineEscape:
		return escapeString(a.raw, escapeNoGlobCharacters)

	case ArgLegacyEscape:
		return escapeString(a.raw, []byte{' '})

	case ArgLegacyNoGlobQuote:
		return escapeString(a.raw, []byte{' ', '*', '?'})
	}
	return a.raw
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
