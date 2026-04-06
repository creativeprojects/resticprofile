package write

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileDefaultOption(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "testfile")

	w, err := NewFile(filename)
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

func TestFileCloseAfterWrite(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "testfile")

	w, err := NewFile(filename, WithFileKeepOpen(false))
	require.NoError(t, err)

	n, err := w.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))

	n, err = w.Write([]byte(" world"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	content, err = os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(content))

	err = w.Close()
	assert.NoError(t, err)
}
