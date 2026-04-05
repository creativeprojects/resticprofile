package ansi

import (
	"fmt"
	"strings"
)

const (
	EscapeByte = 0x1b
	Escape     = "\x1b"
	ClearLine  = Escape + "[2K"
	Reset      = Escape + "[0m"
)

// Sequence builds an arbitrary escape sequence
func Sequence(terminator byte, attributes ...any) string {
	seq := strings.Builder{}
	seq.WriteString(Escape)
	seq.WriteByte('[')
	for i, attribute := range attributes {
		if i > 0 {
			seq.WriteByte(';')
		}
		_, _ = fmt.Fprint(&seq, attribute)
	}
	seq.WriteByte(terminator)
	return seq.String()
}

// CursorUpLeftN creates the escape sequence to move the cursor left and up N lines
func CursorUpLeftN(lines int) string {
	if lines < 0 {
		lines = 0
	}
	return Sequence('F', lines)
}
