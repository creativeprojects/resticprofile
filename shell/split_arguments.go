package shell

import (
	"strings"

	"github.com/creativeprojects/resticprofile/platform"
)

// SplitArguments splits a command line string into individual arguments.
// It handles quoted strings and escape characters, with platform-specific behaviour:
// - On non-Windows platforms, backslashes are treated as escape characters
// - Quoted strings are preserved as single arguments
// - Spaces outside quotes are used as argument delimiters
//
// Example:
//
//	SplitArguments(`echo "Hello World" file\ with\ spaces`)
//	// Returns: []string{"echo", "Hello World", "file with spaces"}
func SplitArguments(commandLine string) []string {
	args := make([]string, 0)
	if len(commandLine) == 0 {
		return args
	}
	sb := &strings.Builder{}
	quoted := false
	escaped := false
	for _, r := range commandLine {
		if r == '\\' && !escaped && !platform.IsWindows() {
			escaped = true
		} else if r == '"' && !escaped {
			quoted = !quoted
			escaped = false
		} else if !quoted && !escaped && r == ' ' {
			args = append(args, sb.String())
			sb.Reset()
		} else {
			sb.WriteRune(r)
			escaped = false
		}
	}
	if sb.Len() > 0 {
		args = append(args, sb.String())
	}
	return args
}
