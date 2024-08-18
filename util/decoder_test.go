package util

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoders(t *testing.T) {
	tests := []struct {
		input      []byte
		expected   string
		expectAsIs bool
	}{
		{input: []byte("hello"), expected: "hello", expectAsIs: true},
		{input: []byte("hËllØ"), expected: "hËllØ", expectAsIs: true},
		{input: []byte("h"), expected: "h", expectAsIs: true},
		{input: []byte(""), expected: "", expectAsIs: true},
		{input: []byte("\ufeffhËllØ"), expected: "hËllØ"},                                      // UTF8-BOM
		{input: []byte{0xef, 0xbb, 0xbf, 'h', 'e', 'l', 'l', 'o'}, expected: "hello"},          // UTF8-BOM
		{input: []byte{0xfe, 0xff, 0, 'h', 0, 'e', 0, 'l', 0, 'l', 0, 'o'}, expected: "hello"}, // UTF16-BE
		{input: []byte{0xff, 0xfe, 'h', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0}, expected: "hello"}, // UTF16-LE
		{input: []byte{'h', 0xcb, 'l', 'l', 0xd8}, expected: "hËllØ"},                          // ISO-8859-1
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			input := bytes.NewReader(test.input)
			reader := NewUTF8Reader(input)
			if test.expectAsIs {
				assert.Same(t, input, reader)
			} else {
				assert.NotSame(t, input, reader)
			}

			data, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, string(data))
		})
	}
}

func TestDecoderOnErrorInput(t *testing.T) {
	custom := fmt.Errorf("custom")
	tests := []struct {
		input    io.ReadSeeker
		expected error
	}{
		{input: &failingSeeker{err: io.EOF}, expected: io.EOF},
		{input: &failingSeeker{err: custom}, expected: custom},
		{input: &failingSeeker{buf: make([]byte, 0), err: custom}, expected: custom},
		{input: &failingSeeker{buf: make([]byte, 1024), err: custom}, expected: custom},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := NewUTF8Reader(test.input).Read(make([]byte, 1))
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestMustRewindToStart(t *testing.T) {
	buf := make([]byte, 1)
	read := func(t *testing.T, reader io.Reader) byte {
		t.Helper()
		n, err := reader.Read(buf)
		assert.Equal(t, n, 1)
		assert.NoError(t, err)
		return buf[0]
	}

	t.Run("rewinds", func(t *testing.T) {
		reader := bytes.NewReader([]byte("AB"))
		assert.Equal(t, byte('A'), read(t, reader))
		assert.Equal(t, byte('B'), read(t, reader))
		mustRewindToStart(reader)
		assert.Equal(t, byte('A'), read(t, reader))
	})

	t.Run("panics-on-offset", func(t *testing.T) {
		assert.PanicsWithError(t, "failed reverting read offset to start: offset was 1", func() {
			mustRewindToStart(&failingSeeker{n: 1})
		})
	})

	t.Run("panics-on-error", func(t *testing.T) {
		assert.PanicsWithError(t, "failed reverting read offset to start: seek-failed", func() {
			mustRewindToStart(&failingSeeker{err: fmt.Errorf("seek-failed")})
		})
	})

	t.Run("no-panic-on-EOF", func(t *testing.T) {
		mustRewindToStart(&failingSeeker{err: io.EOF})
	})
}

type failingSeeker struct {
	buf []byte
	n   int
	err error
}

func (f *failingSeeker) Read(data []byte) (int, error) {
	if f.buf != nil {
		defer func() { f.buf = nil }()
		return copy(data, f.buf), nil
	}
	return f.n, f.err
}
func (f *failingSeeker) Seek(offset int64, whence int) (int64, error) { return int64(f.n), f.err }
