package shell

import (
	"fmt"
	"strings"
)

// Literal values (resolved backup sources) must reach the command exactly as they are, whatever the shell.
// They travel from Args to the shell composer between two NUL characters: NUL cannot appear
// in file names, environment variables or command line arguments, so it cannot be forged.
const (
	literalMarker    = "\x00"
	literalEnvPrefix = "RESTICPROFILE_LITERAL_"
)

func markLiteral(value string) string {
	return literalMarker + value + literalMarker
}

// resolveLiterals replaces every marked literal value inside arg with render(value)
func resolveLiterals(arg string, render func(value string) string) string {
	if !strings.Contains(arg, literalMarker) {
		return arg
	}
	sb := strings.Builder{}
	for i, part := range strings.Split(arg, literalMarker) {
		if i%2 == 1 {
			sb.WriteString(render(part))
		} else {
			sb.WriteString(part)
		}
	}
	return sb.String()
}

// literalEnviron passes literal values through environment variables:
// shells never parse the value of a variable they expand.
type literalEnviron struct {
	environ []string
}

func (l *literalEnviron) add(value string) (name string) {
	name = fmt.Sprintf("%s%d", literalEnvPrefix, len(l.environ)+1)
	l.environ = append(l.environ, name+"="+value)
	return
}

// quotePosixLiteral single-quotes value for sh and bash
func quotePosixLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// escapeWindowsArg quotes value for the Windows command line parser (same rules as syscall.EscapeArg)
func escapeWindowsArg(value string) string {
	if value == "" {
		return `""`
	}
	if !strings.ContainsAny(value, " \t\"") {
		return value
	}
	sb := strings.Builder{}
	sb.WriteByte('"')
	backslashes := 0
	for i := range len(value) {
		switch c := value[i]; c {
		case '\\':
			backslashes++
		case '"':
			sb.WriteString(strings.Repeat(`\`, backslashes*2+1))
			sb.WriteByte(c)
			backslashes = 0
		default:
			sb.WriteString(strings.Repeat(`\`, backslashes))
			sb.WriteByte(c)
			backslashes = 0
		}
	}
	// backslashes before the closing quote must be doubled
	sb.WriteString(strings.Repeat(`\`, backslashes*2))
	sb.WriteByte('"')
	return sb.String()
}
