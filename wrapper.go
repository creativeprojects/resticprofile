package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
)

type resticWrapper struct {
	resticBinary string
	profile      *config.Profile
	moreArgs     []string
}

func newResticWrapper(resticBinary string, profile *config.Profile, moreArgs []string) *resticWrapper {
	return &resticWrapper{
		resticBinary: resticBinary,
		profile:      profile,
		moreArgs:     moreArgs,
	}
}

func (r *resticWrapper) runInitialize() error {
	clog.Info("Initializing repository (if not existing)")
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandInit))
	rCommand := r.prepareCommand(constants.CommandInit, args)
	rCommand.displayStderr = false
	return runCommand(rCommand)
}

func (r *resticWrapper) runCheck() error {
	clog.Info("Checking repository consistency")
	args := convertIntoArgs(r.profile.GetCommandFlags(constants.CommandCheck))
	rCommand := r.prepareCommand(constants.CommandCheck, args)
	return runCommand(rCommand)
}

func (r *resticWrapper) runRetention() error {
	clog.Info("Cleaning up repository using retention information")
	args := convertIntoArgs(r.profile.GetRetentionFlags())
	rCommand := r.prepareCommand(constants.CommandForget, args)
	return runCommand(rCommand)
}

func (r *resticWrapper) runCommand(command string) error {
	clog.Infof("Starting '%s'", command)
	args := convertIntoArgs(r.profile.GetCommandFlags(command))
	rCommand := r.prepareCommand(command, args)
	err := runCommand(rCommand)
	clog.Infof("Finished '%s'", command)
	return err
}

func (r *resticWrapper) prepareCommand(command string, args []string) commandDefinition {
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
	rCommand := newCommand(r.resticBinary, arguments, env)

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
		runShellCommand(preCommand)
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
		runShellCommand(postCommand)
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
