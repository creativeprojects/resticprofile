package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
)

type resticWrapper struct {
	resticBinary string
	profile      *config.Profile
	moreArgs     []string
	sigChan      chan os.Signal
}

func newResticWrapper(resticBinary string, profile *config.Profile, moreArgs []string, c chan os.Signal) *resticWrapper {
	return &resticWrapper{
		resticBinary: resticBinary,
		profile:      profile,
		moreArgs:     moreArgs,
		sigChan:      c,
	}
}

func (r *resticWrapper) runInitialize() error {
	clog.Infof("Profile '%s': Initializing repository (if not existing)", r.profile.Name)
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandInit))
	rCommand := r.prepareCommand(constants.CommandInit, args)
	rCommand.displayStderr = false
	return runShellCommand(rCommand)
}

func (r *resticWrapper) runCheck() error {
	clog.Infof("Profile '%s': Checking repository consistency", r.profile.Name)
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandCheck))
	rCommand := r.prepareCommand(constants.CommandCheck, args)
	return runShellCommand(rCommand)
}

func (r *resticWrapper) runRetention() error {
	clog.Infof("Profile '%s': Cleaning up repository using retention information", r.profile.Name)
	args := convertIntoArgs(r.profile.GetRetentionFlags())
	rCommand := r.prepareCommand(constants.CommandForget, args)
	return runShellCommand(rCommand)
}

func (r *resticWrapper) runCommand(command string) error {
	clog.Infof("Profile '%s': Starting '%s'", r.profile.Name, command)
	args := convertIntoArgs(r.profile.GetCommandFlags(command))
	rCommand := r.prepareCommand(command, args)
	err := runShellCommand(rCommand)
	clog.Infof("Profile '%s': Finished '%s'", r.profile.Name, command)
	return err
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

	clog.Debugf("Starting command: %s %s", r.resticBinary, strings.Join(arguments, " "))
	rCommand := newShellCommand(r.resticBinary, arguments, env)
	rCommand.sigChan = r.sigChan

	if command == constants.CommandBackup && r.profile.Backup.UseStdin {
		clog.Debug("Redirecting stdin to the backup")
		rCommand.useStdin = true
	}
	return rCommand
}

func (r *resticWrapper) runPreCommand(command string) error {
	// Pre/Post commands are only supported for backup
	if command != constants.CommandBackup {
		return nil
	}
	if r.profile.Backup.RunBefore == nil || len(r.profile.Backup.RunBefore) == 0 {
		return nil
	}
	for i, preCommand := range r.profile.Backup.RunBefore {
		clog.Debugf("Starting pre-backup command %d/%d", i+1, len(r.profile.Backup.RunBefore))
		env := append(os.Environ(), r.getEnvironment()...)
		rCommand := newShellCommand(preCommand, nil, env)
		rCommand.sigChan = r.sigChan
		err := runShellCommand(rCommand)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *resticWrapper) runPostCommand(command string) error {
	// Pre/Post commands are only supported for backup
	if command != constants.CommandBackup {
		return nil
	}
	if r.profile.Backup.RunAfter == nil || len(r.profile.Backup.RunAfter) == 0 {
		return nil
	}
	for i, postCommand := range r.profile.Backup.RunAfter {
		clog.Debugf("Starting post-backup command %d/%d", i+1, len(r.profile.Backup.RunAfter))
		env := append(os.Environ(), r.getEnvironment()...)
		rCommand := newShellCommand(postCommand, nil, env)
		rCommand.sigChan = r.sigChan
		err := runShellCommand(rCommand)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *resticWrapper) getEnvironment() []string {
	if r.profile.Environment == nil || len(r.profile.Environment) == 0 {
		return nil
	}
	env := make([]string, len(r.profile.Environment))
	i := 0
	for key, value := range r.profile.Environment {
		// env variables are always uppercase
		key = strings.ToUpper(key)
		clog.Debugf("Setting up environment variable '%s'", key)
		env[i] = fmt.Sprintf("%s=%s", key, value)
		i++
	}
	return env
}

func convertIntoArgs(flags map[string][]string) []string {
	args := make([]string, 0)

	if flags == nil || len(flags) == 0 {
		return args
	}

	for key, values := range flags {
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
