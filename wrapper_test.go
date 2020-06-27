package main

import (
	"os"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func TestGetEmptyEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Empty(t, env)
}

func TestGetSingleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]string{
		"User": "me",
	}
	wrapper := newResticWrapper("restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Equal(t, []string{"USER=me"}, env)
}

func TestGetMultipleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]string{
		"User":     "me",
		"Password": "secret",
	}
	wrapper := newResticWrapper("restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Len(t, env, 2)
	assert.Contains(t, env, "USER=me")
	assert.Contains(t, env, "PASSWORD=secret")
}

func TestEmptyConversionToArgs(t *testing.T) {
	flags := map[string][]string{}
	args := convertIntoArgs(flags)
	assert.Equal(t, []string{}, args)
}

func TestConversionToArgs(t *testing.T) {
	flags := map[string][]string{
		"bool1":   {},
		"bool2":   {""},
		"int1":    {"0"},
		"int2":    {"-100"},
		"string1": {"test"},
		"string2": {"with space"},
		"list":    {"test1", "test2", "test3"},
	}
	args := convertIntoArgs(flags)
	assert.Len(t, args, 16)
	assert.Contains(t, args, "--bool1")
	assert.Contains(t, args, "--bool2")
	assert.Contains(t, args, "0")
	assert.Contains(t, args, "-100")
	assert.Contains(t, args, "test")
	assert.Contains(t, args, "\"with space\"")
	assert.Contains(t, args, "test1")
	assert.Contains(t, args, "test2")
	assert.Contains(t, args, "test3")
}

func TestPreProfileScriptFail(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.RunBefore = []string{"exit 1"} // this should both work on unix shell and windows batch
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "exit status 1")
}

func TestPostProfileScriptFail(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"exit 1"} // this should both work on unix shell and windows batch
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "exit status 1")
}

func TestRunEchoProfile(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}

func TestPostProfileAfterFail(t *testing.T) {
	testFile := "TestPostProfileAfterFail.txt"
	_ = os.Remove(testFile)
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"echo failed > " + testFile}
	wrapper := newResticWrapper("exit", false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "exit status 1")
	assert.NoFileExistsf(t, testFile, "the run-after script should not have been running")
	_ = os.Remove(testFile)
}

func TestPostFailProfile(t *testing.T) {
	testFile := "TestPostFailProfile.txt"
	_ = os.Remove(testFile)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo failed > " + testFile}
	wrapper := newResticWrapper("exit", false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "exit status 1")
	assert.FileExistsf(t, testFile, "the run-after-fail script has not been running")
	_ = os.Remove(testFile)
}
