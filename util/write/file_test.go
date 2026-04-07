package write

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileDefaultOption(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "testfile")

	w, err := NewFile(filename, WithFileKeepOpen(true), WithFileKeepOpenTimeout(1*time.Millisecond))
	require.NoError(t, err)

	n, err := w.Write([]byte("hello world"))
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	time.Sleep(2 * time.Millisecond)

	err = w.Close()
	assert.NoError(t, err)

	opened, closed := w.stats()
	assert.Equal(t, int32(1), opened)
	assert.Equal(t, int32(1), closed)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(content))
}

func TestFileCloseAfterWrite(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "testfile")

	w, err := NewFile(filename, WithFileKeepOpen(false), WithFileKeepOpenTimeout(1*time.Millisecond))
	require.NoError(t, err)

	n, err := w.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	time.Sleep(2 * time.Millisecond)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))

	n, err = w.Write([]byte(" world"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	time.Sleep(2 * time.Millisecond)

	content, err = os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(content))

	err = w.Close()
	assert.NoError(t, err)

	opened, closed := w.stats()
	assert.Equal(t, int32(3), opened) // 1 during instantiation + 1 for each write
	assert.Equal(t, int32(3), closed) // 1 during instantiation + 1 for each write
}

func TestFileNoTimeToCloseAfterWrite(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "testfile")

	w, err := NewFile(filename, WithFileKeepOpen(false), WithFileKeepOpenTimeout(1*time.Second))
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

	opened, closed := w.stats()
	assert.Equal(t, int32(2), opened) // 1 during instantiation + 1 for both write
	assert.Equal(t, int32(2), closed) // 1 during instantiation + 1 for both write
}

func TestFileCanFlush(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "testfile")

	w, err := NewFile(filename)
	require.NoError(t, err)
	assert.NoError(t, w.Flush())

	n, err := w.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.NoError(t, w.Flush())

	n, err = w.Write([]byte(" world"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.NoError(t, w.Flush())

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(content))

	assert.NoError(t, w.Flush())
	err = w.Close()
	assert.NoError(t, err)

	assert.NoError(t, w.Flush())
}
