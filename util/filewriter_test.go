package util

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func TestAsyncFileWriterDoubleClose(t *testing.T) {
// 	dir := t.TempDir()
// 	filename := filepath.Join(dir, "test.log")

// 	w, err := NewAsyncFileWriter(filename)
// 	require.NoError(t, err)

// 	err = w.Close()
// 	assert.NoError(t, err)
// 	err = w.Close()
// 	assert.NoError(t, err)
// }

func TestAsyncFileWriterBasicWrite(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename)
	require.NoError(t, err)

	n, err := w.Write([]byte("hello world"))
	assert.NoError(t, err)
	assert.Equal(t, 11, n)

	err = w.Close()
	assert.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(content))
}

func TestAsyncFileWriterMultipleWrites(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename)
	require.NoError(t, err)

	for _, chunk := range []string{"foo", "bar", "baz"} {
		_, err = w.Write([]byte(chunk))
		require.NoError(t, err)
	}

	err = w.Close()
	assert.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "foobarbaz", string(content))
}

func TestAsyncFileWriterFlush(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename,
		WithAsyncWriteInterval(10*time.Second), // long interval so flush drives the write
	)
	require.NoError(t, err)
	defer w.Close()

	_, err = w.Write([]byte("flushed"))
	require.NoError(t, err)

	fw, ok := w.(*asyncFileWriter)
	require.True(t, ok)
	err = fw.Flush()
	assert.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "flushed", string(content))
}

func TestAsyncFileWriterTruncate(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	// First write
	w, err := NewAsyncFileWriter(filename)
	require.NoError(t, err)
	_, err = w.Write([]byte("original"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	// Second write with truncate
	w, err = NewAsyncFileWriter(filename, WithAsyncFileTruncate())
	require.NoError(t, err)
	_, err = w.Write([]byte("new"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "new", string(content))
}

func TestAsyncFileWriterAppendMode(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	// First write
	w, err := NewAsyncFileWriter(filename)
	require.NoError(t, err)
	_, err = w.Write([]byte("first"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	// Second write appends by default
	w, err = NewAsyncFileWriter(filename)
	require.NoError(t, err)
	_, err = w.Write([]byte("second"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "firstsecond", string(content))
}

func TestAsyncFileWriterFilePerm(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename, WithAsyncFilePerm(0600))
	require.NoError(t, err)
	_, err = w.Write([]byte("data"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestAsyncFileWriterCustomAppendFunc(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	// appender that uppercases every byte
	upperAppender := func(dst []byte, c byte) []byte {
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		return append(dst, c)
	}

	w, err := NewAsyncFileWriter(filename, WithAsyncFileAppendFunc(upperAppender))
	require.NoError(t, err)
	_, err = w.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "HELLO", string(content))
}

func TestAsyncFileWriterKeepOpenFalse(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename, WithAsyncFileKeepOpen(false))
	require.NoError(t, err)
	_, err = w.Write([]byte("keep-closed"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "keep-closed", string(content))
}

func TestAsyncFileWriterInvalidPath(t *testing.T) {
	_, err := NewAsyncFileWriter("/nonexistent/path/test.log")
	assert.Error(t, err)
}

func TestAsyncFileWriterLargeWrite(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	// Write more than asyncWriterBlockSize (4KB)
	large := strings.Repeat("x", asyncWriterBlockSize*3)

	w, err := NewAsyncFileWriter(filename)
	require.NoError(t, err)
	n, err := w.Write([]byte(large))
	assert.NoError(t, err)
	assert.Equal(t, len(large), n)
	require.NoError(t, w.Close())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, large, string(content))
}

func TestAsyncFileWriterWriteInterval(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename, WithAsyncWriteInterval(10*time.Millisecond))
	require.NoError(t, err)
	defer w.Close()

	_, err = w.Write([]byte("interval"))
	require.NoError(t, err)

	// Wait enough time for the ticker to fire
	time.Sleep(50 * time.Millisecond)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "interval", string(content))
}

func TestAsyncFileWriterWriteEmptyData(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename)
	require.NoError(t, err)

	n, err := w.Write([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	require.NoError(t, w.Close())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Empty(t, content)
}

func TestAsyncFileWriterBufferReuse(t *testing.T) {
	// Verify pooled-size writes don't corrupt data due to buffer reuse
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.log")

	w, err := NewAsyncFileWriter(filename)
	require.NoError(t, err)

	var expected bytes.Buffer
	for i := range 10 {
		chunk := []byte(strings.Repeat(string(rune('a'+i)), asyncWriterBlockSize))
		_, err = w.Write(chunk)
		require.NoError(t, err)
		expected.Write(chunk)
	}

	require.NoError(t, w.Close())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, expected.String(), string(content))
}
