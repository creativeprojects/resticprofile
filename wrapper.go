package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/collect"
	"golang.org/x/exp/maps"
)

type resticWrapper struct {
	ctx      *Context
	dryRun   bool // resticprofile dry-run (not restic dry-run via flags added on the command line)
	noLock   bool
	lockWait *time.Duration
	profile  *config.Profile
	global   *config.Global
	command  string
	moreArgs []string
	sigChan  chan os.Signal
	setPID   func(pid int)
	stdin    io.ReadCloser
	progress []monitor.Receiver
	sender   *hook.Sender

	// States
	startTime     time.Time
	executionTime time.Duration
	doneTryUnlock bool
	previousEnv   string
}

func newResticWrapper(ctx *Context) *resticWrapper {
	if ctx.global == nil {
		ctx.global = config.NewGlobal()
	}

	resticDryRun := slices.ContainsFunc(ctx.request.arguments, collect.In("--dry-run", "-n"))

	return &resticWrapper{
		ctx:      ctx,
		dryRun:   ctx.flags.dryRun,
		noLock:   false,
		lockWait: nil,
		profile:  ctx.profile,
		global:   ctx.global,
		command:  ctx.command,
		moreArgs: ctx.request.arguments,
		sigChan:  ctx.sigChan,
		stdin:    os.Stdin,
		progress: make([]monitor.Receiver, 0),
		sender: hook.NewSender(
			ctx.global.CACertificates,
			"resticprofile/"+version,
			ctx.global.SenderTimeout,
			ctx.flags.dryRun || resticDryRun,
		),
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
		if r.profile.Backup != nil && r.profile.Backup.CheckBefore {
			err = r.runCheck()
			if err != nil {
				return
			}
		}

		// Retention before
		if r.profile.Retention != nil && r.profile.Retention.BeforeBackup.IsTrue() {
			err = r.runRetention()
			if err != nil {
				return
			}
		}

		// Backup command
		err = backupAction()
		if err != nil {
			return
		}

		// Retention after
		if r.profile.Retention != nil && r.profile.Retention.AfterBackup.IsTrue() {
			err = r.runRetention()
			if err != nil {
				return
			}
		}

		// Check after
		if r.profile.Backup != nil && r.profile.Backup.CheckAfter {
			err = r.runCheck()
			if err != nil {
				return
			}
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
	sendMonitoring := r.profile.GetMonitoringSections(r.command)

	err := lockRun(lockFile, r.profile.ForceLock, r.lockWait, r.sigChan, func(setPID lock.SetPID) error {
		r.setPID = setPID
		return runOnFailure(
			r.runnerWithBeforeAndAfter(profileShellCommands, "", func() (err error) {
				// breaking change from 0.7.0 and 0.7.1:
				// run the initialization after the pre-profile commands
				if (r.global.Initialize || r.profile.Initialize) && r.command != constants.CommandInit {
					_ = r.runInitialize()
					// it's ok for the initialize to error out when the repository exists
				}

				r.sendBefore(sendMonitoring, r.command)

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
					r.sendAfter(sendMonitoring, r.command)
				}
				return
			}),
			// on failure
			func(err error) {
				r.sendAfterFail(sendMonitoring, r.command, err)
				// "run-after-fail" in section (returns nil when no-error or not defined)
				if r.runAfterFailCommands(shellCommands, err, r.command) == nil {
					// "run-after-fail" in profile
					_ = r.runAfterFailCommands(profileShellCommands, err, "")
				}
			},
			// finally
			func(err error) {
				r.runFinalShellCommands(r.command, err)
				r.sendFinally(sendMonitoring, r.command, err)
			},
		)
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *resticWrapper) getResticVersion() string {
	if r.global != nil {
		return r.global.ResticVersion
	}
	return restic.AnyVersion
}

func (r *resticWrapper) validResticArgumentsList(command string) (allValidArgs []string) {
	var opts [][]restic.Option
	if cmd, found := restic.GetCommandForVersion(command, r.getResticVersion(), true); found {
		opts = append(opts, cmd.GetOptions())
	}
	if o := restic.GetDefaultOptionsForVersion(r.getResticVersion(), true); o != nil {
		opts = append(opts, o)
	}

	for _, options := range opts {
		for _, option := range options {
			if !option.AvailableForOS() {
				continue
			}
			if len(option.Alias) == 1 {
				allValidArgs = append(allValidArgs, fmt.Sprintf("-%s", option.Alias))
			}
			if len(option.Name) > 0 {
				allValidArgs = append(allValidArgs, fmt.Sprintf("--%s", option.Name))
			}
		}
	}
	return
}

type argumentsFilter func(args []string, allowExtraValues bool) (filteredArgs []string)

// validArgumentsFilter returns a filter that removes args not valid for the specified restic command
func (r *resticWrapper) validArgumentsFilter(validArgs []string) argumentsFilter {
	validArgs = slices.Clone(validArgs)
	sort.Strings(validArgs)

	return func(args []string, allowExtraValues bool) (filteredArgs []string) {
		skipValue := !allowExtraValues

		for _, arg := range args {
			if strings.HasPrefix(arg, "-") {
				av := strings.Split(arg, "=")
				includesValue := len(av) > 1

				lookup := strings.TrimSpace(av[0])
				index := sort.SearchStrings(validArgs, lookup)
				if found := index < len(validArgs) && validArgs[index] == lookup; found {
					filteredArgs = append(filteredArgs, arg)
					if !includesValue {
						skipValue = false
					}
				} else if !allowExtraValues { // if "allowExtraValues" => args with values must use "--arg=value"
					skipValue = !includesValue
				}
				continue
			} else if !skipValue {
				filteredArgs = append(filteredArgs, arg)
			}
			skipValue = !allowExtraValues
		}
		return
	}
}

func (r *resticWrapper) getShell() (shell []string) {
	if r.global != nil {
		shell = collect.All(r.global.ShellBinary, collect.Not(collect.In("auto")))
	}
	return
}

// getCommandArgumentsFilter returns a filter to remove unsupported args or nil when the binary
// is not restic (ignoring shim or mock binaries) or filtering was disabled
func (r *resticWrapper) getCommandArgumentsFilter(command string) argumentsFilter {
	binaryIsRestic := strings.EqualFold(
		"restic",
		strings.TrimSuffix(filepath.Base(r.ctx.binary), filepath.Ext(r.ctx.binary)),
	)
	if binaryIsRestic && (r.global == nil || r.global.FilterResticFlags) {
		if validArgs := r.validResticArgumentsList(command); len(validArgs) > 0 {
			return r.validArgumentsFilter(validArgs)
		} else {
			clog.Warningf("failed building valid arguments filter for command %q", command)
		}
	}
	return nil
}

func (r *resticWrapper) containsArguments(arguments []string, subset ...string) (found bool) {
	filter := r.validArgumentsFilter(subset)
	argMatcher := func(arg string) bool { return strings.HasPrefix(arg, "-") }
	found = slices.ContainsFunc(filter(arguments, true), argMatcher)
	return
}

func (r *resticWrapper) prepareCommand(command string, args *shell.Args, allowExtraValues bool) shellCommandDefinition {
	// Create local instance to allow modification
	args = args.Clone()

	filter := r.getCommandArgumentsFilter(command)

	// Add extra commandline arguments
	moreArgs := slices.Clone(r.moreArgs)
	if filter != nil {
		clog.Tracef("unfiltered extra flags: %s", strings.Join(config.GetNonConfidentialValues(r.profile, moreArgs), " "))
		moreArgs = filter(moreArgs, allowExtraValues)
	}
	args.AddArgs(shell.NewArgsSlice(moreArgs, shell.ArgCommandLineEscape))

	// Special case for backup command
	var dir string
	if command == constants.CommandBackup {
		args.AddArgs(shell.NewArgsSlice(r.profile.GetBackupSource(), shell.ArgConfigBackupSource))
		if r.profile.Backup != nil && r.profile.Backup.SourceRelative {
			dir = r.profile.Backup.SourceBase
		}
	}
	// Special case for copy command
	if command == constants.CommandCopy {
		args.AddArgs(shell.NewArgsSlice(r.profile.GetCopySnapshotIDs(), shell.ArgConfigEscape))
	}

	env := r.getEnvironment(true)
	env = append(env, r.getProfileEnvironment()...)

	modifiers := []shell.ArgModifier{
		shell.NewLegacyArgModifier(r.ctx.legacyArgs),
		shell.NewExpandEnvModifier(env),
	}
	for _, modifier := range modifiers {
		args = args.Modify(modifier)
	}

	// Add retry-lock (supported from restic 0.16, depends on filter being enabled)
	if lockRetryTime, enabled := r.remainingLockRetryTime(); enabled && filter != nil {
		// limiting the retry handling in restic, we need to make sure we can retry internally so that unlock is called.
		lockRetryTime = lockRetryTime - r.global.ResticLockRetryAfter - constants.MinResticLockRetryDelay
		if lockRetryTime > constants.MaxResticLockRetryTimeArgument {
			lockRetryTime = constants.MaxResticLockRetryTimeArgument
		}
		lockRetryTime = lockRetryTime.Truncate(time.Minute)

		if lockRetryTime > 0 && !r.containsArguments(args.GetAll(), fmt.Sprintf("--%s", constants.ParameterRetryLock)) {
			args.AddFlag(constants.ParameterRetryLock, shell.NewArg(fmt.Sprintf("%.0fm", lockRetryTime.Minutes()), shell.ArgConfigEscape))
		}
	}

	// Build arguments and publicArguments (for logging)
	arguments, publicArguments := args.GetAll(), config.GetNonConfidentialArgs(r.profile, args).GetAll()
	if filter != nil {
		clog.Tracef("unfiltered command: %s %s", command, strings.Join(publicArguments, " "))
		arguments, publicArguments = filter(arguments, true), filter(publicArguments, true)
	}

	// Add restic command
	arguments = append([]string{command}, arguments...)
	publicArguments = append([]string{command}, publicArguments...)

	clog.Debug(func() string {
		wd := ""
		if dir != "" {
			wd = fmt.Sprintf(" (in %s)", dir)
		}
		return fmt.Sprintf("starting command: %s %s%s", r.ctx.binary, strings.Join(publicArguments, " "), wd)
	})

	rCommand := newShellCommand(r.ctx.binary, arguments, env, r.getShell(), r.dryRun, r.sigChan, r.setPID)
	rCommand.publicArgs = publicArguments
	// stdout are stderr are coming from the default terminal (in case they're redirected)
	rCommand.stdout = term.GetOutput()
	rCommand.stderr = term.GetErrorOutput()
	rCommand.streamError = r.profile.StreamError
	rCommand.dir = dir

	return rCommand
}

// runInitialize tries to initialize the repository
func (r *resticWrapper) runInitialize() error {
	clog.Infof("profile '%s': initializing repository (if not existing)", r.profile.Name)
	args := r.profile.GetCommandFlags(constants.CommandInit)
	rCommand := r.prepareCommand(constants.CommandInit, args, false)
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
	args := r.profile.GetCopyInitializeFlags()
	if args == nil {
		return nil // copy is not configured, do nothing
	}
	rCommand := r.prepareCommand(constants.CommandInit, args, false)
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
		rCommand := r.prepareCommand(constants.CommandCheck, args, false)
		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration
		r.summary(constants.CommandCheck, summary, stderr, err)
		if err != nil {
			retry, interruptedError := r.canRetryAfterError(constants.CommandCheck, summary)
			if retry {
				continue
			}
			if interruptedError != nil {
				// we keep the restic stderr, but we set the final error to interrupted
				err = interruptedError
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
		rCommand := r.prepareCommand(constants.CommandForget, args, false)
		summary, stderr, err := runShellCommand(rCommand)
		r.executionTime += summary.Duration
		r.summary(constants.SectionConfigurationRetention, summary, stderr, err)
		if err != nil {
			retry, interruptedError := r.canRetryAfterError(constants.CommandForget, summary)
			if retry {
				continue
			}
			if interruptedError != nil {
				// we keep the restic stderr, but we set the final error to interrupted
				err = interruptedError
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

		rCommand := r.prepareCommand(command, args, true)

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

		if err != nil && !r.canSucceedAfterError(command, err) {
			retry, interruptedError := r.canRetryAfterError(command, summary)
			if retry {
				continue
			}
			if interruptedError != nil {
				// we keep the restic stderr, but we set the final error to interrupted
				err = interruptedError
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
	rCommand := r.prepareCommand(constants.CommandUnlock, args, false)
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

	for i, shellCommand := range commands {
		clog.Debugf("starting %s on profile %d/%d", commandsType, i+1, len(commands))
		// env might change between runs, creating it for every command
		env := r.getEnvironment(true)
		env = append(env, r.getProfileEnvironment()...)
		env = append(env, r.getFailEnvironment(failure)...)
		// creating command
		rCommand := newShellCommand(shellCommand, nil, env, r.getShell(), r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		term.FlushAllOutput()
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

	for i := len(commands) - 1; i >= 0; i-- {
		// Using defer stack for "finally" to ensure every command is run even on panic
		defer func(index int, cmd string) {
			clog.Debugf("starting final command %d/%d", index+1, len(commands))
			// env might change between runs, creating it for every command
			env := r.getEnvironment(true)
			env = append(env, r.getProfileEnvironment()...)
			env = append(env, r.getFailEnvironment(fail)...)
			// creating command
			rCommand := newShellCommand(cmd, nil, env, r.getShell(), r.dryRun, r.sigChan, r.setPID)
			// stdout are stderr are coming from the default terminal (in case they're redirected)
			rCommand.stdout = term.GetOutput()
			rCommand.stderr = term.GetErrorOutput()
			term.FlushAllOutput()
			_, _, err := runShellCommand(rCommand)
			if err != nil {
				clog.Errorf("run-finally command %d/%d failed ('%s' on profile '%s'): %s",
					index+1, len(commands), command, r.profile.Name, err.Error())
			}
		}(i, commands[i])
	}
}

// sendBefore a command
func (r *resticWrapper) sendBefore(monitoring config.SendMonitoringSections, command string) {
	r.sendMonitoring(monitoring.SendBefore, command, "send-before", nil)
}

// sendAfter a command
func (r *resticWrapper) sendAfter(monitoring config.SendMonitoringSections, command string) {
	r.sendMonitoring(monitoring.SendAfter, command, "send-after", nil)
}

// sendAfterFail a command
func (r *resticWrapper) sendAfterFail(monitoring config.SendMonitoringSections, command string, err error) {
	r.sendMonitoring(monitoring.SendAfterFail, command, "send-after-fail", err)
}

// sendFinally sends all final hooks
func (r *resticWrapper) sendFinally(monitoring config.SendMonitoringSections, command string, err error) {
	r.sendMonitoring(monitoring.SendFinally, command, "send-finally", err)
}

func (r *resticWrapper) sendMonitoring(sections []config.SendMonitoringSection, command, sendType string, err error) {
	for i, section := range sections {
		clog.Debugf("starting %q from %s %d/%d", sendType, command, i+1, len(sections))
		term.FlushAllOutput()
		err := r.sender.Send(section, r.getContextWithError(err))
		if err != nil {
			clog.Warningf("%q returned an error: %s", sendType, err.Error())
		}
	}
}

// getEnvironment returns the environment variables defined in the profile configuration
func (r *resticWrapper) getEnvironment(withOs bool) (values []string) {
	// Note: Variable names match the original case for existing OS variables.
	//       New profile environment variables are all uppercase if not matching any OS variable.
	env := r.profile.GetEnvironment(withOs)
	values = env.Values()

	clog.Debug(func() string {
		appliedEnv := r.stringifyEnvironment(env)
		if r.previousEnv != appliedEnv {
			r.previousEnv = appliedEnv
			return fmt.Sprintf("command environment:\n%s", appliedEnv)
		} else {
			return "command environment: reusing previous"
		}
	})
	return
}

// stringifyEnvironment converts the env into a string that can be used for logging
func (r *resticWrapper) stringifyEnvironment(env *util.Environment) string {
	mapper := collect.KVMapper(collect.CopyMapper[string], config.NewConfidentialValue)
	confidentialEnv := collect.FromMap(env.ValuesAsMap(), mapper)
	config.ProcessConfidentialEnvironment(confidentialEnv)

	names := maps.Keys(confidentialEnv)
	sort.Strings(names)

	out := new(strings.Builder)
	for _, name := range names {
		cev := confidentialEnv[name]
		notSameAsOS := cev.Value() != os.Getenv(name) || cev.Value() == ""
		if notSameAsOS {
			_, _ = fmt.Fprintf(out, "%s=%s\n", name, cev)
		}
	}
	return out.String()
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
func (r *resticWrapper) canSucceedAfterError(command string, err error) bool {
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

// canRetryAfterError returns true if an error reported by running restic in runCommand, runRetention or runCheck can be retried.
// the error is detected from the output analysis.
func (r *resticWrapper) canRetryAfterError(command string, summary monitor.Summary) (bool, error) {
	output := summary.OutputAnalysis
	if output == nil || !output.ContainsRemoteLockFailure() {
		return false, nil
	}

	// Do not count lock-wait time as normal execution time (to calc correct remaining lock-wait time)
	if maxWait, ok := output.GetRemoteLockedMaxWait(); ok {
		r.executionTime -= maxWait
	} else {
		r.executionTime -= summary.Duration
	}
	if r.executionTime < 0 {
		r.executionTime = 0
	}
	clog.Debugf("repository lock failed when running '%s', counted execution time %s", command, r.executionTime.Truncate(time.Second))
	retry, sleep := r.canRetryAfterRemoteLockFailure(output)

	if retry && sleep > 0 {
		err := interruptibleSleep(sleep, r.sigChan)
		if err != nil {
			return false, err
		}
	}

	return retry, nil
}

func (r *resticWrapper) canRetryAfterRemoteLockFailure(output monitor.OutputAnalysis) (bool, time.Duration) {
	if !output.ContainsRemoteLockFailure() {
		return false, 0
	}

	// Check if the remote lock is stale
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

	// Check if we have time left to wait on a non-stale lock
	if remainingTime, enabled := r.remainingLockRetryTime(); enabled && remainingTime > 0 {
		retryDelay := r.global.ResticLockRetryAfter
		if retryDelay < constants.MinResticLockRetryDelay {
			retryDelay = constants.MinResticLockRetryDelay
		} else if retryDelay > constants.MaxResticLockRetryDelay {
			retryDelay = constants.MaxResticLockRetryDelay
		}

		if retryDelay > remainingTime {
			retryDelay = remainingTime
		}

		if retryDelay >= constants.MinResticLockRetryDelay {
			lockName := r.profile.Repository.String()
			if lockedBy, ok := output.GetRemoteLockedBy(); ok {
				lockName = fmt.Sprintf("%s locked by %s", lockName, lockedBy)
			}
			if r.lockWait != nil {
				logLockWait(lockName, r.startTime, time.Unix(0, 0), r.executionTime, *r.lockWait)
			}

			return true, retryDelay
		}
		return false, 0
	}

	return false, 0
}

func (r *resticWrapper) remainingLockRetryTime() (remaining time.Duration, enabled bool) {
	enabled = r.global.ResticLockRetryAfter > 0 && r.lockWait != nil
	if enabled {
		elapsedTime := time.Since(r.startTime)
		remaining = *r.lockWait - elapsedTime + r.executionTime
	}
	if remaining < 0 {
		remaining = 0
	}
	return
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
