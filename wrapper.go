package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/lock"
	"github.com/creativeprojects/resticprofile/prom"
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

func (c *commandError) ExitCode() (int, error) {
	if exitError, ok := asExitError(c.err); ok {
		return exitError.ExitCode(), nil
	} else {
		return 0, errors.New("exit code not available")
	}
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
	executionTime time.Duration
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
		resticBinary:  resticBinary,
		initialize:    initialize,
		dryRun:        dryRun,
		noLock:        false,
		lockWait:      nil,
		profile:       profile,
		global:        config.NewGlobal(),
		command:       command,
		moreArgs:      moreArgs,
		sigChan:       c,
		startTime:     time.Unix(0, 0),
		executionTime: 0,
		doneTryUnlock: false,
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
		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration
		if err != nil {
			if r.canRetryAfterError(constants.CommandCheck, summary, err) {
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
		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration
		if err != nil {
			if r.canRetryAfterError(constants.CommandForget, summary, err) {
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

		if command == constants.CommandBackup && r.profile.Backup != nil && (r.profile.StatusFile != "" || r.profile.PrometheusSaveToFile != "" || r.profile.PrometheusPush != "") {
			if r.profile.Backup.ExtendedStatus {
				rCommand.scanOutput = shell.ScanBackupJson
			} else if !term.OsStdoutIsTerminal() {
				// restic detects its output is not a terminal and no longer displays the progress.
				// Scan plain output only if resticprofile is not run from a terminal (e.g. schedule)
				rCommand.scanOutput = shell.ScanBackupPlain
			}
		}

		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration

		if err != nil && !r.canSucceedAfterError(command, summary, err) {
			if r.canRetryAfterError(command, summary, err) {
				continue
			}

			r.statusError(r.command, summary, stderr, err)
			r.prometheus(r.command, prom.StatusFailed, summary)
			return newCommandError(rCommand, stderr, fmt.Errorf("%s on profile '%s': %w", r.command, r.profile.Name, err))
		}

		r.statusSuccess(r.command, summary, stderr)
		r.prometheus(r.command, prom.StatusSuccess, summary)
		clog.Infof("profile '%s': finished '%s'", r.profile.Name, command)
		return nil
	}
}

func (r *resticWrapper) runUnlock() error {
	clog.Infof("profile '%s': unlock stale locks", r.profile.Name)
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandUnlock))
	rCommand := r.prepareCommand(constants.CommandUnlock, args)
	summary, stderr, err := runShellCommand(rCommand)
	r.executionTime += summary.Duration
	if err != nil {
		r.statusError(constants.CommandUnlock, summary, stderr, err)
		return newCommandError(rCommand, stderr, fmt.Errorf("unlock on profile '%s': %w", r.profile.Name, err))
	}
	r.statusSuccess(constants.CommandUnlock, summary, stderr)
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
		exitCode := -1
		if code, err := fail.ExitCode(); err == nil {
			exitCode = code
		}

		env = append(env,
			fmt.Sprintf("ERROR_COMMANDLINE=%s", fail.Commandline()),
			fmt.Sprintf("ERROR_EXIT_CODE=%d", exitCode),
			fmt.Sprintf("ERROR_STDERR=%s", fail.Stderr()),
			// Deprecated: STDERR can originate from (pre/post)-command which doesn't need to be restic
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

func (r *resticWrapper) prometheus(command string, status prom.Status, summary shell.Summary) {
	if r.profile.PrometheusSaveToFile == "" && r.profile.PrometheusPush == "" || command != constants.CommandBackup {
		return
	}
	p := prom.NewBackup("")
	p.Results("", status, summary)

	if r.profile.PrometheusSaveToFile != "" {
		err := p.SaveTo(r.profile.PrometheusSaveToFile)
		if err != nil {
			// not important enough to throw an error here
			clog.Warningf("saving prometheus file %q: %v", r.profile.PrometheusSaveToFile, err)
		}
	}
	if r.profile.PrometheusPush != "" {
		err := p.Push(r.profile.PrometheusPush, command)
		if err != nil {
			// not important enough to throw an error here
			clog.Warningf("pushing prometheus metrics to %q: %v", r.profile.PrometheusPush, err)
		}
	}
}

// canSucceedAfterError returns true if an error reported by running restic in runCommand can be counted as success
func (r *resticWrapper) canSucceedAfterError(command string, summary shell.Summary, err error) bool {
	if err == nil {
		return true
	}

	// Ignore restic warnings after a backup (if enabled)
	if command == constants.CommandBackup && r.profile.Backup != nil && r.profile.Backup.NoErrorOnWarning {
		if exitErr, ok := asExitError(err); ok && exitErr.ExitCode() == 3 {
			clog.Warningf("profile '%s': finished '%s' with warning: failed to read all source data during backup", r.profile.Name, command)
			return true
		}
	}

	return false
}

// canRetryAfterError returns true if an error reported by running restic in runCommand, runRetention or runCheck can be retried
func (r *resticWrapper) canRetryAfterError(command string, summary shell.Summary, err error) bool {
	if err == nil {
		panic("invalid usage. err is nil.")
	}

	retry := false
	sleep := time.Duration(0)
	output := summary.OutputAnalysis

	if output.ContainsRemoteLockFailure() {
		clog.Debugf("repository lock failed when running '%s'", command)
		retry, sleep = r.canRetryAfterRemoteLockFailure(output)
	}

	if retry && sleep > 0 {
		time.Sleep(sleep)
	}

	return retry
}

func (r *resticWrapper) canRetryAfterRemoteLockFailure(output shell.OutputAnalysis) (bool, time.Duration) {
	if !output.ContainsRemoteLockFailure() {
		return false, 0
	}

	// Check if the remote lock is stale
	{
		staleLock := false
		staleConditionText := ""

		if lockAge, ok := output.GetRemoteLockedSince(); ok {
			requiredAge := r.global.ResticStaleLockAge
			if requiredAge < constants.MinResticStaleLockAge {
				requiredAge = constants.MinResticStaleLockAge
			}

			staleLock = lockAge >= requiredAge
			staleConditionText = fmt.Sprintf("lock age %s >= %s", lockAge, requiredAge)
		}

		if staleLock && r.global.ResticStaleLockAge > 0 {
			staleConditionText = fmt.Sprintf("restic: possible stale lock detected (%s)", staleConditionText)

			// Loop protection for stale unlock attempts
			if r.doneTryUnlock {
				clog.Infof("%s. Unlock already attempted, will not try again.", staleConditionText)
				return false, 0
			} else {
				r.doneTryUnlock = true
			}

			if !r.profile.ForceLock {
				clog.Infof("%s. Set `force-inactive-lock` to `true` to enable automatic unlocking of stale locks.", staleConditionText)
				return false, 0
			}

			clog.Infof("%s. Trying to unlock.", staleConditionText)
			if err := r.runUnlock(); err != nil {
				clog.Errorf("failed removing stale lock. Cause: %s", err.Error())
				return false, 0
			}
			return true, 0
		}
	}

	// Check if we have time left to wait on a non-stale lock
	retryDelay := r.global.ResticLockRetryAfter

	if r.lockWait != nil && retryDelay > 0 {
		elapsedTime := time.Since(r.startTime)
		availableTime := *r.lockWait - elapsedTime + r.executionTime

		if retryDelay < constants.MinResticLockRetryTime {
			retryDelay = constants.MinResticLockRetryTime
		} else if retryDelay > constants.MaxResticLockRetryTime {
			retryDelay = constants.MaxResticLockRetryTime
		}

		if retryDelay > availableTime {
			retryDelay = availableTime
		}

		if retryDelay >= constants.MinResticLockRetryTime {
			lockName := r.profile.Repository
			if lockedBy, ok := output.GetRemoteLockedBy(); ok {
				lockName = fmt.Sprintf("%s locked by %s", lockName, lockedBy)
			}
			logLockWait(lockName, r.startTime, time.Unix(0, 0), *r.lockWait)

			return true, retryDelay
		}
		return false, 0
	}

	return false, 0
}

// convertIntoArgs converts a map of arguments into a list of arguments to send on the command line
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
				args = append(args, quoteArgument(value))
			}
		}
	}
	return args
}

func quoteArgument(value string) string {
	if strings.Contains(value, " ") {
		// quote the string containing spaces
		value = fmt.Sprintf(`"%s"`, value)
	}
	return value
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
	locker := ""
	lockWaitLogged := time.Unix(0, 0)

	for !success {
		if who, err := runLock.Who(); err == nil {
			if locker != who {
				lockWaitLogged = time.Unix(0, 0)
			}
			locker = who
		} else if errors.Is(err, fs.ErrNotExist) {
			locker = "none"
		} else {
			return fmt.Errorf("another process left the lockfile unreadable: %s", err)
		}

		// should we try to force our way?
		if force {
			success = runLock.ForceAcquire()

			if lockWait == nil || success {
				clog.Warningf("previous run of the profile started by %s hasn't finished properly", locker)
			}
		} else {
			success = runLock.TryAcquire()
		}

		// Retry or return?
		if !success {
			if lockWait == nil {
				return fmt.Errorf("another process is already running this profile: %s", locker)

			} else {
				if time.Since(start) < *lockWait {
					lockName := fmt.Sprintf("%s locked by %s", lockFile, locker)
					lockWaitLogged = logLockWait(lockName, start, lockWaitLogged, *lockWait)

					sleep := 3 * time.Second
					if sleep > *lockWait {
						sleep = *lockWait
					}
					time.Sleep(sleep)
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

const logLockWaitEvery = 5 * time.Minute

func logLockWait(lockName string, started, lastLogged time.Time, maxLockWait time.Duration) time.Time {
	now := time.Now()
	lastLog := now.Sub(lastLogged)
	elapsed := now.Sub(started).Truncate(time.Second)
	remaining := (maxLockWait - elapsed).Truncate(time.Second)

	if lastLog > logLockWaitEvery {
		if elapsed > logLockWaitEvery {
			clog.Infof("lock wait (remaining %s / elapsed %s): %s", remaining, elapsed, strings.TrimSpace(lockName))
		} else {
			clog.Infof("lock wait (remaining %s): %s", remaining, strings.TrimSpace(lockName))
		}
		return now
	}

	return lastLogged
}

// runOnFailure will run the onFailure function if an error occurred in the run function
func runOnFailure(run func() error, onFailure func(error)) error {
	err := run()
	if err != nil {
		onFailure(err)
	}
	return err
}

func asExitError(err error) (*exec.ExitError, bool) {
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		return exitErr, true
	}
	return nil, false
}
