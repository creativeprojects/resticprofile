package util

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutableIsAbsolute(t *testing.T) {
	executable, err := Executable()
	require.NoError(t, err)
	assert.NotEmpty(t, executable)

	assert.True(t, filepath.IsAbs(executable))
}
