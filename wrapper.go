package main

import (
	"errors"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/lock"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/status"
	"github.com/creativeprojects/resticprofile/term"
)

type commandError struct {
	scd    shellCommandDefinition
	stderr string
	err    error
}

func newCommandError(command shellCommandDefinition, stderr string, err error) *commandError {
	return &commandError{
		scd:    command,
		stderr: stderr,
		err:    err,
	}
}

func (c *commandError) Error() string {
	return c.err.Error()
}

func (c *commandError) Commandline() string {
	args := ""
	if c.scd.args != nil && len(c.scd.args) > 0 {
		args = fmt.Sprintf(" \"%s\"", strings.Join(c.scd.args, "\" \""))
	}
	return fmt.Sprintf("\"%s\"%s", c.scd.command, args)
}

func (c *commandError) Stderr() string {
	return c.stderr
}

type resticWrapper struct {
	resticBinary string
	initialize   bool
	dryRun       bool
	noLock       bool
	lockWait     *time.Duration
	profile      *config.Profile
	global       *config.Global
	command      string
	moreArgs     []string
	sigChan      chan os.Signal
	setPID       func(pid int)

	// States
	startTime     time.Time
	doneTryUnlock bool
}

func newResticWrapper(
	resticBinary string,
	initialize bool,
	dryRun bool,
	profile *config.Profile,
	command string,
	moreArgs []string,
	c chan os.Signal,
) *resticWrapper {
	return &resticWrapper{
		resticBinary: resticBinary,
		initialize:   initialize,
		dryRun:       dryRun,
		noLock:       false,
		lockWait:     nil,
		profile:      profile,
		global:       config.NewGlobal(),
		command:      command,
		moreArgs:     moreArgs,
		sigChan:      c,
	}
}

// setGlobal sets the global section from config
func (r *resticWrapper) setGlobal(global *config.Global) {
	r.global = global
}

// ignoreLock configures resticWrapper to ignore the lock defined in profile
func (r *resticWrapper) ignoreLock() {
	r.noLock = true
	r.lockWait = nil
}

// ignoreLock configures resticWrapper to wait up to duration to acquire the lock defined in profile
func (r *resticWrapper) maxWaitOnLock(duration time.Duration) {
	r.noLock = false
	if duration > 0 {
		r.lockWait = &duration
	} else {
		r.lockWait = nil
	}
}

func (r *resticWrapper) runProfile() error {
	lockFile := r.profile.Lock
	if r.noLock {
		lockFile = ""
	}

	r.startTime = time.Now()

	err := lockRun(lockFile, r.profile.ForceLock, r.lockWait, func(setPID lock.SetPID) error {
		r.setPID = setPID
		return runOnFailure(
			func() error {
				var err error

				// pre-profile commands
				err = r.runProfilePreCommand()
				if err != nil {
					return err
				}

				// breaking change from 0.7.0 and 0.7.1:
				// run the initialization after the pre-profile commands
				if r.initialize && r.command != constants.CommandInit {
					_ = r.runInitialize()
					// it's ok for the initialize to error out when the repository exists
				}

				// pre-commands (for backup)
				if r.command == constants.CommandBackup {
					// Shell commands
					err = r.runPreCommand(r.command)
					if err != nil {
						return err
					}
					// Check
					if r.profile.Backup != nil && r.profile.Backup.CheckBefore {
						err = r.runCheck()
						if err != nil {
							return err
						}
					}
					// Retention
					if r.profile.Retention != nil && r.profile.Retention.BeforeBackup {
						err = r.runRetention()
						if err != nil {
							return err
						}
					}
				}

				// Main command
				err = r.runCommand(r.command)
				if err != nil {
					return err
				}

				// post-commands (for backup)
				if r.command == constants.CommandBackup {
					// Retention
					if r.profile.Retention != nil && r.profile.Retention.AfterBackup {
						err = r.runRetention()
						if err != nil {
							return err
						}
					}
					// Check
					if r.profile.Backup != nil && r.profile.Backup.CheckAfter {
						err = r.runCheck()
						if err != nil {
							return err
						}
					}
					// Shell commands
					err = r.runPostCommand(r.command)
					if err != nil {
						return err
					}
				}

				// post-profile commands
				err = r.runProfilePostCommand()
				if err != nil {
					return err
				}

				return nil
			},
			// on failure
			func(err error) {
				_ = r.runProfilePostFailCommand(err)
			},
		)
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *resticWrapper) prepareCommand(command string, args []string) shellCommandDefinition {
	// place the restic command first, there are some flags not recognized otherwise (like --stdin)
	arguments := append([]string{command}, args...)

	if r.moreArgs != nil && len(r.moreArgs) > 0 {
		arguments = append(arguments, r.moreArgs...)
	}

	// Special case for backup command
	if command == constants.CommandBackup {
		arguments = append(arguments, r.profile.GetBackupSource()...)
	}

	env := append(os.Environ(), r.getEnvironment()...)

	clog.Debugf("starting command: %s %s", r.resticBinary, strings.Join(arguments, " "))
	rCommand := newShellCommand(r.resticBinary, arguments, env, r.dryRun, r.sigChan, r.setPID)
	// stdout are stderr are coming from the default terminal (in case they're redirected)
	rCommand.stdout = term.GetOutput()
	rCommand.stderr = term.GetErrorOutput()

	if command == constants.CommandBackup && r.profile.Backup != nil && r.profile.Backup.UseStdin {
		clog.Debug("redirecting stdin to the backup")
		rCommand.useStdin = true
	}
	return rCommand
}

func (r *resticWrapper) runInitialize() error {
	clog.Infof("profile '%s': initializing repository (if not existing)", r.profile.Name)
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandInit))
	rCommand := r.prepareCommand(constants.CommandInit, args)
	// don't display any error
	rCommand.stderr = nil
	_, stderr, err := runShellCommand(rCommand)
	if err != nil {
		return newCommandError(rCommand, stderr, fmt.Errorf("repository initialization on profile '%s': %w", r.profile.Name, err))
	}
	return nil
}

func (r *resticWrapper) runCheck() error {
	clog.Infof("profile '%s': checking repository consistency", r.profile.Name)
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandCheck))
	for {
		rCommand := r.prepareCommand(constants.CommandCheck, args)
		output := captureCommandOutput(rCommand, 96*1024)
		summary, stderr, err := runShellCommand(rCommand)
		if err != nil {
			if r.canRetryAfterError(rCommand, output, err) {
				continue
			}
			r.statusError(constants.CommandCheck, summary, stderr, err)
			return newCommandError(rCommand, stderr, fmt.Errorf("backup check on profile '%s': %w", r.profile.Name, err))
		}
		r.statusSuccess(constants.CommandCheck, summary, stderr)
		return nil
	}
}

func (r *resticWrapper) runRetention() error {
	clog.Infof("profile '%s': cleaning up repository using retention information", r.profile.Name)
	args := convertIntoArgs(r.profile.GetRetentionFlags())
	for {
		rCommand := r.prepareCommand(constants.CommandForget, args)
		output := captureCommandOutput(rCommand, 96*1024)
		summary, stderr, err := runShellCommand(rCommand)
		if err != nil {
			if r.canRetryAfterError(rCommand, output, err) {
				continue
			}
			r.statusError(constants.SectionConfigurationRetention, summary, stderr, err)
			return newCommandError(rCommand, stderr, fmt.Errorf("backup retention on profile '%s': %w", r.profile.Name, err))
		}
		r.statusSuccess(constants.SectionConfigurationRetention, summary, stderr)
		return nil
	}
}

func (r *resticWrapper) runCommand(command string) error {
	clog.Infof("profile '%s': starting '%s'", r.profile.Name, command)
	args := convertIntoArgs(r.profile.GetCommandFlags(command))
	for {
		rCommand := r.prepareCommand(command, args)
		output := captureCommandOutput(rCommand, 96*1024)
		err := runShellCommand(rCommand)
		if err != nil {
			if r.canRetryAfterError(rCommand, output, err) {
				continue
			}
			r.statusError(r.command, err)
			return newCommandError(rCommand, fmt.Errorf("%s on profile '%s': %w", r.command, r.profile.Name, err))
		}
		r.statusSuccess(r.command)
		clog.Infof("profile '%s': finished '%s'", r.profile.Name, command)
		return nil
	}
	rCommand := r.prepareCommand(command, args)
	if command == constants.CommandBackup && r.profile.StatusFile != "" {
		if r.profile.Backup != nil && r.profile.Backup.ExtendedStatus {
			rCommand.scanOutput = shell.ScanBackupJson
			// } else {
			// scan plain backup could have been a good idea,
			// except restic detects its output is not a terminal and no longer displays the progress
			// rCommand.scanOutput = shell.ScanBackupPlain
		}
	}
	summary, stderr, err := runShellCommand(rCommand)
	if err != nil {
		if command == constants.CommandBackup && r.profile.Backup != nil && r.profile.Backup.NoErrorOnWarning {
			// ignore restic warnings after a backup
			exitErr := &exec.ExitError{}
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 3 {
				// this is a restic warning only
				r.statusSuccess(r.command, summary, stderr)
				clog.Warningf("profile '%s': finished '%s' with warning: failed to read all source data during backup", r.profile.Name, command)
				return nil
			}
		}
		r.statusError(r.command, summary, stderr, err)
		return newCommandError(rCommand, stderr, fmt.Errorf("%s on profile '%s': %w", r.command, r.profile.Name, err))
	}
	r.statusSuccess(r.command, summary, stderr)
	clog.Infof("profile '%s': finished '%s'", r.profile.Name, command)
	return nil
}

func (r *resticWrapper) runUnlock() error {
	clog.Infof("profile '%s': unlock stale locks", r.profile.Name)
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandUnlock))
	rCommand := r.prepareCommand(constants.CommandUnlock, args)
	err := runShellCommand(rCommand)
	if err != nil {
		r.statusError(constants.CommandUnlock, err)
		return newCommandError(rCommand, fmt.Errorf("unlock on profile '%s': %w", r.profile.Name, err))
	}
	r.statusSuccess(constants.CommandUnlock)
	return nil
}

func (r *resticWrapper) runPreCommand(command string) error {
	// Pre/Post commands are only supported for backup
	if command != constants.CommandBackup {
		return nil
	}
	if r.profile.Backup == nil || r.profile.Backup.RunBefore == nil || len(r.profile.Backup.RunBefore) == 0 {
		return nil
	}
	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)

	for i, preCommand := range r.profile.Backup.RunBefore {
		clog.Debugf("starting pre-backup command %d/%d", i+1, len(r.profile.Backup.RunBefore))
		rCommand := newShellCommand(preCommand, nil, env, r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		_, stderr, err := runShellCommand(rCommand)
		if err != nil {
			return newCommandError(rCommand, stderr, fmt.Errorf("run-before backup on profile '%s': %w", r.profile.Name, err))
		}
	}
	return nil
}

func (r *resticWrapper) runPostCommand(command string) error {
	// Pre/Post commands are only supported for backup
	if command != constants.CommandBackup {
		return nil
	}
	if r.profile.Backup == nil || r.profile.Backup.RunAfter == nil || len(r.profile.Backup.RunAfter) == 0 {
		return nil
	}
	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)

	for i, postCommand := range r.profile.Backup.RunAfter {
		clog.Debugf("starting post-backup command %d/%d", i+1, len(r.profile.Backup.RunAfter))
		rCommand := newShellCommand(postCommand, nil, env, r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		_, stderr, err := runShellCommand(rCommand)
		if err != nil {
			return newCommandError(rCommand, stderr, fmt.Errorf("run-after backup on profile '%s': %w", r.profile.Name, err))
		}
	}
	return nil
}

func (r *resticWrapper) runProfilePreCommand() error {
	if r.profile.RunBefore == nil || len(r.profile.RunBefore) == 0 {
		return nil
	}
	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)

	for i, preCommand := range r.profile.RunBefore {
		clog.Debugf("starting 'run-before' profile command %d/%d", i+1, len(r.profile.RunBefore))
		rCommand := newShellCommand(preCommand, nil, env, r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		_, stderr, err := runShellCommand(rCommand)
		if err != nil {
			return newCommandError(rCommand, stderr, fmt.Errorf("run-before on profile '%s': %w", r.profile.Name, err))
		}
	}
	return nil
}

func (r *resticWrapper) runProfilePostCommand() error {
	if r.profile.RunAfter == nil || len(r.profile.RunAfter) == 0 {
		return nil
	}
	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)

	for i, postCommand := range r.profile.RunAfter {
		clog.Debugf("starting 'run-after' profile command %d/%d", i+1, len(r.profile.RunAfter))
		rCommand := newShellCommand(postCommand, nil, env, r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		_, stderr, err := runShellCommand(rCommand)
		if err != nil {
			return newCommandError(rCommand, stderr, fmt.Errorf("run-after on profile '%s': %w", r.profile.Name, err))
		}
	}
	return nil
}

func (r *resticWrapper) runProfilePostFailCommand(fail error) error {
	if r.profile.RunAfterFail == nil || len(r.profile.RunAfterFail) == 0 {
		return nil
	}
	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)
	env = append(env, fmt.Sprintf("ERROR=%s", fail.Error()))

	if fail, ok := fail.(*commandError); ok {
		env = append(env,
			fmt.Sprintf("ERROR_COMMANDLINE=%s", fail.Commandline()),
			fmt.Sprintf("RESTIC_STDERR=%s", fail.Stderr()),
		)
	}

	for i, postCommand := range r.profile.RunAfterFail {
		clog.Debugf("starting 'run-after-fail' profile command %d/%d", i+1, len(r.profile.RunAfterFail))
		rCommand := newShellCommand(postCommand, nil, env, r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		_, stderr, err := runShellCommand(rCommand)
		if err != nil {
			return newCommandError(rCommand, stderr, err)
		}
	}
	return nil
}

// getEnvironment returns the environment variables defined in the profile configuration
func (r *resticWrapper) getEnvironment() []string {
	if r.profile.Environment == nil || len(r.profile.Environment) == 0 {
		return nil
	}
	env := make([]string, len(r.profile.Environment))
	i := 0
	for key, value := range r.profile.Environment {
		// env variables are always uppercase
		key = strings.ToUpper(key)
		clog.Debugf("setting up environment variable '%s'", key)
		env[i] = fmt.Sprintf("%s=%s", key, value)
		i++
	}
	return env
}

// getProfileEnvironment returns some environment variables about the current profile
// (name and command for now)
func (r *resticWrapper) getProfileEnvironment() []string {
	return []string{
		fmt.Sprintf("PROFILE_NAME=%s", r.profile.Name),
		fmt.Sprintf("PROFILE_COMMAND=%s", r.command),
	}
}

func (r *resticWrapper) statusSuccess(command string, summary shell.Summary, stderr string) {
	if r.profile.StatusFile == "" {
		return
	}
	var err error
	switch command {
	case constants.CommandBackup:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).BackupSuccess(summary, stderr)
		err = status.Save()
	case constants.CommandCheck:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).CheckSuccess(summary, stderr)
		err = status.Save()
	case constants.SectionConfigurationRetention, constants.CommandForget:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).RetentionSuccess(summary, stderr)
		err = status.Save()
	}
	if err != nil {
		// not important enough to throw an error here
		clog.Warningf("saving status file '%s': %v", r.profile.StatusFile, err)
	}
}

func (r *resticWrapper) statusError(command string, summary shell.Summary, stderr string, fail error) {
	if r.profile.StatusFile == "" {
		return
	}
	var err error
	switch command {
	case constants.CommandBackup:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).BackupError(fail, summary, stderr)
		err = status.Save()
	case constants.CommandCheck:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).CheckError(fail, summary, stderr)
		err = status.Save()
	case constants.SectionConfigurationRetention, constants.CommandForget:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).RetentionError(fail, summary, stderr)
		err = status.Save()
	}
	if err != nil {
		// not important enough to throw an error here
		clog.Warningf("saving status file '%s': %v", r.profile.StatusFile, err)
	}
}

var outputPatterns = map[string]*regexp.Regexp{
	"lock-failure":       regexp.MustCompile("unable to create lock.+already locked.+"),
	"lock-failure,age":   regexp.MustCompile("lock was created at.+\\(([^()]+)\\).+ago"),
	"lock-failure,stale": regexp.MustCompile("the\\W+unlock\\W+command can be used to remove stale locks"),
}

type outputAnalysis struct {
	counts  map[string]int
	matches map[string][]string
}

func analyzeOutput(output bytes.Buffer) (a outputAnalysis) {
	a = outputAnalysis{counts: map[string]int{}, matches: map[string][]string{}}

	for {
		line, err := output.ReadString('\n')

		for name, pattern := range outputPatterns {
			match := pattern.FindStringSubmatch(line)
			if match != nil {
				a.matches[name] = match
				a.counts[strings.Split(name, ",")[0]]++
			}
		}

		if err != nil {
			break
		}
	}

	return
}

func (r *resticWrapper) canRetryAfterError(command shellCommandDefinition, output bytes.Buffer, err error) bool {
	if err != nil {
		analysis := analyzeOutput(output)

		if retry := r.canRetryAfterResticLockFailure(analysis); retry {
			return true
		}
	}
	return false
}

func (r *resticWrapper) canRetryAfterResticLockFailure(analysis outputAnalysis) bool {
	// Check if we have enough matched lines indicating a remote lock failure
	if analysis.counts["lock-failure"] < 2 {
		return false
	}

	// Check if the restic lock is stale
	{
		staleLock := false
		if m, ok := analysis.matches["lock-failure,age"]; ok && len(m) > 1 {
			if lockAge, err := time.ParseDuration(m[1]); err == nil {
				requiredAge := r.global.ResticStaleLockAge
				if requiredAge < constants.MinResticStaleLockAge {
					requiredAge = constants.MinResticLockRetryTime
				}

				staleLock = lockAge >= requiredAge
			} else {
				clog.Warningf("Failed parsing restic lock age. Cause %s", err.Error())
			}
		}

		if staleLock {
			if r.doneTryUnlock || !r.profile.ForceLock {
				if !r.profile.ForceLock {
					clog.Info("Possible stale lock detected. Set `force-inactive-lock` to `true` to enable automatic unlocking of stale locks.")
				}
				return false
			}

			r.doneTryUnlock = true

			if err := r.runUnlock(); err != nil {
				clog.Errorf("Failed removing stale lock: %s", err.Error())
				return false
			}
			return true
		}
	}

	// Check if we have time left to wait on a non stale lock
	if r.lockWait != nil && r.global.ResticLockRetryTime > constants.MinResticLockRetryTime {

		retryDelay := r.global.ResticLockRetryTime
		if retryDelay > constants.MaxResticLockRetryTime {
			retryDelay = constants.MaxResticLockRetryTime
		}

		elapsedTime := time.Now().Sub(r.startTime)

		if canWait := elapsedTime+retryDelay <= *r.lockWait; canWait {
			time.Sleep(r.global.ResticLockRetryTime)
			return true
		}
		return false
	}

	return false
}

func convertIntoArgs(flags map[string][]string) []string {
	args := make([]string, 0)

	if len(flags) == 0 {
		return args
	}

	// we make a list of keys first, so we can loop on the map from an ordered list of keys
	keys := make([]string, 0, len(flags))
	for key := range flags {
		keys = append(keys, key)
	}
	// sort the keys in order
	sort.Strings(keys)

	// now we loop from the ordered list of keys
	for _, key := range keys {
		values := flags[key]
		if values == nil {
			continue
		}
		if len(values) == 0 {
			args = append(args, fmt.Sprintf("--%s", key))
			continue
		}
		for _, value := range values {
			args = append(args, fmt.Sprintf("--%s", key))
			if value != "" {
				if strings.Contains(value, " ") {
					// quote the string containing spaces
					value = fmt.Sprintf(`"%s"`, value)
				}
				args = append(args, value)
			}
		}
	}
	return args
}

// lockRun is making sure the function is only run once by putting a lockfile on the disk
func lockRun(lockFile string, force bool, lockWait *time.Duration, run func(setPID lock.SetPID) error) error {
	// No lock
	if lockFile == "" {
		return run(nil)
	}

	// Make sure the path to the lock exists
	dir := filepath.Dir(lockFile)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			clog.Warningf("the profile will run without a lockfile: %v", err)
			return run(nil)
		}
	}

	// Acquire lock
	runLock := lock.NewLock(lockFile)
	success := runLock.TryAcquire()
	start := time.Now()

	for !success {
		who, err := runLock.Who()
		if err != nil {
			return fmt.Errorf("another process left the lockfile unreadable: %s", err)
		}

		// should we try to force our way?
		if force {
			success = runLock.ForceAcquire()

			if lockWait == nil || success {
				clog.Warningf("previous run of the profile started by %s hasn't finished properly", who)
			}
		} else {
			success = runLock.TryAcquire()
		}

		// Retry or return?
		if !success {
			if lockWait == nil {
				return fmt.Errorf("another process is already running this profile: %s", who)

			} else {
				if time.Now().Sub(start) < *lockWait {
					time.Sleep(3 * time.Second)
				} else {
					clog.Warningf("previous run of the profile hasn't finished after %s", *lockWait)
					lockWait = nil
				}
			}
		}
	}

	// Run locked
	defer runLock.Release()
	return run(runLock.SetPID)
}

// runOnFailure will run the onFailure function if an error occurred in the run function
func runOnFailure(run func() error, onFailure func(error)) error {
	err := run()
	if err != nil {
		onFailure(err)
	}
	return err
}

// captureCommandOutput installs an output capture filter in the stdout & stderr streams of the specified command
// the captured output can be read from the returned buffer after the command was executed.
func captureCommandOutput(command shellCommandDefinition, captureLimit int) bytes.Buffer {
	buffer := bytes.Buffer{}
	mutex := sync.Mutex{}

	bufferWriter := func(p []byte) (n int, err error) {
		mutex.Lock()
		defer mutex.Unlock()

		n, err = buffer.Write(p)

		if buffer.Len() > captureLimit {
			data := buffer.Bytes()
			newLen := len(data) / 2
			copy(data, data[newLen:])
			buffer.Truncate(newLen)
		}

		return
	}

	// Capture stdout & stderr
	command.stdout = newMultiWriter(command.stdout.Write, bufferWriter)
	command.stderr = newMultiWriter(command.stderr.Write, bufferWriter)

	return buffer
}

// multiWriter implements io.Writer and redirects to multiple outputs.
type multiWriter struct {
	outputs []func(p []byte) (n int, err error)
}

func newMultiWriter(outputs ...func(p []byte) (n int, err error)) *multiWriter {
	return &multiWriter{outputs: outputs}
}

func (s *multiWriter) Write(p []byte) (n int, err error) {
	for _, target := range s.outputs {
		n, err = target(p)

		if err != nil {
			break
		}

		if n < len(p) {
			p = p[0:n]
		}
	}
	return
}
