package shell

import "strings"

func SplitArguments(commandLine string) []string {
	args := make([]string, 0)
	sb := &strings.Builder{}
	quoted := false
	for _, r := range commandLine {
		escape := false
		if r == '\\' {
			sb.WriteRune(r)
			escape = true
		} else if r == '"' && !escape {
			quoted = !quoted
		} else if !quoted && r == ' ' {
			args = append(args, sb.String())
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
	}
	if sb.Len() > 0 {
		args = append(args, sb.String())
	}
	return args
}
