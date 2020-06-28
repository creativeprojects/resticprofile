package config

import (
	"bytes"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
)

func TestEmptyGlobalSection(t *testing.T) {
	configString := `[default]
something = 1
`
	global, err := getGlobalSection(configString)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, constants.DefaultCommand, global.DefaultCommand)
	assert.False(t, global.Initialize)
}

func TestSimpleGlobalSection(t *testing.T) {
	configString := `[global]
default-command = "test"
`
	global, err := getGlobalSection(configString)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", global.DefaultCommand)
	assert.False(t, global.Initialize)
}

func TestFullGlobalSection(t *testing.T) {
	configString := `[global]
ionice = true
ionice-class = 2
ionice-level = 6
nice = 1
priority = "low"
default-command = "version"
initialize = true
restic-binary = "/tmp/restic"
`
	global, err := getGlobalSection(configString)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, global.IONice)
	assert.Equal(t, 2, global.IONiceClass)
	assert.Equal(t, 6, global.IONiceLevel)
	assert.Equal(t, 1, global.Nice)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.True(t, global.Initialize)
	assert.Equal(t, "/tmp/restic", global.ResticBinary)
}

func getGlobalSection(configString string) (*Global, error) {
	c, err := Load(bytes.NewBufferString(configString), "toml")
	if err != nil {
		return nil, err
	}

	global, err := c.GetGlobalSection()
	if err != nil {
		return nil, err
	}
	return global, nil
}
