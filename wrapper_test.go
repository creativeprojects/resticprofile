package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/progress"
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
}

func TestGetEmptyEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Empty(t, env)
}

func TestGetSingleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]config.ConfidentialValue{
		"User": config.NewConfidentialValue("me"),
	}
	wrapper := newResticWrapper("restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Equal(t, []string{"USER=me"}, env)
}

func TestGetMultipleEnvironment(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]config.ConfidentialValue{
		"User":     config.NewConfidentialValue("me"),
		"Password": config.NewConfidentialValue("secret"),
	}
	wrapper := newResticWrapper("restic", false, profile, "test", nil, nil)
	env := wrapper.getEnvironment()
	assert.Len(t, env, 2)
	assert.Contains(t, env, "USER=me")
	assert.Contains(t, env, "PASSWORD=secret")
}

func TestPreProfileScriptFail(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.RunBefore = []string{"exit 1"} // this should both work on unix shell and windows batch
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-before on profile 'name': exit status 1")
}

func TestPostProfileScriptFail(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"exit 1"} // this should both work on unix shell and windows batch
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-after on profile 'name': exit status 1")
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
	assert.EqualError(t, err, "1 on profile 'name': exit status 1")
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
	assert.EqualError(t, err, "1 on profile 'name': exit status 1")
	assert.FileExistsf(t, testFile, "the run-after-fail script has not been running")
	_ = os.Remove(testFile)
}

func Example_runProfile() {
	term.SetOutput(os.Stdout)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	wrapper.runProfile()
	// Output: test
}

func TestRunRedirectOutputOfEchoProfile(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "test", strings.TrimSpace(buffer.String()))
}

func TestDryRun(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper("echo", true, profile, "test", nil, nil)
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
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
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
	wrapper := newResticWrapper("echo", false, profile, "test-command", nil, nil)
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
	wrapper := newResticWrapper("exit", false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "error: 1 on profile 'name': exit status 1\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvErrorCommandLine(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	if runtime.GOOS == "windows" {
		profile.RunAfterFail = []string{"echo cmd: %ERROR_COMMANDLINE%"}
	} else {
		profile.RunAfterFail = []string{"echo cmd: $ERROR_COMMANDLINE"}
	}
	wrapper := newResticWrapper("exit", false, profile, "1", nil, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "cmd: \"exit\" \"1\"\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvErrorExitCode(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	if runtime.GOOS == "windows" {
		profile.RunAfterFail = []string{"echo exit-code: %ERROR_EXIT_CODE%"}
	} else {
		profile.RunAfterFail = []string{"echo exit-code: $ERROR_EXIT_CODE"}
	}
	wrapper := newResticWrapper("exit", false, profile, "5", nil, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "exit-code: 5\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvStderr(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	if runtime.GOOS == "windows" {
		profile.RunAfterFail = []string{"echo stderr: %ERROR_STDERR%"}
	} else {
		profile.RunAfterFail = []string{"echo stderr: $ERROR_STDERR"}
	}
	wrapper := newResticWrapper(mockBinary, false, profile, "command", []string{"--stderr", "error_message", "--exit", "1"}, nil)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "stderr: error_message\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestRunProfileWithSetPIDCallback(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Lock = filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestRunProfileWithSetPIDCallback", time.Now().UnixNano(), os.Getpid()))
	t.Logf("lockfile = %s", profile.Lock)
	wrapper := newResticWrapper("echo", false, profile, "test", nil, nil)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}

func TestInitializeNoError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(mockBinary, false, profile, "", nil, nil)
	err := wrapper.runInitialize()
	require.NoError(t, err)
}

func TestInitializeWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(mockBinary, false, profile, "", []string{"--exit", "10"}, nil)
	err := wrapper.runInitialize()
	require.Error(t, err)
}

func TestCheckNoError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(mockBinary, false, profile, "", nil, nil)
	err := wrapper.runCheck()
	require.NoError(t, err)
}

func TestCheckWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(mockBinary, false, profile, "", []string{"--exit", "10"}, nil)
	err := wrapper.runCheck()
	require.Error(t, err)
}

func TestRetentionNoError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(mockBinary, false, profile, "", nil, nil)
	err := wrapper.runRetention()
	require.NoError(t, err)
}

func TestRetentionWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(mockBinary, false, profile, "", []string{"--exit", "10"}, nil)
	err := wrapper.runRetention()
	require.Error(t, err)
}

func TestBackupWithSuccess(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	wrapper := newResticWrapper(mockBinary, false, profile, "", nil, nil)
	err := wrapper.runCommand("backup")
	require.NoError(t, err)
}

func TestBackupWithError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	wrapper := newResticWrapper(mockBinary, false, profile, "", []string{"--exit", "1"}, nil)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithWarningAsError(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	wrapper := newResticWrapper(mockBinary, false, profile, "", []string{"--exit", "3"}, nil)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithSupressedWarnings(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{NoErrorOnWarning: true}
	wrapper := newResticWrapper(mockBinary, false, profile, "", []string{"--exit", "3"}, nil)
	err := wrapper.runCommand("backup")
	require.NoError(t, err)
}

func TestRunBeforeBackupFailed(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{RunBefore: []string{"exit 2"}}
	wrapper := newResticWrapper(mockBinary, false, profile, "backup", nil, nil)
	err := wrapper.runProfile()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exit status 2")
}

func TestRunAfterBackupFailed(t *testing.T) {
	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{RunAfter: []string{"exit 2"}}
	wrapper := newResticWrapper(mockBinary, false, profile, "backup", nil, nil)
	err := wrapper.runProfile()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exit status 2")
}

type mockOutputAnalysis struct {
	progress.OutputAnalysis
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
	wrapper := newResticWrapper(mockBinary, false, profile, "backup", nil, nil)
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
	wrapper := newResticWrapper(mockBinary, false, profile, "backup", nil, nil)
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

	w1 := newResticWrapper(mockBinary, false, profile, "backup", []string{"--sleep", "1500"}, nil)
	w2 := newResticWrapper(mockBinary, false, profile, "backup", nil, nil)
	w3 := newResticWrapper(mockBinary, false, profile, "backup", nil, nil)

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
