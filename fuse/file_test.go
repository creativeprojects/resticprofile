package fuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFile(t *testing.T) {
	file := NewFile("name.txt", nil, []byte("data"))
	assert.Equal(t, "name.txt", file.Name())
	assert.Nil(t, file.FileInfo())
}
