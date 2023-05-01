package ansi

import (
	"strings"

	"github.com/fatih/color"
)

var (
	bold      = color.New(color.Bold)
	Bold      = bold.SprintFunc()
	blue      = color.New(color.FgBlue)
	Blue      = bold.SprintFunc()
	cyan      = color.New(color.FgCyan)
	Cyan      = cyan.SprintFunc()
	gray      = New256FgColor(243)
	Gray      = gray.SprintFunc()
	green     = color.New(color.FgGreen)
	Green     = green.SprintFunc()
	yellow    = color.New(color.FgYellow)
	Yellow    = yellow.SprintFunc()
	underline = color.New(color.Underline)
	Underline = underline.Sprint()
)

// New256FgColor return a new xterm 256 (8bit) foreground color
func New256FgColor(code uint8) *color.Color {
	return color.New(38, 5, color.Attribute(code))
}

// New256BgColor return a new xterm 256 (8bit) background color
func New256BgColor(code uint8) *color.Color {
	return color.New(48, 5, color.Attribute(code))
}

func ColorSequence(fn func(a ...interface{}) string) (start, stop string) {
	s := fn("||")
	d := strings.Index(s, "||")
	start, stop = s[:d], s[d+2:]
	return
}
