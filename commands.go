package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/win"
	"golang.org/x/exp/maps"
)

var (
	ownCommands = NewOwnCommands()
)

func init() {
	ownCommands.Register(getOwnCommands())

	// own commands have no profile section, prevent their definition
	for _, command := range ownCommands.All() {
		config.ExcludeProfileSection(command.name)
	}
}

func getOwnCommands() []ownCommand {
	return []ownCommand{
		// commands that don't need loading the configuration
		{
			name:              "help",
			description:       "display help (use resticprofile help [command])",
			longDescription:   "The \"help\" command prints commandline help on resticprofile flags & commands (own and restic) including information on flags passed to restic from various profiles defined in the current configuration.\n\nHelp on a specific command is displayed by adding the command name as argument after \"help\", e.g. use \"resticprofile help version\" to get details on the \"version\" command.",
			action:            displayHelpCommand,
			needConfiguration: false,
		},
		{
			name:              "version",
			description:       "display version (run in verbose mode for detailed information)",
			longDescription:   "The \"version\" command displays brief or detailed version information",
			action:            displayVersion,
			needConfiguration: false,
			flags:             map[string]string{"-v, --verbose": "display detailed version information"},
		},
		{
			name:              "random-key",
			description:       "generate a cryptographically secure random key to use as a restic keyfile",
			action:            randomKey,
			needConfiguration: false,
			hide:              true, // replaced by the generate command
		},
		{
			name:              "generate",
			description:       "generate resources such as random key, bash/zsh completion scripts, etc.",
			longDescription:   "The \"generate\" command is used to create various resources and print them to stdout",
			action:            generateCommand,
			needConfiguration: false,
			hide:              false,
			flags: map[string]string{
				"--random-key [size]":                            "generate a cryptographically secure random key to use as a restic keyfile (size defaults to 1024 when omitted)",
				"--config-reference [--version 0.15] [template]": "generate a config file reference from a go template (defaults to the built-in markdown template when omitted)",
				"--json-schema [--version 0.15] [v1|v2]":         "generate a JSON schema that validates resticprofile configuration files in YAML or JSON format",
				"--bash-completion":                              "generate a shell completion script for bash",
				"--zsh-completion":                               "generate a shell completion script for zsh",
			},
		},
		// commands that need the configuration
		{
			name:              "profiles",
			description:       "display profile names from the configuration file",
			longDescription:   "The \"profiles\" command prints brief information on all profiles and groups that are declared in the configuration file",
			action:            displayProfilesCommand,
			needConfiguration: true,
		},
		{
			name:              "show",
			description:       "show all the details of the current profile or group",
			longDescription:   "The \"show\" command prints the effective configuration of the selected profile or group.\n\nThe effective profile or group configuration is built by loading all includes, applying inheritance, mixins, templates and variables and parsing the result.",
			action:            showProfileOrGroup,
			needConfiguration: true,
		},
		{
			name:              "schedule",
			description:       "schedule jobs from a profile or group (or of all profiles and groups)",
			longDescription:   "The \"schedule\" command registers declared schedules of the selected profile or group (or of all profiles and groups) as scheduled jobs within the scheduling service of the operating system",
			action:            createSchedule,
			needConfiguration: true,
			hide:              false,
			flags: map[string]string{
				"--no-start": "don't start the timer/service (systemd/launch only)",
				"--all":      "add all scheduled jobs of all profiles and groups",
			},
		},
		{
			name:              "unschedule",
			description:       "remove scheduled jobs of a profile or group (or of all profiles and groups)",
			longDescription:   "The \"unschedule\" command removes scheduled jobs from the scheduling service of the operating system. The command removes jobs for schedules declared in the selected profile or group (or of all profiles and groups)",
			action:            removeSchedule,
			needConfiguration: true,
			hide:              false,
			flags:             map[string]string{"--all": "remove all scheduled jobs of all profiles and groups"},
		},
		{
			name:              "status",
			description:       "display the status of scheduled jobs of a profile or group (or of all profiles and groups)",
			longDescription:   "The \"status\" command prints all declared schedules of the selected profile or group (or of all profiles and groups) and shows the status of related scheduled jobs in the scheduling service of the operating system",
			action:            statusSchedule,
			needConfiguration: true,
			hide:              false,
			flags:             map[string]string{"--all": "display the status of all scheduled jobs of all profiles and groups"},
		},
		{
			name:              "run-schedule",
			description:       "runs a scheduled job. This command should only be called by the scheduling service",
			longDescription:   "The \"run-schedule\" command loads the scheduled job configuration from the name in parameter and runs the restic command with the arguments defined in the profile or group. The name in parameter is <command>@<profile-or-group-name>.",
			pre:               preRunSchedule,
			action:            runSchedule,
			needConfiguration: true,
			hide:              false,
			hideInCompletion:  true,
			noProfile:         true,
		},
		// hidden commands
		{
			name:              "complete",
			description:       "create commandline completion results based on given args",
			action:            completeCommand,
			needConfiguration: false,
			hide:              true,
		},
		{
			name:              "elevation",
			description:       "test windows elevated mode",
			action:            testElevationCommand,
			needConfiguration: false,
			hide:              true,
		},
		{
			name:              "panic",
			description:       "(debug only) simulates a panic",
			action:            panicCommand,
			needConfiguration: false,
			hide:              true,
		},
	}
}

func panicCommand(_ io.Writer, _ commandContext) error {
	panic("you asked for it")
}

func completeCommand(output io.Writer, ctx commandContext) error {
	args := ctx.request.arguments
	requester := "unknown"
	requesterVersion := 0

	// Parse requester as first argument. Format "[kind]:v[version]", e.g. "bash:v1"
	if len(args) > 0 {
		matcher := regexp.MustCompile(`^(bash|zsh):v(\d+)$`)
		if matches := matcher.FindStringSubmatch(args[0]); matches != nil {
			requester = matches[1]
			if v, err := strconv.Atoi(matches[2]); err == nil {
				requesterVersion = v
			}
			args = args[1:]
		}
	}

	// Log when requester isn't specified.
	if requester == "unknown" || requesterVersion < 1 {
		clog.Warningf("Requester %q version %d not explicitly supported", requester, requesterVersion)
	}

	// Ensure newer completion scripts will not fail on outdated resticprofile
	if requester == "zsh" || requesterVersion > 9 {
		return nil
	}

	completions := NewCompleter(ctx.ownCommands.All(), DefaultFlagsLoader).Complete(args)
	if len(completions) > 0 {
		for _, completion := range completions {
			fmt.Fprintln(output, completion)
		}
	}
	return nil
}

func showProfileOrGroup(output io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags

	// Load global section
	global, err := c.GetGlobalSection()
	if err != nil {
		return fmt.Errorf("cannot load global section: %w", err)
	}

	var profileOrGroup config.Schedulable
	if c.HasProfile(flags.name) {
		// Load profile
		profile, cleanup, err := openProfile(c, flags.name)
		defer cleanup()
		if err != nil {
			return fmt.Errorf("profile '%s': %w", flags.name, err)
		}
		if profile == nil {
			return fmt.Errorf("cannot load profile '%s'", flags.name)
		}
		profileOrGroup = profile

	} else if c.HasProfileGroup(flags.name) {
		group, err := c.GetProfileGroup(flags.name)
		if err != nil {
			return fmt.Errorf("group '%s': %w", flags.name, err)
		}
		if group == nil {
			return fmt.Errorf("cannot load group '%s'", flags.name)
		}
		profileOrGroup = group

	} else {
		return fmt.Errorf("profile or group '%s': %w", flags.name, config.ErrNotFound)
	}

	// Show global
	err = config.ShowStruct(output, global, constants.SectionConfigurationGlobal)
	if err != nil {
		clog.Errorf("cannot show global section: %s", err.Error())
	}
	_, _ = fmt.Fprintln(output)

	// Show profile or group
	err = config.ShowStruct(output, profileOrGroup, profileOrGroup.Kind()+" "+flags.name)
	if err != nil {
		clog.Errorf("cannot show profile or group '%s': %s", flags.name, err.Error())
	}
	_, _ = fmt.Fprintln(output)

	// Show schedules
	showSchedules(output, maps.Values(profileOrGroup.Schedules()))

	if profile, ok := profileOrGroup.(*config.Profile); ok {
		// Show deprecation notice
		displayDeprecationNotices(profile)
	}

	// Show config issues
	c.DisplayConfigurationIssues()

	return nil
}

func showSchedules(output io.Writer, schedules []*config.Schedule) {
	slices.SortFunc(schedules, config.CompareSchedules)
	for _, schedule := range schedules {
		err := config.ShowStruct(output, schedule.ScheduleConfig, fmt.Sprintf("schedule %s", schedule.ScheduleOrigin()))
		if err != nil {
			_, _ = fmt.Fprintln(output, err)
		}
		_, _ = fmt.Fprintln(output, "")
	}
}

// randomKey simply display a base64'd random key to the console
func randomKey(output io.Writer, ctx commandContext) error {
	var err error
	flags := ctx.flags
	size := uint64(1024)
	// flags.resticArgs contain the command and the rest of the command line
	if len(flags.resticArgs) > 1 {
		// second parameter should be an integer
		size, err = strconv.ParseUint(flags.resticArgs[1], 10, 32)
		if err != nil {
			return fmt.Errorf("cannot parse the key size: %w", err)
		}
		if size < 1 {
			return fmt.Errorf("invalid key size: %v", size)
		}
	}
	buffer := make([]byte, size)
	_, err = rand.Read(buffer)
	if err != nil {
		return err
	}
	encoder := base64.NewEncoder(base64.StdEncoding, output)
	_, err = encoder.Write(buffer)
	encoder.Close()
	fmt.Fprintln(output, "")
	return err
}

func testElevationCommand(_ io.Writer, ctx commandContext) error {
	if ctx.flags.isChild {
		client := remote.NewClient(ctx.flags.parentPort)
		term.Print("first line", "\n")
		term.Println("second", "one")
		term.Printf("value = %d\n", 11)
		err := client.Done()
		if err != nil {
			return err
		}
		return nil
	}

	return elevated()
}

func retryElevated(err error, flags commandLineFlags) error {
	if err == nil {
		return nil
	}
	// maybe can find a better way than searching for the word "denied"?
	if platform.IsWindows() && !flags.isChild && strings.Contains(err.Error(), "denied") {
		clog.Info("restarting resticprofile in elevated mode...")
		err := elevated()
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

func elevated() error {
	if !platform.IsWindows() {
		return errors.New("only available on Windows platform")
	}

	done := make(chan interface{})
	err := remote.StartServer(done)
	if err != nil {
		return err
	}
	err = win.RunElevated(remote.GetPort())
	if err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		remote.StopServer(ctx)
		return err
	}

	// wait until the server is done
	<-done

	return nil
}
