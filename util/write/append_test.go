package write

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendLinesNoAppender(t *testing.T) {
	buffer := new(bytes.Buffer)
	a := NewAppend(buffer, nil)
	_, _ = a.Write([]byte("a\n"))
	_, _ = a.Write([]byte("b\n"))
	_, _ = a.Write([]byte("c\n"))

	assert.Equal(t, "a\nb\nc\n", buffer.String())
}

func TestAppendLinesWithAppender(t *testing.T) {
	buffer := new(bytes.Buffer)
	appender := func(dst []byte, c byte) []byte {
		switch c {
		case '\n':
			return append(dst, '\r', '\n') // normalize to CRLF on Windows
		case '\r':
			return dst
		}
		return append(dst, c)
	}

	a := NewAppend(buffer, appender)
	_, _ = a.Write([]byte("a\n"))
	_, _ = a.Write([]byte("b\n"))
	_, _ = a.Write([]byte("c\n"))

	assert.Equal(t, "a\r\nb\r\nc\r\n", buffer.String())
}
