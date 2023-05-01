package ansi

import (
	"io"
	"unicode/utf8"
)

type lineLengthWriter struct {
	writer                                            io.Writer
	tokens                                            []byte
	maxLineLength, lastWhite, breakLength, lineLength int
	invisibleLength, lastWhiteInvisibleLength         int
	inAnsi                                            bool
}

// NewLineLengthWriter return an io.Writer that limits the max line length, adding line breaks ('\n') as needed.
// The writer detects the right most column (consecutive whitespace) and aligns content if possible.
// UTF sequences are counted as single character and ANSI escape sequences are not counted at all.
func NewLineLengthWriter(writer io.Writer, maxLineLength int) io.Writer {
	return &lineLengthWriter{
		tokens:        []byte{' ', '\n'},
		writer:        writer,
		maxLineLength: maxLineLength,
	}
}

func (l *lineLengthWriter) visibleLineLength() int { return l.lineLength - l.invisibleLength }

func (l *lineLengthWriter) Write(p []byte) (n int, err error) {
	written := 0
	offset := l.lineLength

	for i := 0; i < len(p); i++ {
		l.lineLength++
		ws := p[i] == l.tokens[0] // whitespace
		br := p[i] == l.tokens[1] // linebreak

		// don't count ansi control sequences
		if l.inAnsi = l.inAnsi || p[i] == EscapeByte; l.inAnsi {
			terminator := (p[i] >= 'a' && p[i] <= 'z') || (p[i] >= 'A' && p[i] <= 'Z')
			l.inAnsi = !terminator
			l.invisibleLength++
			continue
		}

		// count UTF sequence as one character
		if p[i] >= utf8.RuneSelf {
			if !utf8.RuneStart(p[i]) {
				l.invisibleLength++
			}
			continue
		}

		if !br && l.visibleLineLength() > l.maxLineLength && l.lastWhite-offset > 0 {
			lastWhiteIndex := l.lastWhite - offset - 1
			remainder := i - lastWhiteIndex

			if written, err = l.writer.Write(p[:lastWhiteIndex]); err == nil {
				p = p[lastWhiteIndex+1:]
				i = remainder - 1
				n += written + 1

				_, _ = l.writer.Write(l.tokens[1:]) // write break (instead of WS at lastWhiteIndex)
				for j := 0; j < l.breakLength; j++ {
					_, _ = l.writer.Write(l.tokens[0:1]) // fill spaces for alignment
				}

				l.lineLength = l.breakLength + remainder
				l.lastWhite = l.breakLength
				offset = l.breakLength

				l.invisibleLength -= l.lastWhiteInvisibleLength
				l.lastWhiteInvisibleLength = 0
			} else {
				return
			}
		}

		if ws {
			if l.lastWhite == l.lineLength-1 && l.visibleLineLength() < l.maxLineLength*2/3 {
				l.breakLength = l.visibleLineLength()
			}
			l.lastWhite = l.lineLength
			l.lastWhiteInvisibleLength = l.invisibleLength

		} else if br {
			if written, err = l.writer.Write(p[:i+1]); err == nil {
				p = p[i+1:]
				i = -1
				n += written

				l.lineLength = 0
				l.lastWhite = 0
				l.breakLength = 0
				offset = 0

				l.invisibleLength = 0
				l.lastWhiteInvisibleLength = 0
			} else {
				return
			}
		}
	}

	// write remainder
	if written, err = l.writer.Write(p); err == nil {
		n += written
	}
	return
}
