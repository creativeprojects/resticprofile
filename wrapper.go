package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/lock"
	"github.com/creativeprojects/resticprofile/status"
	"github.com/creativeprojects/resticprofile/term"
)

type resticWrapper struct {
	resticBinary string
	initialize   bool
	dryRun       bool
	profile      *config.Profile
	command      string
	moreArgs     []string
	sigChan      chan os.Signal
	setPID       func(pid int)
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
		profile:      profile,
		command:      command,
		moreArgs:     moreArgs,
		sigChan:      c,
	}
}

func (r *resticWrapper) runProfile() error {
	err := lockRun(r.profile.Lock, r.profile.ForceLock, func(setPID lock.SetPID) error {
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
	err := runShellCommand(rCommand)
	if err != nil {
		return fmt.Errorf("repository initialization on profile '%s': %w", r.profile.Name, err)
	}
	return nil
}

func (r *resticWrapper) runCheck() error {
	clog.Infof("profile '%s': checking repository consistency", r.profile.Name)
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandCheck))
	rCommand := r.prepareCommand(constants.CommandCheck, args)
	err := runShellCommand(rCommand)
	if err != nil {
		r.statusError(constants.CommandCheck, err)
		return fmt.Errorf("backup check on profile '%s': %w", r.profile.Name, err)
	}
	r.statusSuccess(constants.CommandCheck)
	return nil
}

func (r *resticWrapper) runRetention() error {
	clog.Infof("profile '%s': cleaning up repository using retention information", r.profile.Name)
	args := convertIntoArgs(r.profile.GetRetentionFlags())
	rCommand := r.prepareCommand(constants.CommandForget, args)
	err := runShellCommand(rCommand)
	if err != nil {
		r.statusError(constants.SectionConfigurationRetention, err)
		return fmt.Errorf("backup retention on profile '%s': %w", r.profile.Name, err)
	}
	r.statusSuccess(constants.SectionConfigurationRetention)
	return nil
}

func (r *resticWrapper) runCommand(command string) error {
	clog.Infof("profile '%s': starting '%s'", r.profile.Name, command)
	args := convertIntoArgs(r.profile.GetCommandFlags(command))
	rCommand := r.prepareCommand(command, args)
	err := runShellCommand(rCommand)
	if err != nil {
		r.statusError(r.command, err)
		return fmt.Errorf("%s on profile '%s': %w", r.command, r.profile.Name, err)
	}
	r.statusSuccess(r.command)
	clog.Infof("profile '%s': finished '%s'", r.profile.Name, command)
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
		err := runShellCommand(rCommand)
		if err != nil {
			return fmt.Errorf("run-before backup on profile '%s': %w", r.profile.Name, err)
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
		err := runShellCommand(rCommand)
		if err != nil {
			return fmt.Errorf("run-after backup on profile '%s': %w", r.profile.Name, err)
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
		err := runShellCommand(rCommand)
		if err != nil {
			return fmt.Errorf("run-before on profile '%s': %w", r.profile.Name, err)
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
		err := runShellCommand(rCommand)
		if err != nil {
			return fmt.Errorf("run-after on profile '%s': %w", r.profile.Name, err)
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

	for i, postCommand := range r.profile.RunAfterFail {
		clog.Debugf("starting 'run-after-fail' profile command %d/%d", i+1, len(r.profile.RunAfterFail))
		rCommand := newShellCommand(postCommand, nil, env, r.dryRun, r.sigChan, r.setPID)
		// stdout are stderr are coming from the default terminal (in case they're redirected)
		rCommand.stdout = term.GetOutput()
		rCommand.stderr = term.GetErrorOutput()
		err := runShellCommand(rCommand)
		if err != nil {
			return err
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

func (r *resticWrapper) statusSuccess(command string) {
	if r.profile.StatusFile == "" {
		return
	}
	var err error
	switch command {
	case constants.CommandBackup:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).BackupSuccess()
		err = status.Save()
	case constants.CommandCheck:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).CheckSuccess()
		err = status.Save()
	case constants.SectionConfigurationRetention, constants.CommandForget:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).RetentionSuccess()
		err = status.Save()
	}
	if err != nil {
		// not important enough to throw an error here
		clog.Warningf("saving status file '%s': %v", r.profile.StatusFile, err)
	}
}

func (r *resticWrapper) statusError(command string, fail error) {
	if r.profile.StatusFile == "" {
		return
	}
	var err error
	switch command {
	case constants.CommandBackup:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).BackupError(fail)
		err = status.Save()
	case constants.CommandCheck:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).CheckError(fail)
		err = status.Save()
	case constants.SectionConfigurationRetention, constants.CommandForget:
		status := status.NewStatus(r.profile.StatusFile).Load()
		status.Profile(r.profile.Name).RetentionError(fail)
		err = status.Save()
	}
	if err != nil {
		// not important enough to throw an error here
		clog.Warningf("saving status file '%s': %v", r.profile.StatusFile, err)
	}
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
func lockRun(filename string, force bool, run func(setPID lock.SetPID) error) error {
	if filename == "" {
		// No lock
		return run(nil)
	}
	// Make sure the path to the lock exists
	dir := filepath.Dir(filename)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			clog.Warningf("the profile will run without a lockfile: %v", err)
			return run(nil)
		}
	}
	runLock := lock.NewLock(filename)
	success := runLock.TryAcquire()
	if !success {
		who, err := runLock.Who()
		if err != nil {
			return fmt.Errorf("another process left the lockfile unreadable: %s", err)
		}
		// should we try to force our way?
		if force {
			clog.Warningf("previous run of the profile started by %s hasn't finished properly", who)
			success = runLock.ForceAcquire()
		}
		if !success {
			return fmt.Errorf("another process is already running this profile: %s", who)
		}
	}
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
