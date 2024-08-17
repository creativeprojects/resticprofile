package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/monitor/mocks"
	"github.com/creativeprojects/resticprofile/monitor/status"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidResticArgumentsList(t *testing.T) {
	t.Parallel()

	wrapper := &resticWrapper{}

	for _, command := range restic.CommandNames() {
		t.Run(command, func(t *testing.T) {
			resticVersion := wrapper.getResticVersion()
			cmd, _ := restic.GetCommandForVersion(command, resticVersion, true)
			require.NotNil(t, cmd)
			defaultOptions := restic.GetDefaultOptionsForVersion(resticVersion, true)
			require.NotEmpty(t, defaultOptions)

			arguments := wrapper.validResticArgumentsList(command)

			for _, options := range [][]restic.Option{cmd.GetOptions(), defaultOptions} {
				for _, option := range options {
					if option.AvailableForOS() {
						if option.Name != "" {
							assert.Contains(t, arguments, fmt.Sprintf("--%s", option.Name))
						}
						if option.Alias != "" {
							assert.Contains(t, arguments, fmt.Sprintf("-%s", option.Alias))
						}
					} else {
						if option.Name != "" {
							assert.NotContains(t, arguments, fmt.Sprintf("--%s", option.Name))
						}
						if option.Alias != "" {
							assert.NotContains(t, arguments, fmt.Sprintf("-%s", option.Alias))
						}
					}
				}
			}
		})
	}
}

func TestVersionedResticArgumentsList(t *testing.T) {
	t.Parallel()

	wrapper := &resticWrapper{global: new(config.Global)}

	wrapper.global.ResticVersion = "0.14"
	arguments := wrapper.validResticArgumentsList("init")
	assert.Contains(t, arguments, "--from-repo")
	assert.Contains(t, arguments, "--repo2") // filter keeps legacy (removed) flags

	wrapper.global.ResticVersion = "0.13"
	arguments = wrapper.validResticArgumentsList("init")
	assert.Contains(t, arguments, "--repo2")
	assert.NotContains(t, arguments, "--from-repo") // exists not yet in restic 0.13
}

func TestValidArgumentsFilter(t *testing.T) {
	t.Parallel()

	wrapper := &resticWrapper{}
	validArgs := collect.All(wrapper.validResticArgumentsList(constants.CommandBackup), func(arg string) bool {
		return arg != "-x" && arg != "--xxx"
	})
	require.NotEmpty(t, validArgs)
	filter := wrapper.validArgumentsFilter(validArgs)

	for _, arg := range validArgs {
		var args []string
		list := []string{
			"-x",
			"-x=1",
			"-x 2 extra x=y",
			"--xxx",
			"--xxx=v",
			"--xxx k=v",
			arg,
			arg + "=1",
			arg + " ka=va",
			arg + " ka=va extra2",
		}

		for i := 0; i < 20; i++ {
			rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
			args = args[:0]
			for _, item := range list {
				args = append(args, strings.Split(item, " ")...)
			}

			// Test filter when extra values are allowed (args with values must use --arg=value)
			filtered := filter(args, true)

			assert.Len(t, filtered, 11)
			for _, item := range []string{arg, arg + "=1", "ka=va", "k=v", "x=y", "extra", "extra2", "2"} {
				assert.Contains(t, filtered, item)
			}
			for _, item := range []string{"-x", "--xxx", "--xxx=v"} {
				assert.NotContains(t, filtered, item)
			}

			// Test filter when extra values are disallowed (args with values can be --arg=value or --arg value)
			filtered = filter(args, false)

			assert.Len(t, filtered, 6)
			assert.Subset(t, []string{arg, arg + "=1", "ka=va"}, filtered)
			for _, item := range []string{"-x", "--xxx", "--xxx=v", "x=y", "k=v", "extra", "extra2", "2"} {
				assert.NotContains(t, filtered, item)
			}
		}
	}
}

func TestFilteredArgumentsRegression(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
		t.Skip()
	}

	tests := []struct {
		format, config string
		expected       map[string][]string
	}{
		{
			format: "toml",
			config: `
				version = "1"
				
				[default]
				password-command = 'echo password'
				initialize = true
				no-error-on-warning = true
				repository = 'backup'
				
				[default.backup]
				source = [
					'test-folder',
					'test-folder-2'
				]`,
			expected: map[string][]string{
				"backup": {"backup", "--password-command=echo\\ password", "--repo=backup", "test-folder", "test-folder-2"},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			cfg, err := config.Load(strings.NewReader(test.config), test.format)
			require.NoError(t, err)
			profile, err := cfg.GetProfile("default")
			require.NoError(t, err)
			wrapper := newResticWrapper(&Context{
				flags:   commandLineFlags{dryRun: true},
				binary:  "restic",
				profile: profile,
				command: "test",
			})

			for command, commandline := range test.expected {
				args := profile.GetCommandFlags(command)
				cmd := wrapper.prepareCommand(command, args, true)

				assert.Equal(t, commandline, cmd.args)
			}
		})
	}
}

func TestGetEmptyEnvironment(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  "restic",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	env := wrapper.getEnvironment(false)
	assert.Empty(t, env)
}

func TestGetSingleEnvironment(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]config.ConfidentialValue{
		"User": config.NewConfidentialValue("me"),
	}
	profile.ResolveConfiguration()
	ctx := &Context{
		binary:  "restic",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	env := wrapper.getEnvironment(false)
	assert.Equal(t, []string{"USER=me"}, env)
}

func TestGetMultipleEnvironment(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Environment = map[string]config.ConfidentialValue{
		"User":     config.NewConfidentialValue("me"),
		"Password": config.NewConfidentialValue("secret"),
	}
	profile.ResolveConfiguration()
	ctx := &Context{
		binary:  "restic",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)

	t.Run("getEnvironment", func(t *testing.T) {
		env := wrapper.getEnvironment(false)
		assert.Len(t, env, 2)
		assert.Contains(t, env, "USER=me")
		assert.Contains(t, env, "PASSWORD=secret")
	})

	t.Run("stringifyEnvironment", func(t *testing.T) {
		env := profile.GetEnvironment(false)
		str := wrapper.stringifyEnvironment(env)
		assert.Equal(t, "PASSWORD=×××\nUSER=me\n", str)
		assert.Equal(t, "secret", env.Get("PASSWORD"))
	})
}

func TestPreProfileScriptFail(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.RunBefore = []string{"exit 1"} // this should both work on unix shell and windows batch
	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-before on profile 'name': exit status 1")
}

func TestPostProfileScriptFail(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"exit 1"} // this should both work on unix shell and windows batch
	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "run-after on profile 'name': exit status 1")
}

func TestRunEchoProfile(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}

func TestPostProfileAfterFail(t *testing.T) {
	t.Parallel()

	testFile := "TestPostProfileAfterFail.txt"
	_ = os.Remove(testFile)
	profile := config.NewProfile(nil, "name")
	profile.RunAfter = []string{"echo failed > " + testFile}
	ctx := &Context{
		binary:  "exit",
		profile: profile,
		command: "1",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "1 on profile 'name': exit status 1")
	assert.NoFileExistsf(t, testFile, "the run-after script should not have been running")
	_ = os.Remove(testFile)
}

func TestPostFailProfile(t *testing.T) {
	t.Parallel()

	testFile := "TestPostFailProfile.txt"
	_ = os.Remove(testFile)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo failed > " + testFile}
	ctx := &Context{
		binary:  "exit",
		profile: profile,
		command: "1",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.EqualError(t, err, "1 on profile 'name': exit status 1")
	assert.FileExistsf(t, testFile, "the run-after-fail script has not been running")
	_ = os.Remove(testFile)
}

func TestFinallyProfile(t *testing.T) {
	t.Parallel()

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
		ctx := &Context{
			binary:  "echo",
			profile: profile,
			command: "backup",
		}
		wrapper := newResticWrapper(ctx)
		err := wrapper.runProfile()
		assert.NoError(t, err)
		assertFileEquals(t, "finally")
	})

	t.Run("on-backup-only", func(t *testing.T) {
		newProfile()
		profile.RunFinally = nil
		ctx := &Context{
			binary:  "echo",
			profile: profile,
			command: "backup",
		}
		wrapper := newResticWrapper(ctx)
		err := wrapper.runProfile()
		assert.NoError(t, err)
		assertFileEquals(t, "finally-backup")
	})

	t.Run("on-error", func(t *testing.T) {
		newProfile()
		ctx := &Context{
			binary:  "exit",
			profile: profile,
			command: "1",
		}
		wrapper := newResticWrapper(ctx)
		err := wrapper.runProfile()
		assert.EqualError(t, err, "1 on profile 'name': exit status 1")
		assertFileEquals(t, "finally")
	})
}

func Example_runProfile() {
	term.SetOutput(os.Stdout)
	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	if err != nil {
		log.Fatal(err)
	}
	// Output: test
}

func TestRunRedirectOutputOfEchoProfile(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "test", strings.TrimSpace(buffer.String()))
}

func TestDryRun(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	wrapper := newResticWrapper(&Context{
		flags:   commandLineFlags{dryRun: true},
		binary:  "echo",
		profile: profile,
		command: "test",
	})
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "", buffer.String())
}

func TestEnvProfileName(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "TestEnvProfileName")
	profile.RunBefore = []string{"echo profile name = $PROFILE_NAME"}

	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "profile name = TestEnvProfileName\ntest\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvProfileCommand(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunBefore = []string{"echo profile command = $PROFILE_COMMAND"}

	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test-command",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.NoError(t, err)
	assert.Equal(t, "profile command = test-command\ntest-command\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvError(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo error: $ERROR_MESSAGE"}

	ctx := &Context{
		binary:  "exit",
		profile: profile,
		command: "1",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "error: 1 on profile 'name': exit status 1\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvErrorCommandLine(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo cmd: $ERROR_COMMANDLINE"}

	ctx := &Context{
		binary:  "exit",
		profile: profile,
		command: "1",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "cmd: \"exit\" \"1\"\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvErrorExitCode(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo exit-code: $ERROR_EXIT_CODE"}

	ctx := &Context{
		binary:  "exit",
		profile: profile,
		command: "5",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "exit-code: 5\n", strings.ReplaceAll(buffer.String(), "\r\n", "\n"))
}

func TestEnvStderr(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	profile := config.NewProfile(nil, "name")
	profile.RunAfterFail = []string{"echo stderr: $ERROR_STDERR"}

	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "command",
		request: Request{arguments: []string{"--stderr", "error_message", "--exit", "1"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.Error(t, err)
	assert.Equal(t, "stderr: error_message", strings.TrimSpace(strings.ReplaceAll(buffer.String(), "\r\n", "\n")))
}

func TestRunProfileWithSetPIDCallback(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Lock = filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestRunProfileWithSetPIDCallback", time.Now().UnixNano(), os.Getpid()))
	t.Logf("lockfile = %s", profile.Lock)
	ctx := &Context{
		binary:  "echo",
		profile: profile,
		command: "test",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runProfile()
	assert.NoError(t, err)
}

func TestInitializeNoError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runInitialize()
	require.NoError(t, err)
}

func TestInitializeWithError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "10"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runInitialize()
	require.Error(t, err)
}

func TestInitializeCopyNoError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Copy = &config.CopySection{InitializeCopyChunkerParams: maybe.False()}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runInitializeCopy()
	require.NoError(t, err)
}

func TestInitializeCopyWithError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Copy = &config.CopySection{InitializeCopyChunkerParams: maybe.False()}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "10"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runInitializeCopy()
	require.Error(t, err)
}

func TestCheckNoError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runCheck()
	require.NoError(t, err)
}

func TestCheckWithError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "10"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runCheck()
	require.Error(t, err)
}

func TestRetentionNoError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runRetention()
	require.NoError(t, err)
}

func TestRetentionWithError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "10"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runRetention()
	require.Error(t, err)
}

func TestBackupWithStreamSource(t *testing.T) {
	t.Parallel()

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
		ctx := &Context{
			binary:  mockBinary,
			profile: profile,
			command: "stdin-test",
			sigChan: signals,
		}
		wrapper = newResticWrapper(ctx)
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
		wrapper.moreArgs = []string{"--sleep", "15000"}

		profile.Backup.StdinCommand = append(fillBufferCommand(), "exit 2")
		profile.ResolveConfiguration()

		start := time.Now()
		_, err := run(t, wrapper)
		assert.Less(t, time.Since(start), time.Second*12, "timeout, interrupt not sent to restic")

		require.NotNil(t, err)
		assert.Contains(t, expectedInterruptedError, err.Error())
	})

	t.Run("CanTerminateStreamSource", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("signal handling is not supported on Windows")
		}
		profile, wrapper := profileAndWrapper(t)
		profile.Backup.StdinCommand = append(fillBufferCommand(), mockBinary+" cmd --sleep 6000")
		profile.ResolveConfiguration()

		go func() {
			time.Sleep(500 * time.Millisecond)
			wrapper.sigChan <- os.Interrupt
		}()
		start := time.Now()
		_, err := run(t, wrapper)
		assert.Less(t, time.Since(start), time.Second*5, "timeout, interrupt not sent to stdin-command")

		require.NotNil(t, err)
		assert.Error(t, err)
		if err.Error() != "stdin-test on profile 'name': io: read/write on closed pipe" &&
			err.Error() != "stdin-test on profile 'name': signal: interrupt" {
			t.Errorf("unexpected error: %s", err)
		}
	})
}

func TestBackupWithSuccess(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runCommand("backup")
	require.NoError(t, err)
}

func TestBackupWithError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "1"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithResticLockFailureRetried(t *testing.T) {
	t.Parallel()

	lockWait := constants.MinResticLockRetryDelay + time.Second
	lockMessage := "unable to create lock in backend: repository is already locked exclusively by PID 60485 on VM by user (UID 503, GID 23)" + platform.LineSeparator +
		"lock was created at 2023-09-24 15:29:57 (69.406ms ago)" + platform.LineSeparator +
		"storage ID c8a44e77" + platform.LineSeparator +
		"the `unlock` command can be used to remove stale locks" + platform.LineSeparator
	tempfile := filepath.Join(t.TempDir(), "TestBackupWithResticLockFailureRetried.txt")
	err := os.WriteFile(tempfile, []byte(lockMessage), 0o600)
	require.NoError(t, err)
	defer os.Remove(tempfile)

	sigChan := make(chan os.Signal, 1)
	global := &config.Global{
		ResticLockRetryAfter: lockWait,
	}
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	ctx := &Context{
		global:  global,
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--stderr", "@" + tempfile, "--exit", "1"}},
		sigChan: sigChan,
	}
	wrapper := newResticWrapper(ctx)
	wrapper.lockWait = &lockWait
	wrapper.startTime = time.Now()

	err = wrapper.runCommand("backup")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, errInterrupt)
}

func TestBackupWithResticLockFailureCancelled(t *testing.T) {
	t.Parallel()

	lockWait := constants.MinResticLockRetryDelay + time.Second
	lockMessage := "unable to create lock in backend: repository is already locked exclusively by PID 60485 on VM by user (UID 503, GID 23)" + platform.LineSeparator +
		"lock was created at 2023-09-24 15:29:57 (69.406ms ago)" + platform.LineSeparator +
		"storage ID c8a44e77" + platform.LineSeparator +
		"the `unlock` command can be used to remove stale locks" + platform.LineSeparator
	tempfile := filepath.Join(t.TempDir(), "TestBackupWithResticLockFailureCancelled.txt")
	err := os.WriteFile(tempfile, []byte(lockMessage), 0o600)
	require.NoError(t, err)
	defer os.Remove(tempfile)

	sigChan := make(chan os.Signal, 1)
	global := &config.Global{
		ResticLockRetryAfter: lockWait,
	}
	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	ctx := &Context{
		global:  global,
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--stderr", "@" + tempfile, "--exit", "1"}},
		sigChan: sigChan,
	}
	wrapper := newResticWrapper(ctx)
	wrapper.lockWait = &lockWait
	wrapper.startTime = time.Now()

	timer := time.AfterFunc(1*time.Second, func() {
		sigChan <- os.Interrupt
	})
	defer timer.Stop()

	err = wrapper.runCommand("backup")
	assert.ErrorIs(t, err, errInterrupt)
}

func TestBackupWithNoConfiguration(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "1"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithNoConfigurationButStatusFile(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.StatusFile = "status.json"
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "1"}},
	}
	wrapper := newResticWrapper(ctx)
	wrapper.addProgress(status.NewProgress(profile, status.NewStatus("status.json")))
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithWarningAsError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(nil, "name")
	profile.Backup = &config.BackupSection{}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "3"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runCommand("backup")
	require.Error(t, err)
}

func TestBackupWithSupressedWarnings(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{NoErrorOnWarning: true}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
		request: Request{arguments: []string{"--exit", "3"}},
	}
	wrapper := newResticWrapper(ctx)
	err := wrapper.runCommand("backup")
	require.NoError(t, err)
}

func TestRunShellCommands(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{}
	profile.Check = &config.SectionWithScheduleAndMonitoring{}
	profile.Copy = &config.CopySection{}
	profile.Forget = &config.SectionWithScheduleAndMonitoring{}
	profile.Init = &config.InitSection{}
	profile.Prune = &config.SectionWithScheduleAndMonitoring{}
	for name := range profile.OtherSections {
		profile.OtherSections[name] = new(config.GenericSection)
	}

	sections := make(map[string]*config.RunShellCommandsSection)
	for name, s := range config.GetDeclaredSectionsWith[config.RunShellCommands](profile) {
		sections[name] = s.GetRunShellCommands()
	}
	require.Greater(t, len(sections), 10)

	for command, section := range sections {
		t.Run(fmt.Sprintf("run-before '%s'", command), func(t *testing.T) {
			section.RunBefore = []string{"exit 2"}
			ctx := &Context{
				binary:  mockBinary,
				profile: profile,
				command: command,
			}
			wrapper := newResticWrapper(ctx)
			err := wrapper.runProfile()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "exit status 2")

			section.RunBefore = []string{""}
			wrapper = newResticWrapper(ctx)
			err = wrapper.runProfile()
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("run-after '%s'", command), func(t *testing.T) {
			section.RunAfter = []string{"exit 2"}
			ctx := &Context{
				binary:  mockBinary,
				profile: profile,
				command: command,
			}
			wrapper := newResticWrapper(ctx)
			err := wrapper.runProfile()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "exit status 2")

			section.RunAfter = []string{""}
			wrapper = newResticWrapper(ctx)
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
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "backup",
		request: Request{arguments: []string{"--stderr", "--error-line--"}},
	}
	wrapper := newResticWrapper(ctx)

	err := wrapper.runProfile()
	require.NoError(t, err)
	assert.Contains(t, buffer.String(), "detected error in backup")
}

func TestRunStreamErrorHandlerDoesNotBreakCommand(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{}
	profile.StreamError = []config.StreamErrorSection{{Pattern: ".+error-line.+", Run: "exit 1"}}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "backup",
		request: Request{arguments: []string{"--stderr", "--error-line--"}},
	}
	wrapper := newResticWrapper(ctx)

	err := wrapper.runProfile()
	require.NoError(t, err)
}

func TestStreamErrorHandlerWithInvalidRegex(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Backup = &config.BackupSection{}
	profile.StreamError = []config.StreamErrorSection{{Pattern: "(", Run: "echo pass"}}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "backup",
		request: Request{arguments: []string{}},
	}
	wrapper := newResticWrapper(ctx)

	err := wrapper.runProfile()
	assert.EqualError(t, err, "backup on profile 'name': stream error callback: echo pass failed to register (: error parsing regexp: missing closing ): `(`")
}

func TestCanRetryAfterErrorDontFailWhenNoOutputAnalysis(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "backup",
	}
	wrapper := newResticWrapper(ctx)
	summary := monitor.Summary{}
	retry, err := wrapper.canRetryAfterError("backup", summary)
	assert.False(t, retry)
	assert.NoError(t, err)
}

func TestCanRetryAfterRemoteStaleLockFailure(t *testing.T) {
	t.Parallel()

	lockedSince := time.Duration(0)
	mockOutput := mocks.NewOutputAnalysis(t)
	mockOutput.EXPECT().ContainsRemoteLockFailure().Return(true)
	mockOutput.EXPECT().GetRemoteLockedSince().RunAndReturn(func() (time.Duration, bool) { return lockedSince, lockedSince > 0 })

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Repository = config.NewConfidentialValue("my-repo")
	profile.ForceLock = true
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "backup",
	}
	wrapper := newResticWrapper(ctx)
	wrapper.startTime = time.Now()
	wrapper.global.ResticStaleLockAge = 0 // disable stale lock handling

	// No retry when no stale remote-lock failure
	assert.True(t, mockOutput.ContainsRemoteLockFailure())
	retry, sleep := wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), sleep)

	// Ignores stale lock when disabled
	lockedSince = constants.MinResticStaleLockAge
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), sleep)

	// Ignores non-stale lock
	lockedSince = constants.MinResticStaleLockAge - time.Nanosecond
	wrapper.global.ResticStaleLockAge = time.Millisecond
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), sleep)

	// Unlocks stale lock
	lockedSince = constants.MinResticStaleLockAge
	assert.False(t, wrapper.doneTryUnlock)
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)
	assert.Equal(t, time.Duration(0), sleep)
	assert.True(t, wrapper.doneTryUnlock)

	// Unlock is run only once
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), sleep)

	// Unlock is not run when ForceLock is disabled
	wrapper.doneTryUnlock = false
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)
	assert.Equal(t, time.Duration(0), sleep)

	profile.ForceLock = false
	wrapper.doneTryUnlock = false
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), sleep)
}

func TestCanRetryAfterRemoteLockFailure(t *testing.T) {
	t.Parallel()

	lockFailure := false
	mockOutput := mocks.NewOutputAnalysis(t)
	mockOutput.EXPECT().ContainsRemoteLockFailure().RunAndReturn(func() bool { return lockFailure })
	mockOutput.EXPECT().GetRemoteLockedBy().Return(t.Name(), true)
	mockOutput.EXPECT().GetRemoteLockedSince().Return(5*time.Minute, true)

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Repository = config.NewConfidentialValue("my-repo")
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "backup",
	}
	wrapper := newResticWrapper(ctx)
	wrapper.startTime = time.Now()
	wrapper.global.ResticLockRetryAfter = 0 // disable remote lock retry

	// No retry when no remote-lock failure
	assert.False(t, mockOutput.ContainsRemoteLockFailure())
	retry, sleep := wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), sleep)

	// No retry when lockWait is nil
	lockFailure = true
	assert.True(t, mockOutput.ContainsRemoteLockFailure())
	retry, _ = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// No retry when disabled
	wrapper.maxWaitOnLock(constants.MinResticLockRetryDelay + 50*time.Millisecond)
	retry, _ = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)

	// No retry when no time left
	wrapper.maxWaitOnLock(constants.MinResticLockRetryDelay - time.Nanosecond)
	wrapper.global.ResticLockRetryAfter = constants.MinResticLockRetryDelay // enable remote lock retry
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), sleep)

	// Retry is acceptable when there is enough remaining time for the delay (ResticLockRetryAfter)
	wrapper.maxWaitOnLock(constants.MinResticLockRetryDelay + 50*time.Millisecond)
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)
	assert.Equal(t, constants.MinResticLockRetryDelay, sleep)

	wrapper.maxWaitOnLock(constants.MaxResticLockRetryDelay + 50*time.Millisecond)
	wrapper.global.ResticLockRetryAfter = 2 * constants.MaxResticLockRetryDelay
	retry, sleep = wrapper.canRetryAfterRemoteLockFailure(mockOutput)
	assert.True(t, retry)
	assert.Equal(t, constants.MaxResticLockRetryDelay, sleep)
}

func TestCanUseResticLockRetry(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Repository = config.NewConfidentialValue("my-repo")
	emptyArgs := shell.NewArgs()
	argMatcher := regexp.MustCompile(".*retry-lock.*").MatchString

	getWrapper := func() *resticWrapper {
		wrapper := newResticWrapper(&Context{
			flags:   commandLineFlags{dryRun: true},
			binary:  "restic",
			profile: profile,
			command: constants.CommandBackup,
		})
		wrapper.startTime = time.Now()
		wrapper.global.ResticLockRetryAfter = 1 * time.Minute
		wrapper.global.ResticVersion = "0.16"
		return wrapper
	}

	t.Run("SubtractsResticLockRetryAfter", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.maxWaitOnLock(10*time.Minute + 30*time.Second)
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.Contains(t, command.args, "--retry-lock=9m")
	})

	t.Run("SubtractsRemainingTime", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.maxWaitOnLock(10*time.Minute + 30*time.Second)
		wrapper.executionTime = 3 * time.Minute
		wrapper.startTime = wrapper.startTime.Add(-5 * time.Minute) // 2 minutes for locks, 3 minutes for execution
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.Contains(t, command.args, "--retry-lock=7m")
	})

	t.Run("10MinutesIsMax", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.maxWaitOnLock(10 * constants.MaxResticLockRetryTimeArgument)
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.Contains(t, command.args, "--retry-lock=10m")
	})

	t.Run("1MinuteIsMin", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.maxWaitOnLock(2*time.Minute + 30*time.Second)
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.Contains(t, command.args, "--retry-lock=1m")
		assert.True(t, slices.ContainsFunc(command.args, argMatcher))

		wrapper.maxWaitOnLock(2 * time.Minute)
		command = wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.False(t, slices.ContainsFunc(command.args, argMatcher))
	})

	t.Run("NotAddedInRestic15", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.global.ResticVersion = "0.15"
		wrapper.maxWaitOnLock(30 * time.Minute)
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.False(t, slices.ContainsFunc(command.args, argMatcher))
	})

	t.Run("NotAddedWithoutFilter", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.global.FilterResticFlags = false
		wrapper.maxWaitOnLock(30 * time.Minute)
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.False(t, slices.ContainsFunc(command.args, argMatcher))
	})

	t.Run("NotOverwritingAlreadyProvided", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.moreArgs = []string{"--retry-lock", "25m"}
		wrapper.maxWaitOnLock(30 * time.Minute)
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.Subset(t, command.args, wrapper.moreArgs)
		assert.NotContains(t, command.args, "--retry-lock=10m")
	})

	t.Run("Regression-WorksWithExtraValues", func(t *testing.T) {
		wrapper := getWrapper()
		wrapper.moreArgs = []string{"/some/path", "some-other-option"}
		wrapper.maxWaitOnLock(30 * time.Minute)
		command := wrapper.prepareCommand(constants.CommandBackup, emptyArgs, true)
		assert.Subset(t, command.args, wrapper.moreArgs)
		assert.Contains(t, command.args, "--retry-lock=10m")
	})
}

func TestLocksAndLockWait(t *testing.T) {
	profile := config.NewProfile(nil, "name")
	profile.Lock = filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestLockWait", time.Now().UnixNano(), os.Getpid()))
	defer os.Remove(profile.Lock)

	term.SetOutput(os.Stdout)

	ctx1 := &Context{
		binary:  mockBinary,
		profile: profile,
		command: constants.CommandBackup,
		request: Request{arguments: []string{"--sleep", "1500"}},
	}
	ctx2 := &Context{
		binary:  mockBinary,
		profile: profile,
		command: constants.CommandBackup,
	}
	ctx3 := &Context{
		binary:  mockBinary,
		profile: profile,
		command: constants.CommandBackup,
	}
	w1 := newResticWrapper(ctx1)
	w2 := newResticWrapper(ctx2)
	w3 := newResticWrapper(ctx3)

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
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "TestProfile")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "TestCommand",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)
	hookCtx := wrapper.getContext()
	assert.Equal(t, "TestProfile", hookCtx.ProfileName)
	assert.Equal(t, "TestCommand", hookCtx.ProfileCommand)
	assert.Equal(t, "", hookCtx.Error.Message)
	assert.Equal(t, "", hookCtx.Error.ExitCode)
	assert.Equal(t, "", hookCtx.Error.CommandLine)
	assert.Equal(t, "", hookCtx.Error.Stderr)
}

func TestGetContextWithError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "TestProfile")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "TestCommand",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)
	hookCtx := wrapper.getContextWithError(nil)
	assert.Equal(t, "TestProfile", hookCtx.ProfileName)
	assert.Equal(t, "TestCommand", hookCtx.ProfileCommand)
	assert.Equal(t, "", hookCtx.Error.Message)
	assert.Equal(t, "", hookCtx.Error.ExitCode)
	assert.Equal(t, "", hookCtx.Error.CommandLine)
	assert.Equal(t, "", hookCtx.Error.Stderr)
}

func TestGetErrorContext(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)
	hookCtx := wrapper.getErrorContext(nil)
	assert.Equal(t, "", hookCtx.Message)
	assert.Equal(t, "", hookCtx.ExitCode)
	assert.Equal(t, "", hookCtx.CommandLine)
	assert.Equal(t, "", hookCtx.Stderr)
}

func TestGetErrorContextWithStandardError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)
	hookCtx := wrapper.getErrorContext(errors.New("test error message 1"))
	assert.Equal(t, "test error message 1", hookCtx.Message)
	assert.Equal(t, "", hookCtx.ExitCode)
	assert.Equal(t, "", hookCtx.CommandLine)
	assert.Equal(t, "", hookCtx.Stderr)
}

func TestGetErrorContextWithCommandError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)

	def := shellCommandDefinition{
		command:    "command",
		args:       []string{"arg1"},
		publicArgs: []string{"publicArg1"},
	}
	hookCtx := wrapper.getErrorContext(newCommandError(def, "stderr", errors.New("test error message 2")))
	assert.Equal(t, "test error message 2", hookCtx.Message)
	assert.Equal(t, "-1", hookCtx.ExitCode)
	assert.Equal(t, "\"command\" \"publicArg1\"", hookCtx.CommandLine)
	assert.Equal(t, "stderr", hookCtx.Stderr)
}

func TestGetProfileEnvironment(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "TestProfile")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "TestCommand",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)

	env := wrapper.getProfileEnvironment()
	assert.ElementsMatch(t, []string{"PROFILE_NAME=TestProfile", "PROFILE_COMMAND=TestCommand"}, env)
}

func TestGetFailEnvironmentNoError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)

	env := wrapper.getFailEnvironment(nil)
	assert.Empty(t, env)
}

func TestGetFailEnvironmentWithStandardError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	require.NotNil(t, wrapper)

	env := wrapper.getFailEnvironment(errors.New("test error message 3"))
	assert.ElementsMatch(t, []string{"ERROR=test error message 3", "ERROR_MESSAGE=test error message 3"}, env)
}

func TestGetFailEnvironmentWithCommandError(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "")
	ctx := &Context{
		binary:  "",
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
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
	makeProfile := func(copyChunkerParams maybe.Bool, resticVersion string) (p *config.Profile) {
		p = &config.Profile{
			Name:         "profile",
			Repository:   config.NewConfidentialValue("repo_origin"),
			PasswordFile: "password_origin",
			Copy: &config.CopySection{
				InitializeCopyChunkerParams: copyChunkerParams,
				Repository:                  config.NewConfidentialValue("repo_copy"),
				PasswordFile:                "password_copy",
			},
		}
		require.NoError(t, p.SetResticVersion(resticVersion))
		return p
	}

	testCases := []struct {
		profile      *config.Profile
		expectedInit string
		expectedCopy string
	}{
		{
			profile:      makeProfile(maybe.True(), "0.13"),
			expectedInit: "dry-run: test init --copy-chunker-params --password-file=password_copy --password-file2=password_origin --repo=repo_copy --repo2=repo_origin",
			expectedCopy: "dry-run: test copy --password-file=password_origin --password-file2=password_copy --repo=repo_origin --repo2=repo_copy",
		},
		{
			profile:      makeProfile(maybe.False(), "0.13"),
			expectedInit: "dry-run: test init --password-file=password_copy --repo=repo_copy",
			expectedCopy: "dry-run: test copy --password-file=password_origin --password-file2=password_copy --repo=repo_origin --repo2=repo_copy",
		},
		{
			profile:      makeProfile(maybe.True(), "0.14"),
			expectedInit: "dry-run: test init --copy-chunker-params --from-password-file=password_origin --from-repo=repo_origin --password-file=password_copy --repo=repo_copy",
			expectedCopy: "dry-run: test copy --from-password-file=password_origin --from-repo=repo_origin --password-file=password_copy --repo=repo_copy",
		},
		{
			profile:      makeProfile(maybe.False(), "0.14"),
			expectedInit: "dry-run: test init --password-file=password_copy --repo=repo_copy",
			expectedCopy: "dry-run: test copy --from-password-file=password_origin --from-repo=repo_origin --password-file=password_copy --repo=repo_copy",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// We use the logger to run our test. It kind of sucks but does the job without having to tweak
			// the code to send the command line parameters somewhere
			defaultLogger := clog.GetDefaultLogger()
			mem := clog.NewMemoryHandler()
			clog.SetDefaultLogger(clog.NewLogger(mem))
			defer clog.SetDefaultLogger(defaultLogger)

			wrapper := newResticWrapper(&Context{
				flags:   commandLineFlags{dryRun: true},
				global:  config.NewGlobal(),
				binary:  "test",
				profile: testCase.profile,
				command: "copy",
			})
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

func TestCopyNoSnapshot(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Copy = &config.CopySection{}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	args := shell.NewArgs()
	cmd := wrapper.prepareCommand("copy", args, false)
	assert.Equal(t, []string{"copy"}, cmd.args)
}

func TestCopySnapshot(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	profile.Copy = &config.CopySection{Snapshots: []string{"snapshot1", "snapshot2"}}
	ctx := &Context{
		binary:  mockBinary,
		profile: profile,
		command: "",
	}
	wrapper := newResticWrapper(ctx)
	args := shell.NewArgs()
	cmd := wrapper.prepareCommand("copy", args, false)
	assert.Equal(t, []string{"copy", "snapshot1", "snapshot2"}, cmd.args)
}
