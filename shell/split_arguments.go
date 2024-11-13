package shell

import (
	"strings"

	"github.com/creativeprojects/resticprofile/platform"
)

func SplitArguments(commandLine string) []string {
	args := make([]string, 0)
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
