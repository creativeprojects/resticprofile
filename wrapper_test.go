package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
)

func TestGetEmptyEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("restic", false, false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Empty(t, env)
}

func TestGetSingleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]string{
		"User": "me",
	}
	wrapper := newResticWrapper("restic", false, false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Equal(t, []string{"USER=me"}, env)
}

func TestGetMultipleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]string{
		"User":     "me",
		"Password": "secret",
	}
	wrapper := newResticWrapper("restic", false, false, profile, "test", nil, nil)
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
	wrapper := newResticWrapper("echo", false, false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-before on profile 'name': exit status 1")
}

func TestPostProfileScriptFail(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"exit 1"} // this should both work on unix shell and windows batch
	wrapper := newResticWrapper("echo", false, false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-after on profile 'name': exit status 1")
}

func TestRunEchoProfile(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", false, false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}

func TestPostProfileAfterFail(t *testing.T) {
	testFile := "TestPostProfileAfterFail.txt"
	_ = os.Remove(testFile)
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"echo failed > " + testFile}
	wrapper := newResticWrapper("exit", false, false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "1 on profile 'name': exit status 1")
	assert.NoFileExistsf(t, testFile, "the run-after script should not have been running")
	_ = os.Remove(testFile)
}

func TestPostFailProfile(t *testing.T) {
	testFile := "TestPostFailProfile.txt"
	_ = os.Remove(testFile)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo failed > " + testFile}
	wrapper := newResticWrapper("exit", false, false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "1 on profile 'name': exit status 1")
	assert.FileExistsf(t, testFile, "the run-after-fail script has not been running")
	_ = os.Remove(testFile)
}

func Example_runProfile() {
	term.SetOutput(os.Stdout)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", false, false, profile, "test", nil, nil)
	wrapper.runProfile()
	// Output: test
}

func TestRunRedirectOutputOfEchoProfile(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", false, false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "test", strings.TrimSpace(buffer.String()))
}

func TestDryRun(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", false, true, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "", buffer.String())
}

func TestEnvProfileName(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "TestEnvProfileName")
	if runtime.GOOS == "windows" {
		profile.RunBefore = []string{"echo profile name = %PROFILE_NAME%"}
	} else {
		profile.RunBefore = []string{"echo profile name = $PROFILE_NAME"}
	}
	wrapper := newResticWrapper("echo", false, false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "profile name = TestEnvProfileName\ntest\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvProfileCommand(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	if runtime.GOOS == "windows" {
		profile.RunBefore = []string{"echo profile command = %PROFILE_COMMAND%"}
	} else {
		profile.RunBefore = []string{"echo profile command = $PROFILE_COMMAND"}
	}
	wrapper := newResticWrapper("echo", false, false, profile, "test-command", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "profile command = test-command\ntest-command\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvError(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	if runtime.GOOS == "windows" {
		profile.RunAfterFail = []string{"echo error: %ERROR%"}
	} else {
		profile.RunAfterFail = []string{"echo error: $ERROR"}
	}
	wrapper := newResticWrapper("exit", false, false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "error: 1 on profile 'name': exit status 1\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestRunProfileWithSetPIDCallback(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Lock = filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestRunProfileWithSetPIDCallback", time.Now().UnixNano(), os.Getpid()))
	t.Logf("lockfile = %s", profile.Lock)
	wrapper := newResticWrapper("echo", false, false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}
