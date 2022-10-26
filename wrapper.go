package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/lock"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/monitor/hook"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/term"
)

type resticWrapper struct {
	resticBinary string
	dryRun       bool
	noLock       bool
	lockWait     *time.Duration
	profile      *config.Profile
	global       *config.Global
	command      string
	moreArgs     []string
	sigChan      chan os.Signal
	setPID       func(pid int)
	stdin        io.ReadCloser
	progress     []monitor.Receiver
	sender       *hook.Sender

	// States
	startTime     time.Time
	executionTime time.Duration
	doneTryUnlock bool
}

func newResticWrapper(
	global *config.Global,
	resticBinary string,
	dryRun bool,
	profile *config.Profile,
	command string,
	moreArgs []string,
	c chan os.Signal,
) *resticWrapper {
	if global == nil {
		global = config.NewGlobal()
	}
	return &resticWrapper{
		resticBinary:  resticBinary,
		dryRun:        dryRun,
		noLock:        false,
		lockWait:      nil,
		profile:       profile,
		global:        global,
		command:       command,
		moreArgs:      moreArgs,
		sigChan:       c,
		stdin:         os.Stdin,
		progress:      make([]monitor.Receiver, 0),
		sender:        hook.NewSender(global.CACertificates, "resticprofile/"+version, global.SenderTimeout),
		startTime:     time.Unix(0, 0),
		executionTime: 0,
		doneTryUnlock: false,
	}
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

// addProgress instance to report back
func (r *resticWrapper) addProgress(p monitor.Receiver) {
	r.progress = append(r.progress, p)
}

func (r *resticWrapper) start(command string) {
	if r.dryRun {
		return
	}
	for _, p := range r.progress {
		p.Start(command)
	}
}

func (r *resticWrapper) summary(command string, summary monitor.Summary, stderr string, result error) {
	if r.dryRun {
		return
	}
	for _, p := range r.progress {
		p.Summary(command, summary, stderr, result)
	}
}

func (r *resticWrapper) runnerWithBeforeAndAfter(commands config.RunShellCommandsSection, command string, action func() error) func() error {
	return func() (err error) {
		err = r.runBeforeCommands(commands, command)

		if err == nil {
			err = action()
		}

		if err == nil {
			err = r.runAfterCommands(commands, command)
		}
		return
	}
}

func (r *resticWrapper) getCommandAction(command string) func() error {
	return func() error { return r.runCommand(command) }
}

func (r *resticWrapper) getCopyAction() func() error {
	copyAction := r.getCommandAction(constants.CommandCopy)

	return func() error {
		// we might need to initialize the secondary repository (the copy target)
		if r.global.Initialize || (r.profile.Copy != nil && r.profile.Copy.Initialize) {
			_ = r.runInitializeCopy() // it's ok if the initialization returned an error
		}

		return copyAction()
	}
}

func (r *resticWrapper) getBackupAction() func() error {
	backupAction := r.getCommandAction(constants.CommandBackup)

	return func() (err error) {
		// Check before
		if err == nil && r.profile.Backup != nil && r.profile.Backup.CheckBefore {
			err = r.runCheck()
		}

		// Retention before
		if err == nil && r.profile.Retention != nil && r.profile.Retention.BeforeBackup {
			err = r.runRetention()
		}

		// Backup command
		if err == nil {
			err = backupAction()
		}

		// Retention after
		if err == nil && r.profile.Retention != nil && r.profile.Retention.AfterBackup {
			err = r.runRetention()
		}

		// Check after
		if err == nil && r.profile.Backup != nil && r.profile.Backup.CheckAfter {
			err = r.runCheck()
		}

		return
	}
}

func (r *resticWrapper) runProfile() error {
	lockFile := r.profile.Lock
	if r.noLock || r.dryRun {
		lockFile = ""
	}

	r.startTime = time.Now()
	profileShellCommands, shellCommands := r.profile.GetRunShellCommandsSections(r.command)

	err := lockRun(lockFile, r.profile.ForceLock, r.lockWait, func(setPID lock.SetPID) error {
		r.setPID = setPID
		return runOnFailure(
			r.runnerWithBeforeAndAfter(profileShellCommands, "", func() (err error) {
				// breaking change from 0.7.0 and 0.7.1:
				// run the initialization after the pre-profile commands
				if (r.global.Initialize || r.profile.Initialize) && r.command != constants.CommandInit {
					_ = r.runInitialize()
					// it's ok for the initialize to error out when the repository exists
				}

				r.sendBefore(r.command)

				// Main command
				{
					var runner func() error
					switch r.command {
					case constants.CommandCopy:
						runner = r.getCopyAction()
					case constants.CommandBackup:
						runner = r.getBackupAction()
					default:
						runner = r.getCommandAction(r.command)
					}

					// Wrap command action in "run-before" & "run-after" from section
					runner = r.runnerWithBeforeAndAfter(shellCommands, r.command, runner)

					// Execute command sequence
					err = runner()
				}

				if err == nil {
					r.sendAfter(r.command)
				}
				return
			}),
			// on failure
			func(err error) {
				r.sendAfterFail(r.command, err)
				// "run-after-fail" in section (returns nil when no-error or not defined)
				if r.runAfterFailCommands(shellCommands, err, r.command) == nil {
					// "run-after-fail" in profile
					_ = r.runAfterFailCommands(profileShellCommands, err, "")
				}
			},
			// finally
			func(err error) {
				r.runFinalShellCommands(r.command, err)
				r.sendFinally(r.command, err)
			},
		)
	})
	if err != nil {
		return err
	}
	return nil
}

var commonResticArgsList = []string{
	"--cacert",
	"--cache-dir",
	"--cleanup-cache",
	"-h", "--help",
	"--insecure-tls",
	"--json",
	"--key-hint",
	"--limit-download",
	"--limit-upload",
	"--no-cache",
	"-o", "--option",
	"--password-command",
	"-p", "--password-file",
	"-q", "--quiet",
	"-r", "--repo",
	"--repository-file",
	"--tls-client-cert",
	"-v", "--verbose",
}

// commonResticArgs turns args into commonArgs containing only those args that all restic commands understand
func (r *resticWrapper) commonResticArgs(args []string) (commonArgs []string) {
	if !sort.StringsAreSorted(commonResticArgsList) {
		sort.Strings(commonResticArgsList)
	}
	skipValue := true

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			lookup := strings.TrimSpace(strings.Split(arg, "=")[0])
			index := sort.SearchStrings(commonResticArgsList, lookup)
			if index < len(commonResticArgsList) && commonResticArgsList[index] == lookup {
				commonArgs = append(commonArgs, arg)
				skipValue = strings.Contains(arg, "=")
				continue
			}
		} else if !skipValue {
			commonArgs = append(commonArgs, arg)
		}
		skipValue = true
	}
	return
}

func (r *resticWrapper) getShell() (shell []string) {
	if r.global != nil {
		shell = r.global.ShellBinary
	}
	return
}

func (r *resticWrapper) prepareCommand(command string, args *shell.Args, moreArgs ...string) shellCommandDefinition {
	// Create local instance to allow modification
	args = args.Clone()

	if len(moreArgs) > 0 {
		args.AddArgs(moreArgs, shell.ArgCommandLineEscape)
	}

	// Special case for backup command
	if command == constants.CommandBackup {
		args.AddArgs(r.profile.GetBackupSource(), shell.ArgConfigBackupSource)
	}

	// place the restic command first, there are some flags not recognized otherwise (like --stdin)
	arguments := append([]string{command}, args.GetAll()...)

	// Create non-confidential arguments list for logging
	publicArguments := append([]string{command}, config.GetNonConfidentialArgs(r.profile, args).GetAll()...)

	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)

	clog.Debugf("starting command: %s %s", r.resticBinary, strings.Join(publicArguments, " "))
	rCommand := newShellCommand(r.resticBinary, arguments, env, r.getShell(), r.dryRun, r.sigChan, r.setPID)
	rCommand.publicArgs = publicArguments
	// stdout are stderr are coming from the default terminal (in case they're redirected)
	rCommand.stdout = term.GetOutput()
	rCommand.stderr = term.GetErrorOutput()
	rCommand.streamError = r.profile.StreamError

	return rCommand
}

// runInitialize tries to initialize the repository
func (r *resticWrapper) runInitialize() error {
	clog.Infof("profile '%s': initializing repository (if not existing)", r.profile.Name)
	args := r.profile.GetCommandFlags(constants.CommandInit)
	rCommand := r.prepareCommand(constants.CommandInit, args, r.commonResticArgs(r.moreArgs)...)
	// don't display any error
	rCommand.stderr = nil
	_, stderr, err := runShellCommand(rCommand)
	if err != nil {
		return newCommandError(rCommand, stderr, fmt.Errorf("repository initialization on profile '%s': %w", r.profile.Name, err))
	}
	return nil
}

// runInitializeCopy tries to initialize the secondary repository used by the copy command
func (r *resticWrapper) runInitializeCopy() error {
	clog.Infof("profile '%s': initializing secondary repository (if not existing)", r.profile.Name)
	args := r.profile.GetCommandFlags(constants.CommandCopy)
	swap := false
	if r.profile.Copy != nil && r.profile.Copy.InitializeCopyChunkerParams {
		swap = true
		// this a bit hacky, but we need to add this flag manually since it's coming from
		// the configuration of the "copy" section, but cannot be a flag of the copy section
		args.AddFlag(constants.ParameterCopyChunkerParams, "", shell.ArgConfigEscape)
	}
	// the copy command adds a "2" behind each flag about the secondary repository
	// in the case of init, we want to promote the secondary repository as primary
	// but if we use the copy-chunker-params we actually need to swap primary with secondary
	args.PromoteSecondaryToPrimary(swap)
	rCommand := r.prepareCommand(constants.CommandInit, args, r.commonResticArgs(r.moreArgs)...)
	// don't display any error
	rCommand.stderr = nil
	_, stderr, err := runShellCommand(rCommand)
	if err != nil {
		return newCommandError(rCommand, stderr, fmt.Errorf("copy repository initialization on profile '%s': %w", r.profile.Name, err))
	}
	return nil
}

func (r *resticWrapper) runCheck() error {
	clog.Infof("profile '%s': checking repository consistency", r.profile.Name)
	r.start(constants.CommandCheck)
	args := r.profile.GetCommandFlags(constants.CommandCheck)
	for {
		rCommand := r.prepareCommand(constants.CommandCheck, args, r.commonResticArgs(r.moreArgs)...)
		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration
		r.summary(constants.CommandCheck, summary, stderr, err)
		if err != nil {
			if r.canRetryAfterError(constants.CommandCheck, summary, err) {
				continue
			}
			return newCommandError(rCommand, stderr, fmt.Errorf("backup check on profile '%s': %w", r.profile.Name, err))
		}
		return nil
	}
}

func (r *resticWrapper) runRetention() error {
	clog.Infof("profile '%s': cleaning up repository using retention information", r.profile.Name)
	r.start(constants.SectionConfigurationRetention)
	args := r.profile.GetRetentionFlags()
	for {
		rCommand := r.prepareCommand(constants.CommandForget, args, r.commonResticArgs(r.moreArgs)...)
		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration
		r.summary(constants.SectionConfigurationRetention, summary, stderr, err)
		if err != nil {
			if r.canRetryAfterError(constants.CommandForget, summary, err) {
				continue
			}
			return newCommandError(rCommand, stderr, fmt.Errorf("backup retention on profile '%s': %w", r.profile.Name, err))
		}
		return nil
	}
}

func (r *resticWrapper) runCommand(command string) error {
	clog.Infof("profile '%s': starting '%s'", r.profile.Name, command)
	r.start(command)
	args := r.profile.GetCommandFlags(command)

	streamSource := io.NopCloser(strings.NewReader(""))
	defer func() { streamSource.Close() }()

	for {
		if err := streamSource.Close(); err != nil {
			return fmt.Errorf("%s on profile '%s'. Failed closing stream source: %w", r.command, r.profile.Name, err)
		}

		rCommand := r.prepareCommand(command, args, r.moreArgs...)

		if command == constants.CommandBackup && r.profile.Backup != nil {
			// Add output scanners
			if len(r.progress) > 0 {
				if r.profile.Backup.ExtendedStatus {
					rCommand.scanOutput = shell.ScanBackupJson
				} else if !term.OsStdoutIsTerminal() {
					// restic detects its output is not a terminal and no longer displays the monitor.
					// Scan plain output only if resticprofile is not run from a terminal (e.g. schedule)
					rCommand.scanOutput = shell.ScanBackupPlain
				}
			}

			// Redirect a stream source to stdin of restic if configured
			if source, err := r.prepareStreamSource(); err == nil {
				if source != nil {
					streamSource = source
					rCommand.stdin = streamSource
				}
			} else {
				return newCommandError(rCommand, "", fmt.Errorf("%s on profile '%s': %w", r.command, r.profile.Name, err))
			}
		}

		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration
		r.summary(r.command, summary, stderr, err)

		if err != nil && !r.canSucceedAfterError(command, summary, err) {
			if r.canRetryAfterError(command, summary, err) {
				continue
			}
			return newCommandError(rCommand, stderr, fmt.Errorf("%s on profile '%s': %w", r.command, r.profile.Name, err))
		}
		clog.Infof("profile '%s': finished '%s'", r.profile.Name, command)
		return nil
	}
}

func (r *resticWrapper) runUnlock() error {
	clog.Infof("profile '%s': unlock stale locks", r.profile.Name)
	r.start(constants.CommandUnlock)
	args := r.profile.GetCommandFlags(constants.CommandUnlock)
	rCommand := r.prepareCommand(constants.CommandUnlock, args, r.commonResticArgs(r.moreArgs)...)
	summary, stderr, err := runShellCommand(rCommand)
	r.executionTime += summary.Duration
	r.summary(constants.CommandUnlock, summary, stderr, err)
	if err != nil {
		return newCommandError(rCommand, stderr, fmt.Errorf("unlock on profile '%s': %w", r.profile.Name, err))
	}
	return nil
}

// runBeforeCommands runs the "run-before" commands (use empty command for profile commands)
func (r *resticWrapper) runBeforeCommands(commands config.RunShellCommandsSection, command string) error {
	return r.runShellCommands(commands.RunBefore, "run-before", command, nil)
}

// runAfterCommands runs the "run-after" commands (use empty command for profile commands)
func (r *resticWrapper) runAfterCommands(commands config.RunShellCommandsSection, command string) error {
	return r.runShellCommands(commands.RunAfter, "run-after", command, nil)
}

// runAfterFailCommands runs the "run-after-fail" commands (use empty command for profile commands)
func (r *resticWrapper) runAfterFailCommands(commands config.RunShellCommandsSection, failure error, command string) error {
	return r.runShellCommands(commands.RunAfterFail, "run-after-fail", command, failure)
}

// runShellCommands runs a set of shell commands and stops at the first error (if any).
// commandsType and command is used for logging and in error messages but has no other influence.
// set failure to a non-nil value to initialize a fail environment (e.g. run-after-fail).
func (r *resticWrapper) runShellCommands(commands []string, commandsType, command string, failure error) error {
	if len(command) > 0 {
		commandsType = commandsType + " " + command
	}

	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)
	env = append(env, r.getFailEnvironment(failure)...)

	for i, shellCommand := range commands {
		clog.Debugf("starting %s on profile %d/%d", commandsType, i+1, len(commands))
		rCommand := newShellCommand(shellCommand, nil, env, r.getShell(), r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		_, stderr, err := runShellCommand(rCommand)
		if err != nil {
			err = fmt.Errorf("%s on profile '%s': %w", commandsType, r.profile.Name, err)
			return newCommandError(rCommand, stderr, err)
		}
	}
	return nil
}

// runFinalShellCommands runs all shell commands defined in "run-finally".
func (r *resticWrapper) runFinalShellCommands(command string, fail error) {
	var commands []string

	profileCommands, sectionCommands := r.profile.GetRunShellCommandsSections(command)
	commands = append(commands, sectionCommands.RunFinally...)
	commands = append(commands, profileCommands.RunFinally...)

	env := append(os.Environ(), r.getEnvironment()...)
	env = append(env, r.getProfileEnvironment()...)
	env = append(env, r.getFailEnvironment(fail)...)

	for i := len(commands) - 1; i >= 0; i-- {
		// Using defer stack for "finally" to ensure every command is run even on panic
		defer func(index int, cmd string) {
			clog.Debugf("starting final command %d/%d", index+1, len(commands))
			rCommand := newShellCommand(cmd, nil, env, r.getShell(), r.dryRun, r.sigChan, r.setPID)
			// stdout are stderr are coming from the default terminal (in case they're redirected)
			rCommand.stdout = term.GetOutput()
			rCommand.stderr = term.GetErrorOutput()
			_, _, err := runShellCommand(rCommand)
			if err != nil {
				clog.Errorf("run-finally command %d/%d failed ('%s' on profile '%s'): %w",
					index+1, len(commands), command, r.profile.Name, err)
			}
		}(i, commands[i])
	}
}

// sendBefore a command
func (r *resticWrapper) sendBefore(command string) {
	monitoringSections := r.profile.GetMonitoringSections(command)
	if monitoringSections == nil {
		return
	}

	for i, send := range monitoringSections.SendBefore {
		clog.Debugf("starting 'send-before' from %s %d/%d", command, i+1, len(monitoringSections.SendBefore))
		err := r.sender.Send(send, r.getContext())
		if err != nil {
			clog.Warningf("'send-before' returned an error: %s", err)
		}
	}
}

// sendAfter a command
func (r *resticWrapper) sendAfter(command string) {
	monitoringSections := r.profile.GetMonitoringSections(command)
	if monitoringSections == nil {
		return
	}

	for i, send := range monitoringSections.SendAfter {
		clog.Debugf("starting 'send-after' from %s %d/%d", command, i+1, len(monitoringSections.SendAfter))
		err := r.sender.Send(send, r.getContext())
		if err != nil {
			clog.Warningf("'send-after' returned an error: %s", err)
		}
	}
}

// sendAfterFail a command
func (r *resticWrapper) sendAfterFail(command string, err error) {
	monitoringSections := r.profile.GetMonitoringSections(command)
	if monitoringSections == nil {
		return
	}

	for i, send := range monitoringSections.SendAfterFail {
		clog.Debugf("starting 'send-after-fail' from %s %d/%d", command, i+1, len(monitoringSections.SendAfterFail))
		err := r.sender.Send(send, r.getContextWithError(err))
		if err != nil {
			clog.Warningf("'send-after-fail' returned an error: %s", err)
		}
	}
}

// sendFinally sends all final hooks
func (r *resticWrapper) sendFinally(command string, err error) {
	monitoringSections := r.profile.GetMonitoringSections(command)
	if monitoringSections == nil {
		return
	}

	for i, send := range monitoringSections.SendFinally {
		clog.Debugf("starting 'send-finally' from %s %d/%d", command, i+1, len(monitoringSections.SendFinally))
		err := r.sender.Send(send, r.getContextWithError(err))
		if err != nil {
			clog.Warningf("'send-finally' returned an error: %s", err)
		}
	}
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
		env[i] = fmt.Sprintf("%s=%s", key, value.Value())
		i++
	}
	return env
}

// getProfileEnvironment returns some environment variables about the current profile
// (name and command for now)
func (r *resticWrapper) getProfileEnvironment() []string {
	ctx := r.getContext()
	return []string{
		fmt.Sprintf("%s=%s", constants.EnvProfileName, ctx.ProfileName),
		fmt.Sprintf("%s=%s", constants.EnvProfileCommand, ctx.ProfileCommand),
	}
}

// getFailEnvironment returns additional environment variables describing the failure
func (r *resticWrapper) getFailEnvironment(err error) (env []string) {
	ctx := r.getErrorContext(err)
	if ctx.Message != "" {
		env = append(env, fmt.Sprintf("%s=%s", constants.EnvError, ctx.Message)) // powershell already has $ERROR
		env = append(env, fmt.Sprintf("%s=%s", constants.EnvErrorMessage, ctx.Message))
	}
	if ctx.CommandLine != "" {
		env = append(env, fmt.Sprintf("%s=%s", constants.EnvErrorCommandLine, ctx.CommandLine))
	}
	if ctx.ExitCode != "" {
		env = append(env, fmt.Sprintf("%s=%s", constants.EnvErrorExitCode, ctx.ExitCode))
	}
	if ctx.Stderr != "" {
		env = append(env, fmt.Sprintf("%s=%s", constants.EnvErrorStderr, ctx.Stderr))
		// Deprecated: STDERR can originate from (pre/post)-command which doesn't need to be restic
		env = append(env, fmt.Sprintf("RESTIC_STDERR=%s", ctx.Stderr))
	}
	return
}

func (r *resticWrapper) getContext() hook.Context {
	return hook.Context{
		ProfileName:    r.profile.Name,
		ProfileCommand: r.command,
	}
}

func (r *resticWrapper) getContextWithError(err error) hook.Context {
	ctx := r.getContext()
	ctx.Error = r.getErrorContext(err)
	return ctx
}

func (r *resticWrapper) getErrorContext(err error) hook.ErrorContext {
	ctx := hook.ErrorContext{}
	if err == nil {
		return ctx
	}
	ctx.Message = err.Error()

	if fail, ok := err.(*commandError); ok {
		exitCode := -1
		if code, err := fail.ExitCode(); err == nil {
			exitCode = code
		}

		ctx.CommandLine = fail.Commandline()
		ctx.ExitCode = strconv.Itoa(exitCode)
		ctx.Stderr = fail.Stderr()
	}
	return ctx
}

// canSucceedAfterError returns true if an error reported by running restic in runCommand can be counted as success
func (r *resticWrapper) canSucceedAfterError(command string, summary monitor.Summary, err error) bool {
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
func (r *resticWrapper) canRetryAfterError(command string, summary monitor.Summary, err error) bool {
	if err == nil {
		panic("invalid usage. err is nil.")
	}

	retry := false
	sleep := time.Duration(0)
	output := summary.OutputAnalysis

	if output != nil && output.ContainsRemoteLockFailure() {
		clog.Debugf("repository lock failed when running '%s'", command)
		retry, sleep = r.canRetryAfterRemoteLockFailure(output)
	}

	if retry && sleep > 0 {
		time.Sleep(sleep)
	}

	return retry
}

func (r *resticWrapper) canRetryAfterRemoteLockFailure(output monitor.OutputAnalysis) (bool, time.Duration) {
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
			}
			r.doneTryUnlock = true

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
			lockName := r.profile.Repository.String()
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
			}
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
func runOnFailure(run func() error, onFailure func(error), finally func(error)) (err error) {
	// Using "defer" for finally to ensure it runs even on panic
	if finally != nil {
		defer func() {
			finally(err)
		}()
	}

	err = run()
	if err != nil {
		onFailure(err)
	}

	return
}

func asExitError(err error) (*exec.ExitError, bool) {
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		return exitErr, true
	}
	return nil, false
}
