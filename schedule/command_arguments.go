package schedule

import "strings"

type CommandArguments struct {
	args []string
}

func NewCommandArguments(args []string) CommandArguments {
	return CommandArguments{
		args: args,
	}
}

func (ca CommandArguments) RawArgs() []string {
	return ca.args
}

// String returns the arguments as a string, with quotes around arguments that contain spaces
func (ca CommandArguments) String() string {
	if len(ca.args) == 0 {
		return ""
	}

	var n int
	for _, elem := range ca.args {
		n += len(elem) + 2 // add 2 if quotes are needed
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

func (ca CommandArguments) writeString(b *strings.Builder, str string) {
	if strings.Contains(str, " ") {
		b.WriteString(`"`)
		b.WriteString(str)
		b.WriteString(`"`)
	} else {
		b.WriteString(str)
	}
}
