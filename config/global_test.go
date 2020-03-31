package config

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
)

func TestEmptyGlobalSection(t *testing.T) {
	configString := `[default]
something = 1
`
	global, err := getGlobalSection(configString)
	if err != nil {
		t.Fatal(err)
	}

	expectString(t, "DefaultCommand", global.DefaultCommand, global.DefaultCommand)
	expectBool(t, "Initialize", global.Initialize, false)
}

func TestSimpleGlobalSection(t *testing.T) {
	configString := `[global]
default-command = "test"
`
	global, err := getGlobalSection(configString)
	if err != nil {
		t.Fatal(err)
	}

	expectString(t, "DefaultCommand", global.DefaultCommand, "test")
	expectBool(t, "Initialize", global.Initialize, false)
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

	expectBool(t, "IONice", global.IONice, true)
	expectInt(t, "IONiceClass", global.IONiceClass, 2)
	expectInt(t, "IONiceLevel", global.IONiceLevel, 6)
	expectInt(t, "Nice", global.Nice, 1)
	expectString(t, "Priority", global.Priority, "low")
	expectString(t, "DefaultCommand", global.DefaultCommand, "version")
	expectBool(t, "Initialize", global.Initialize, true)
	expectString(t, "ResticBinary", global.ResticBinary, "/tmp/restic")
}

func getGlobalSection(configString string) (*Global, error) {
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(configString))
	if err != nil {
		return nil, err
	}

	global, err := GetGlobalSection()
	if err != nil {
		return nil, err
	}
	return global, nil
}

func expectBool(t *testing.T, name string, value, expected bool) {
	if value != expected {
		t.Errorf("Expected %s to be %t but found %t", name, expected, value)
	}
}

func expectInt(t *testing.T, name string, value, expected int) {
	if value != expected {
		t.Errorf("Expected %s to be %d but found %d", name, expected, value)
	}
}

func expectString(t *testing.T, name, value, expected string) {
	if value != expected {
		t.Errorf("Expected %s to be '%s' but found '%s'", name, expected, value)
	}
}
