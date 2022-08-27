package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/monitor/status"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mockBinary string
)

func init() {
	// build restic mock
	cmd := exec.Command("go", "build", "./shell/mock")
	cmd.Run()
	if runtime.GOOS == "windows" {
		mockBinary = "mock.exe"
	} else {
		mockBinary = "./mock"
	}

	// Add params that need to be passed to the mock binary
	commonResticArgsList = append(commonResticArgsList, "--exit")
}

func TestCommonResticArgs(t *testing.T) {
	wrapper := &resticWrapper{}
	for _, arg := range commonResticArgsList {
		var args []string
		list := []string{"-x", "-x=1", "-x 2 x=y", "--xxx", "--xxx=v", "--xxx k=v", arg, arg + "=1", arg + " ka=va"}

		for i := 0; i < 20; i++ {
			rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
			args = args[:0]
			for _, item := range list {
				args = append(args, strings.Split(item, " ")...)
			}

			args = wrapper.commonResticArgs(args)

			assert.Len(t, args, 4)
			assert.Subset(t, args, []string{arg, arg + "=1", "ka=va"})
			for _, item := range []string{"-x", "--xxx", "x=y", "k=v", "2"} {
				assert.NotContains(t, args, item)
			}
		}
	}
}

func TestGetEmptyEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, "restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Empty(t, env)
}

func TestGetSingleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]config.ConfidentialValue{
		"User": config.NewConfidentialValue("me"),
	}
	wrapper := newResticWrapper(nil, "restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Equal(t, []string{"USER=me"}, env)
}

func TestGetMultipleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]config.ConfidentialValue{
		"User":     config.NewConfidentialValue("me"),
		"Password": config.NewConfidentialValue("secret"),
	}
	wrapper := newResticWrapper(nil, "restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Len(t, env, 2)
	assert.Contains(t, env, "USER=me")
	assert.Contains(t, env, "PASSWORD=secret")
}

func TestPreProfileScriptFail(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.RunBefore = []string{"exit 1"} // this should both work on unix shell and windows batch
	wrapper := newResticWrapper(nil, "echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-before on profile 'name': exit status 1")
}

func TestPostProfileScriptFail(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"exit 1"} // this should both work on unix shell and windows batch
	wrapper := newResticWrapper(nil, "echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-after on profile 'name': exit status 1")
}

func TestRunEchoProfile(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, "echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}

func TestPostProfileAfterFail(t *testing.T) {
	testFile := "TestPostProfileAfterFail.txt"
	_ = os.Remove(testFile)
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"echo failed > " + testFile}
	wrapper := newResticWrapper(nil, "exit", false, profile, "1", nil, nil)
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
	wrapper := newResticWrapper(nil, "exit", false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "1 on profile 'name': exit status 1")
	assert.FileExistsf(t, testFile, "the run-after-fail script has not been running")
	_ = os.Remove(testFile)
}

func TestFinallyProfile(t *testing.T) {
	testFile := "TestFinallyProfile.txt"
	defer os.Remove(testFile)

	var profile *config.Profile
	newProfile := func() {
		_ = os.Remove(testFile)
		profile = config.NewProfile(nil, "name")
		profile.RunFinally = []string{"echo finally > " + testFile}
		profile.Backup = &config.BackupSection{}
		profile.Backup.RunFinally = []string{"echo finally-backup > " + testFile}
	}

	assertFileEquals := func(t *testing.T, expected string) {
		content, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(string(content)), expected)
	}

	t.Run("backup-before-profile", func(t *testing.T) {
		newProfile()
		wrapper := newResticWrapper(nil, "echo", false, profile, "backup", nil, nil)
		err := wrapper.runProfile()
		assert.NoError(t, err)
		assertFileEquals(t, "finally")
	})

	t.Run("on-backup-only", func(t *testing.T) {
		newProfile()
		profile.RunFinally = nil
		wrapper := newResticWrapper(nil, "echo", false, profile, "backup", nil, nil)
		err := wrapper.runProfile()
		assert.NoError(t, err)
		assertFileEquals(t, "finally-backup")
	})

	t.Run("on-error", func(t *testing.T) {
		newProfile()
		wrapper := newResticWrapper(nil, "exit", false, profile, "1", nil, nil)
		err := wrapper.runProfile()
		assert.EqualError(t, err, "1 on profile 'name': exit status 1")
		assertFileEquals(t, "finally")
	})
}

func Example_runProfile() {
	term.SetOutput(os.Stdout)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, "echo", false, profile, "test", nil, nil)
	wrapper.runProfile()
	// Output: test
}

func TestRunRedirectOutputOfEchoProfile(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, "echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "test", strings.TrimSpace(buffer.String()))
}

func TestDryRun(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, "echo", true, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "", buffer.String())
}

func TestEnvProfileName(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "TestEnvProfileName")
	profile.RunBefore = []string{"echo profile name = $PROFILE_NAME"}

	wrapper := newResticWrapper(nil, "echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "profile name = TestEnvProfileName\ntest\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvProfileCommand(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunBefore = []string{"echo profile command = $PROFILE_COMMAND"}

	wrapper := newResticWrapper(nil, "echo", false, profile, "test-command", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "profile command = test-command\ntest-command\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvError(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo error: $ERROR_MESSAGE"}

	wrapper := newResticWrapper(nil, "exit", false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "error: 1 on profile 'name': exit status 1\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvErrorCommandLine(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo cmd: $ERROR_COMMANDLINE"}

	wrapper := newResticWrapper(nil, "exit", false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "cmd: \"exit\" \"1\"\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvErrorExitCode(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo exit-code: $ERROR_EXIT_CODE"}

	wrapper := newResticWrapper(nil, "exit", false, profile, "5", nil, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "exit-code: 5\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvStderr(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo stderr: $ERROR_STDERR"}

	wrapper := newResticWrapper(nil, mockBinary, false, profile, "command", []string{"--stderr", "error_message", "--exit", "1"}, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "stderr: error_message", strings.TrimSpace(strings.ReplaceAll(buffer.String(), "\r\n", "\n")))
}

func TestRunProfileWithSetPIDCallback(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Lock = filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestRunProfileWithSetPIDCallback", time.Now().UnixNano(), os.Getpid()))
	t.Logf("lockfile = %s", profile.Lock)
	wrapper := newResticWrapper(nil, "echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}

func TestInitializeNoError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", nil, nil)
	err := wrapper.runInitialize()
	require.NoError(t, err)
}

func TestInitializeWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "10"}, nil)
	err := wrapper.runInitialize()
	require.Error(t, err)
}

func TestInitializeCopyNoError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", nil, nil)
	err := wrapper.runInitializeCopy()
	require.NoError(t, err)
}

func TestInitializeCopyWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "10"}, nil)
	err := wrapper.runInitializeCopy()
	require.Error(t, err)
}

func TestCheckNoError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", nil, nil)
	err := wrapper.runCheck()
	require.NoError(t, err)
}

func TestCheckWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "10"}, nil)
	err := wrapper.runCheck()
	require.Error(t, err)
}

func TestRetentionNoError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", nil, nil)
	err := wrapper.runRetention()
	require.NoError(t, err)
}

func TestRetentionWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "10"}, nil)
	err := wrapper.runRetention()
	require.Error(t, err)
}

func TestBackupWithStreamSource(t *testing.T) {
	expected := "---Backup-Content---"
	expectedInterruptedError := []string{
		"stdin-test on profile 'name': exit status 128",
		"stdin-test on profile 'name': signal: interrupt",
	}

	fillBufferCommand := func() (cmds []string) {
		cmd := "echo " + strings.Repeat("-", 240)
		for i := 0; i < 35; i++ { // 35 * 240 = 8400 (buffer is 8192)
			cmds = append(cmds, cmd)
		}
		return
	}

	profileAndWrapper := func(*testing.T) (profile *config.Profile, wrapper *resticWrapper) {
		profile = config.NewProfile(nil, "name")
		profile.Backup = &config.BackupSection{}
		signals := make(chan os.Signal, 1)
		wrapper = newResticWrapper(nil, mockBinary, false, profile, "stdin-test", nil, signals)
		return
	}

	run := func(t *testing.T, wrapper *resticWrapper) (string, error) {
		file := path.Join(os.TempDir(), fmt.Sprintf("TestBackupWithStreamSource.%d.txt", rand.Int()))
		defer os.Remove(file)

		args := wrapper.moreArgs
		wrapper.moreArgs = append([]string{"--stdout-file=" + file}, args...)
		err := wrapper.runCommand("backup")
		wrapper.moreArgs = args
		if err != nil {
			return "", err
		}

		content, err := os.ReadFile(file)
		if wrapper.dryRun {
			require.Error(t, err, "mock was called")
			content = []byte{}
		} else {
			require.NoError(t, err, "mock was not called")
		}
		return string(content), nil
	}

	t.Run("ReadStdin", func(t *testing.T) {
		profile, wrapper := profileAndWrapper(t)
		profile.Backup.UseStdin = true
		wrapper.stdin = io.NopCloser(strings.NewReader(expected))

		content, err := run(t, wrapper)

		assert.NoError(t, err)
		assert.Equal(t, expected, content)
		assert.Nil(t, wrapper.stdin, "stdin must be set to nil when consumed")
	})

	t.Run("ReadStdinNotTwice", func(t *testing.T) {
		profile, wrapper := profileAndWrapper(t)
		profile.Backup.UseStdin = true
		require.NotNil(t, wrapper.stdin)
		wrapper.stdin = nil

		_, err := run(t, wrapper)

		assert.EqualError(t, err, "stdin-test on profile 'name': stdin was already consumed. cannot read it twice")
	})

	t.Run("ReadStreamSource", func(t *testing.T) {
		profile, wrapper := profileAndWrapper(t)
		profile.Backup.StdinCommand = []string{
			fmt.Sprintf("echo %s", expected),
			fmt.Sprintf("echo %s", expected),
			fmt.Sprintf("echo %s", expected),
		}
		profile.ResolveConfiguration()

		expectedResult := strings.Repeat(fmt.Sprintln(expected), 3)

		// can be retried, test multiple invocations
		for i := 0; i < 3; i++ {
			content, err := run(t, wrapper)
			assert.NoError(t, err)
			assert.Equal(t, expectedResult, strings.ReplaceAll(content, "\r\n", "\n"))
		}
	})

	t.Run("StreamSourceReportsInitialError", func(t *testing.T) {
		profile, wrapper := profileAndWrapper(t)
		profile.Backup.StdinCommand = []string{"exit 2"}
		profile.ResolveConfiguration()

		_, err := run(t, wrapper)
		require.NotNil(t, err)
		assert.EqualError(t, err, "stdin-test on profile 'name': 'stdin-command' on profile 'name': exit status 2")
	})

	t.Run("StreamSourceWorksWithDryRun", func(t *testing.T) {
		profile, wrapper := profileAndWrapper(t)
		wrapper.dryRun = true
		profile.Backup.StdinCommand = []string{"exit 2"}
		profile.ResolveConfiguration()

		_, err := run(t, wrapper)
		assert.Nil(t, err)
	})

	t.Run("StreamSourceErrorSendsSIGINT", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("signal handling is not supported on Windows")
		}
		profile, wrapper := profileAndWrapper(t)
		wrapper.moreArgs = []string{"--sleep", "12000"}

		profile.Backup.StdinCommand = append(fillBufferCommand(), "exit 2")
		profile.ResolveConfiguration()

		start := time.Now()
		_, err := run(t, wrapper)
		assert.Less(t, time.Now().Sub(start), time.Second*10, "timeout, interrupt not sent to restic")

		require.NotNil(t, err)
		assert.Contains(t, expectedInterruptedError, err.Error())
	})

	t.Run("CanTerminateStreamSource", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("signal handling is not supported on Windows")
		}
		profile, wrapper := profileAndWrapper(t)
		profile.Backup.StdinCommand = append(fillBufferCommand(), mockBinary+" cmd --sleep 12000")
		profile.ResolveConfiguration()

		go func() {
			time.Sleep(500 * time.Millisecond)
			wrapper.sigChan <- os.Interrupt
		}()
		start := time.Now()
		_, err := run(t, wrapper)
		assert.Less(t, time.Now().Sub(start), time.Second*10, "timeout, interrupt not sent to stdin-command")

		require.NotNil(t, err)
		assert.EqualError(t, err, "stdin-test on profile 'name': io: read/write on closed pipe")
	})
}

func TestBackupWithSuccess(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", nil, nil)
	err := wrapper.runCommand("backup")
	require.NoError(t, err)
}

func TestBackupWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "1"}, nil)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithNoConfiguration(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "1"}, nil)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithNoConfigurationButStatusFile(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.StatusFile = "status.json"
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "1"}, nil)
	wrapper.addProgress(status.NewProgress(profile, status.NewStatus("status.json")))
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithWarningAsError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "3"}, nil)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithSupressedWarnings(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{NoErrorOnWarning: true}
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "", []string{"--exit", "3"}, nil)
	err := wrapper.runCommand("backup")
	require.NoError(t, err)
}

func TestRunShellCommands(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{}
	profile.Check = &config.SectionWithScheduleAndMonitoring{}
	profile.Copy = &config.CopySection{}
	profile.Forget = &config.SectionWithScheduleAndMonitoring{}
	profile.Init = &config.InitSection{}
	profile.Prune = &config.SectionWithScheduleAndMonitoring{}

	sections := map[string]*config.RunShellCommandsSection{
		"backup": profile.Backup.GetRunShellCommands(),
		"check":  profile.Check.GetRunShellCommands(),
		"copy":   profile.Copy.GetRunShellCommands(),
		"forget": profile.Forget.GetRunShellCommands(),
		"init":   profile.Init.GetRunShellCommands(),
		"prune":  profile.Prune.GetRunShellCommands(),
	}

	for command, section := range sections {
		t.Run(fmt.Sprintf("run-before '%s'", command), func(t *testing.T) {
			section.RunBefore = []string{"exit 2"}
			wrapper := newResticWrapper(nil, mockBinary, false, profile, command, nil, nil)
			err := wrapper.runProfile()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "exit status 2")

			section.RunBefore = []string{""}
			wrapper = newResticWrapper(nil, mockBinary, false, profile, command, nil, nil)
			err = wrapper.runProfile()
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("run-after '%s'", command), func(t *testing.T) {
			section.RunAfter = []string{"exit 2"}
			wrapper := newResticWrapper(nil, mockBinary, false, profile, command, nil, nil)
			err := wrapper.runProfile()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "exit status 2")

			section.RunAfter = []string{""}
			wrapper = newResticWrapper(nil, mockBinary, false, profile, command, nil, nil)
			err = wrapper.runProfile()
			require.NoError(t, err)
		})
	}
}

func TestRunStreamErrorHandler(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)

	errorCommand := `echo "detected error in $PROFILE_COMMAND"`

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{}
	profile.StreamError = []config.StreamErrorSection{{Pattern: ".+error-line.+", Run: errorCommand}}
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "backup", []string{"--stderr", "--error-line--"}, nil)

	err := wrapper.runProfile()
	require.NoError(t, err)
	assert.Contains(t, buffer.String(), "detected error in backup")
}

func TestRunStreamErrorHandlerDoesNotBreakCommand(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{}
	profile.StreamError = []config.StreamErrorSection{{Pattern: ".+error-line.+", Run: "exit 1"}}
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "backup", []string{"--stderr", "--error-line--"}, nil)

	err := wrapper.runProfile()
	require.NoError(t, err)
}

func TestStreamErrorHandlerWithInvalidRegex(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{}
	profile.StreamError = []config.StreamErrorSection{{Pattern: "(", Run: "echo pass"}}
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "backup", []string{}, nil)

	err := wrapper.runProfile()
	assert.EqualError(t, err, "backup on profile 'name': stream error callback: echo pass failed to register (: error parsing regexp: missing closing ): `(`")
}

type mockOutputAnalysis struct {
	monitor.OutputAnalysis
	lockWho      string
	lockDuration time.Duration
}

func (m *mockOutputAnalysis) ContainsRemoteLockFailure() bool {
	return m.lockWho != ""
}

func (m *mockOutputAnalysis) GetRemoteLockedSince() (time.Duration, bool) {
	return m.lockDuration, m.lockDuration > 0
}

func (m *mockOutputAnalysis) GetRemoteLockedBy() (string, bool) {
	return m.lockWho, len(m.lockWho) > 1
}

func TestCanRetryAfterRemoteStaleLockFailure(t *testing.T) {
	mockOutput := &mockOutputAnalysis{lockWho: "TestCanRetryAfterRemoteStaleLockFailure"}

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Repository = config.NewConfidentialValue("my-repo")
	profile.ForceLock = true
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "backup", nil, nil)
	wrapper.startTime = time.Now()
	wrapper.global.ResticStaleLockAge = 0 // disable stale lock handling

	// No retry when no stale remote-lock failure
	assert.True(t, mockOutput.ContainsRemoteLockFailure())
	retry, sleep := wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// Ignores stale lock when disabled
	mockOutput.lockDuration = constants.MinResticStaleLockAge
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// Ignores non-stale lock
	mockOutput.lockDuration = constants.MinResticStaleLockAge - time.Nanosecond
	wrapper.global.ResticStaleLockAge = time.Millisecond
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// Unlocks stale lock
	mockOutput.lockDuration = constants.MinResticStaleLockAge
	assert.False(t, wrapper.doneTryUnlock)
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)
	assert.Equal(t, time.Duration(0), sleep)
	assert.True(t, wrapper.doneTryUnlock)

	// Unlock is run only once
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// Unlock is not run when ForceLock is disabled
	wrapper.doneTryUnlock = false
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)

	profile.ForceLock = false
	wrapper.doneTryUnlock = false
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
}

func TestCanRetryAfterRemoteLockFailure(t *testing.T) {
	mockOutput := &mockOutputAnalysis{}

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Repository = config.NewConfidentialValue("my-repo")
	wrapper := newResticWrapper(nil, mockBinary, false, profile, "backup", nil, nil)
	wrapper.startTime = time.Now()
	wrapper.global.ResticLockRetryAfter = 0 // disable remote lock retry

	// No retry when no remote-lock failure
	assert.False(t, mockOutput.ContainsRemoteLockFailure())
	retry, sleep := wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// No retry when lockWait is nil
	mockOutput.lockWho = "TestCanRetryAfterRemoteLockFailure"
	assert.True(t, mockOutput.ContainsRemoteLockFailure())
	retry, _ = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// No retry when disabled
	wrapper.maxWaitOnLock(constants.MinResticLockRetryTime + 50*time.Millisecond)
	retry, _ = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// No retry when no time left
	wrapper.maxWaitOnLock(constants.MinResticLockRetryTime - time.Nanosecond)
	wrapper.global.ResticLockRetryAfter = constants.MinResticLockRetryTime // enable remote lock retry
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// Retry is acceptable when there is enough remaining time for the delay (ResticLockRetryAfter)
	wrapper.maxWaitOnLock(constants.MinResticLockRetryTime + 50*time.Millisecond)
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)
	assert.Equal(t, constants.MinResticLockRetryTime, sleep)

	wrapper.maxWaitOnLock(constants.MaxResticLockRetryTime + 50*time.Millisecond)
	wrapper.global.ResticLockRetryAfter = 2 * constants.MaxResticLockRetryTime
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)
	assert.Equal(t, constants.MaxResticLockRetryTime, sleep)
}

func TestLocksAndLockWait(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Lock = filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestLockWait", time.Now().UnixNano(), os.Getpid()))
	defer os.Remove(profile.Lock)

	term.SetOutput(os.Stdout)

	w1 := newResticWrapper(nil, mockBinary, false, profile, "backup", []string{"--sleep", "1500"}, nil)
	w2 := newResticWrapper(nil, mockBinary, false, profile, "backup", nil, nil)
	w3 := newResticWrapper(nil, mockBinary, false, profile, "backup", nil, nil)

	assertIsLockError := func(err error) bool {
		return err != nil && strings.HasPrefix(err.Error(), "another process is already running this profile")
	}

	// Setup 2 processes (w1, w2), one that locks and one that fails on the lock
	{
		w1Chan := make(chan bool, 1)
		defer func() { <-w1Chan }()

		go func() {
			for retries := 2; retries >= 0; retries-- {
				if err := w1.runProfile(); err == nil {
					break
				} else if retries == 0 || !assertIsLockError(err) {
					assert.NoError(t, err, "TestLockWait-w1")
				}
			}
			w1Chan <- true
		}()

		for i := 10; i >= 0; i++ {
			time.Sleep(20 * time.Millisecond)
			if err := w2.runProfile(); assertIsLockError(err) {
				break
			}
			if i == 0 {
				assert.Fail(t, "Did not wait on lock file")
			}
		}
	}

	// W2: Run produces lock failure
	assertIsLockError(w2.runProfile())

	// W3: Ignore lock can run despite the lock
	w3.ignoreLock()
	assert.NoError(t, w3.runProfile())

	// W2: Too little lock wait produces lock failure
	w2.maxWaitOnLock(100 * time.Millisecond)
	assertIsLockError(w2.runProfile())

	// W2: Succeeds to wait when lockWait is large enough
	w2.maxWaitOnLock(2 * time.Second)
	assert.NoError(t, w2.runProfile())
}

func TestGetContext(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "TestProfile")
	wrapper := newResticWrapper(nil, "", false, profile, "TestCommand", nil, nil)
	require.NotNil(t, wrapper)
	ctx := wrapper.getContext()
	assert.Equal(t, "TestProfile", ctx.ProfileName)
	assert.Equal(t, "TestCommand", ctx.ProfileCommand)
	assert.Equal(t, "", ctx.Error.Message)
	assert.Equal(t, "", ctx.Error.ExitCode)
	assert.Equal(t, "", ctx.Error.CommandLine)
	assert.Equal(t, "", ctx.Error.Stderr)
}

func TestGetContextWithError(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "TestProfile")
	wrapper := newResticWrapper(nil, "", false, profile, "TestCommand", nil, nil)
	require.NotNil(t, wrapper)
	ctx := wrapper.getContextWithError(nil)
	assert.Equal(t, "TestProfile", ctx.ProfileName)
	assert.Equal(t, "TestCommand", ctx.ProfileCommand)
	assert.Equal(t, "", ctx.Error.Message)
	assert.Equal(t, "", ctx.Error.ExitCode)
	assert.Equal(t, "", ctx.Error.CommandLine)
	assert.Equal(t, "", ctx.Error.Stderr)
}

func TestGetErrorContext(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "")
	wrapper := newResticWrapper(nil, "", false, profile, "", nil, nil)
	require.NotNil(t, wrapper)
	ctx := wrapper.getErrorContext(nil)
	assert.Equal(t, "", ctx.Message)
	assert.Equal(t, "", ctx.ExitCode)
	assert.Equal(t, "", ctx.CommandLine)
	assert.Equal(t, "", ctx.Stderr)
}

func TestGetErrorContextWithStandardError(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "")
	wrapper := newResticWrapper(nil, "", false, profile, "", nil, nil)
	require.NotNil(t, wrapper)
	ctx := wrapper.getErrorContext(errors.New("test error message 1"))
	assert.Equal(t, "test error message 1", ctx.Message)
	assert.Equal(t, "", ctx.ExitCode)
	assert.Equal(t, "", ctx.CommandLine)
	assert.Equal(t, "", ctx.Stderr)
}

func TestGetErrorContextWithCommandError(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "")
	wrapper := newResticWrapper(nil, "", false, profile, "", nil, nil)
	require.NotNil(t, wrapper)

	def := shellCommandDefinition{
		command:    "command",
		args:       []string{"arg1"},
		publicArgs: []string{"publicArg1"},
	}
	ctx := wrapper.getErrorContext(newCommandError(def, "stderr", errors.New("test error message 2")))
	assert.Equal(t, "test error message 2", ctx.Message)
	assert.Equal(t, "-1", ctx.ExitCode)
	assert.Equal(t, "\"command\" \"publicArg1\"", ctx.CommandLine)
	assert.Equal(t, "stderr", ctx.Stderr)
}

func TestGetProfileEnvironment(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "TestProfile")
	wrapper := newResticWrapper(nil, "", false, profile, "TestCommand", nil, nil)
	require.NotNil(t, wrapper)

	env := wrapper.getProfileEnvironment()
	assert.ElementsMatch(t, []string{"PROFILE_NAME=TestProfile", "PROFILE_COMMAND=TestCommand"}, env)
}

func TestGetFailEnvironmentNoError(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "")
	wrapper := newResticWrapper(nil, "", false, profile, "", nil, nil)
	require.NotNil(t, wrapper)

	env := wrapper.getFailEnvironment(nil)
	assert.Empty(t, env)
}

func TestGetFailEnvironmentWithStandardError(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "")
	wrapper := newResticWrapper(nil, "", false, profile, "", nil, nil)
	require.NotNil(t, wrapper)

	env := wrapper.getFailEnvironment(errors.New("test error message 3"))
	assert.ElementsMatch(t, []string{"ERROR=test error message 3", "ERROR_MESSAGE=test error message 3"}, env)
}

func TestGetFailEnvironmentWithCommandError(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "")
	wrapper := newResticWrapper(nil, "", false, profile, "", nil, nil)
	require.NotNil(t, wrapper)

	def := shellCommandDefinition{
		command:    "command",
		args:       []string{"arg1"},
		publicArgs: []string{"publicArg1"},
	}
	env := wrapper.getFailEnvironment(newCommandError(def, "stderr", errors.New("test error message 4")))
	assert.ElementsMatch(t, []string{
		"ERROR=test error message 4",
		"ERROR_MESSAGE=test error message 4",
		"ERROR_COMMANDLINE=\"command\" \"publicArg1\"",
		"ERROR_EXIT_CODE=-1",
		"ERROR_STDERR=stderr",
		"RESTIC_STDERR=stderr",
	}, env)
}

func popUntilPrefix(prefix string, log *clog.MemoryHandler) (line string) {
	for !strings.HasPrefix(line, prefix) && len(log.Logs()) > 0 {
		line = log.Pop()
	}
	return
}

func TestRunInitCopyCommand(t *testing.T) {
	testCases := []struct {
		profile      *config.Profile
		expectedInit string
		expectedCopy string
	}{
		{
			profile: &config.Profile{
				Name:         "profile",
				Repository:   config.NewConfidentialValue("repo_origin"),
				PasswordFile: "password_origin",
				Copy: &config.CopySection{
					InitializeCopyChunkerParams: true,
					Repository:                  config.NewConfidentialValue("repo_copy"),
					PasswordFile:                "password_copy",
				},
			},
			expectedInit: "dry-run: test init --copy-chunker-params --password-file password_copy --password-file2 password_origin --repo repo_copy --repo2 repo_origin",
			expectedCopy: "dry-run: test copy --password-file password_origin --password-file2 password_copy --repo repo_origin --repo2 repo_copy",
		},
		{
			profile: &config.Profile{
				Name:         "profile",
				Repository:   config.NewConfidentialValue("repo_origin"),
				PasswordFile: "password_origin",
				Copy: &config.CopySection{
					InitializeCopyChunkerParams: false,
					Repository:                  config.NewConfidentialValue("repo_copy"),
					PasswordFile:                "password_copy",
				},
			},
			expectedInit: "dry-run: test init --password-file password_copy --repo repo_copy",
			expectedCopy: "dry-run: test copy --password-file password_origin --password-file2 password_copy --repo repo_origin --repo2 repo_copy",
		},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			// We use the logger to run our test. It kind of sucks but does the job without having to tweak
			// the code to send the command line parameters somewhere
			defaultLogger := clog.GetDefaultLogger()
			mem := clog.NewMemoryHandler()
			clog.SetDefaultLogger(clog.NewLogger(mem))
			defer clog.SetDefaultLogger(defaultLogger)

			wrapper := newResticWrapper(config.NewGlobal(), "test", true, testCase.profile, "copy", nil, nil)
			// 1. run init command with copy profile
			err := wrapper.runInitializeCopy()
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedInit, popUntilPrefix("dry-run:", mem))

			// 2. run copy command
			err = wrapper.runCommand(constants.CommandCopy)
			require.NoError(t, err)

			// the latest message is saying the profile is finished
			assert.Equal(t, testCase.expectedCopy, popUntilPrefix("dry-run:", mem))
		})
	}
}
