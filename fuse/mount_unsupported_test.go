//go:build windows || netbsd || openbsd || solaris || (linux && 386) || (linux && arm) || (linux && mips) || (linux && mipsle)

package fuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMountFS(t *testing.T) {
	_, err := MountFS("mnt", []File{})
	assert.Error(t, err)
}
