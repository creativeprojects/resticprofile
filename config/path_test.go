package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandEnv(t *testing.T) {
	path := os.Getenv("PATH")
	assert.Equal(t, path, expandEnv("$PATH"))
	assert.Equal(t, path, expandEnv("${PATH}"))
	assert.Equal(t, "%PATH%", expandEnv("%PATH%"))
	assert.Equal(t, "$PATH", expandEnv("$$PATH"))
	assert.Equal(t, "", expandEnv("${__UNDEFINED_ENV_VAR__}"))
	assert.Equal(t, "", expandEnv("$__UNDEFINED_ENV_VAR__"))
}
