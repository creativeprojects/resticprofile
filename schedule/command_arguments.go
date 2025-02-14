package schedule

import (
	"slices"
	"strings"
)

type CommandArguments struct {
	args []string
}

func NewCommandArguments(args []string) CommandArguments {
	return CommandArguments{
		args: args,
	}
}

// Trim returns a new CommandArguments with the specified flags removed from the arguments
func (ca CommandArguments) Trim(removeArgs []string) CommandArguments {
	args := make([]string, 0, len(ca.args))
	for _, arg := range ca.args {
		if slices.Contains(removeArgs, arg) {
			continue
		}
		args = append(args, arg)
	}
	return NewCommandArguments(args)
}

func (ca CommandArguments) RawArgs() []string {
	result := make([]string, len(ca.args))
	copy(result, ca.args)
	return result
}

// String returns the arguments as a string, with quotes around arguments that contain spaces
func (ca CommandArguments) String() string {
	if len(ca.args) == 0 {
		return ""
	}

	var n int
	for _, elem := range ca.args {
		n += len(elem) + 3 // add 2 if quotes are needed, plus 1 for the space
	}

	b := new(strings.Builder)
	b.Grow(n)
	ca.writeString(b, ca.args[0])
	for _, s := range ca.args[1:] {
		b.WriteString(" ")
		ca.writeString(b, s)
	}
	return b.String()
}

// ConfigFile returns the value of the --config argument, if present.
// if multiple --config are present, it will return the last value
func (ca CommandArguments) ConfigFile() string {
	if len(ca.args) == 0 {
		return ""
	}
	const configFlag = "--config"
	const configPrefix = configFlag + "="

	var lastConfig string
	for i, arg := range ca.args {
		if arg == configFlag {
			if i+1 < len(ca.args) {
				lastConfig = ca.args[i+1]
			}
		} else if strings.HasPrefix(arg, configPrefix) {
			lastConfig = strings.TrimPrefix(arg, configPrefix)
		}
	}
	return lastConfig
}

func (ca CommandArguments) writeString(b *strings.Builder, str string) {
	if strings.Contains(str, " ") {
		b.WriteString(`"`)
		b.WriteString(str)
		b.WriteString(`"`)
	} else {
		b.WriteString(str)
	}
}
