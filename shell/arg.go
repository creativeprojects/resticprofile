package shell

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/platform"
)

var (
	escapeNoGlobCharacters = []byte{'|', '&', ';', '<', '>', '(', ')', '$', '`', '\\', '"', '\'', ' ', '\t', '\r', '\n'}
	// escapeGlobCharacters   = []byte{'*', '?', '['}
	doubleQuotePattern = regexp.MustCompile(`[^\\][|&;<>()$'" \t\r\n*?[]`)
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

var emptyArgValueMarker = func() string {
	token := make([]byte, 16)
	n, err := rand.Read(token)
	if err == nil && n != 16 {
		err = fmt.Errorf("insufficient random bytes %d", n)
	}
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(token)
}()

func EmptyArgValue() string {
	return emptyArgValueMarker
}

func NewEmptyValueArg() Arg {
	return NewArg(EmptyArgValue(), ArgConfigKeepGlobQuote)
}

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

func (a Arg) Value() string {
	if a.raw == EmptyArgValue() {
		return ""
	}
	return a.raw
}

func (a Arg) Type() ArgType {
	return a.argType
}

func (a Arg) String() string {
	value := a.Value()

	if !a.HasValue() {
		return ""
	}

	if a.Value() == "" {
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
