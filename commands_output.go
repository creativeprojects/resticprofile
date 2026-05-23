package main

import (
	"strings"
)

// wantsStructuredOutput reports whether args request a machine-readable
// output format via --output=<format> (anything other than plain). Used to
// route diagnostic logs to stderr so stdout stays parseable for tools like
// jq. Format validation is left to parseOutputFormat; an unrecognised value
// is still treated as "structured" here so logs do not leak before the
// command errors out.
func wantsStructuredOutput(args []string) bool {
	value, ok := findOutputValue(args)
	if !ok {
		return false
	}
	return value != "" && value != "plain"
}

func parseOutputFormat(args []string) string {
	format, found := findOutputValue(args)
	if !found {
		return "plain"
	}
	return format
}

func findOutputValue(args []string) (string, bool) {
	const outputFlag = "--output"
	value := ""
	found := false
	for i, arg := range args {
		if v, ok := strings.CutPrefix(arg, outputFlag+"="); ok {
			value, found = v, true
			continue
		}
		if arg == outputFlag && i+1 < len(args) {
			value, found = args[i+1], true
		}
	}
	return value, found
}
