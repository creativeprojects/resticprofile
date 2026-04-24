package ssh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigForInternalSSH(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Username: "user",
	}
	err := config.ValidateInternal()
	require.NoError(t, err)
	assert.NotEmpty(t, config.KnownHostsPath)

	// check there's no public key path
	for _, path := range config.PrivateKeyPaths {
		t.Logf("private key: %s", path)
		assert.False(t, strings.HasSuffix(path, ".pub"), "public key path should not be included: %s", path)
	}
}

func TestValidateConfigForOpenSSH(t *testing.T) {
	config := Config{
		Host: "localhost",
	}
	err := config.ValidateOpenSSH()
	require.NoError(t, err)
}
