//go:build windows

package fuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMountFS(t *testing.T) {
	_, err := MountFS("mnt", []File{})
	assert.Error(t, err)
}
