package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/config/jsonschema"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util/templates"
	"github.com/creativeprojects/resticprofile/win"
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
			description:       "show all the details of the current profile",
			longDescription:   "The \"show\" command prints the effective configuration of the selected profile.\n\nThe effective profile configuration is built by loading all includes, applying inheritance, mixins, templates and variables and parsing the result.",
			action:            showProfile,
			needConfiguration: true,
		},
		{
			name:              "schedule",
			description:       "schedule jobs from a profile (or of all profiles)",
			longDescription:   "The \"schedule\" command registers declared schedules of the selected profile (or of all profiles) as scheduled jobs within the scheduling service of the operating system",
			action:            createSchedule,
			needConfiguration: true,
			hide:              false,
			flags: map[string]string{
				"--no-start": "don't start the timer/service (systemd/launch only)",
				"--all":      "add all scheduled jobs of all profiles",
			},
		},
		{
			name:              "unschedule",
			description:       "remove scheduled jobs of a profile (or of all profiles)",
			longDescription:   "The \"unschedule\" command removes scheduled jobs from the scheduling service of the operating system. The command removes jobs for schedules declared in the selected profile (or of all profiles)",
			action:            removeSchedule,
			needConfiguration: true,
			hide:              false,
			flags:             map[string]string{"--all": "remove all scheduled jobs of all profiles"},
		},
		{
			name:              "status",
			description:       "display the status of scheduled jobs of a profile (or of all profiles)",
			longDescription:   "The \"status\" command prints all declared schedules of the selected profile (or of all profiles) and shows the status of related scheduled jobs in the scheduling service of the operating system",
			action:            statusSchedule,
			needConfiguration: true,
			hide:              false,
			flags:             map[string]string{"--all": "display the status of all scheduled jobs of all profiles"},
		},
		{
			name:              "run-schedule",
			description:       "runs a scheduled job. This command should only be called by the scheduling service",
			longDescription:   "The \"run-schedule\" command loads the scheduled job configuration from the name in parameter and runs the restic command with the arguments defined in the profile. The name in parameter is <command>@<profile-name> for the configuration file v1, and the schedule name for the configuration file v2+.",
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
		matcher := regexp.MustCompile("^(bash|zsh):v(\\d+)$")
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

//go:embed contrib/completion/bash-completion.sh
var bashCompletionScript string

//go:embed contrib/completion/zsh-completion.sh
var zshCompletionScript string

func generateCommand(output io.Writer, ctx commandContext) (err error) {
	args := ctx.request.arguments
	// enforce no-log
	logger := clog.GetDefaultLogger()
	handler := logger.GetHandler()
	logger.SetHandler(clog.NewDiscardHandler())

	if slices.Contains(args, "--bash-completion") {
		_, err = fmt.Fprintln(output, bashCompletionScript)
	} else if slices.Contains(args, "--config-reference") {
		err = generateConfigReference(output, args[slices.Index(args, "--config-reference")+1:])
	} else if slices.Contains(args, "--json-schema") {
		err = generateJsonSchema(output, args[slices.Index(args, "--json-schema")+1:])
	} else if slices.Contains(args, "--random-key") {
		ctx.flags.resticArgs = args[slices.Index(args, "--random-key"):]
		err = randomKey(output, ctx)
	} else if slices.Contains(args, "--zsh-completion") {
		_, err = fmt.Fprintln(output, zshCompletionScript)
	} else {
		err = fmt.Errorf("nothing to generate for: %s", strings.Join(args, ", "))
	}

	if err != nil {
		logger.SetHandler(handler)
	}
	return
}

//go:embed contrib/templates/config-reference.gomd
var configReferenceTemplate string

func generateConfigReference(output io.Writer, args []string) (err error) {
	resticVersion := restic.AnyVersion
	if slices.Contains(args, "--version") {
		args = args[slices.Index(args, "--version"):]
		if len(args) > 1 {
			resticVersion = args[1]
			args = args[2:]
		}
	}

	data := config.NewTemplateInfoData(resticVersion)
	tpl := templates.New("config-reference", data.GetFuncs())

	if len(args) > 0 {
		tpl, err = tpl.ParseFiles(args...)
	} else {
		tpl, err = tpl.Parse(configReferenceTemplate)
	}

	if err != nil {
		err = fmt.Errorf("parsing failed: %w", err)
	} else {
		err = tpl.Execute(output, data)
	}
	return
}

func generateJsonSchema(output io.Writer, args []string) (err error) {
	resticVersion := restic.AnyVersion
	if slices.Contains(args, "--version") {
		args = args[slices.Index(args, "--version"):]
		if len(args) > 1 {
			resticVersion = args[1]
			args = args[2:]
		}
	}

	version := config.Version02
	if len(args) > 0 && args[0] == "v1" {
		version = config.Version01
	}

	return jsonschema.WriteJsonSchema(version, resticVersion, output)
}

func sortedProfileKeys(data map[string]*config.Profile) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func showProfile(output io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags

	// Load global section
	global, err := c.GetGlobalSection()
	if err != nil {
		return fmt.Errorf("cannot load global section: %w", err)
	}

	// Load profile
	profile, cleanup, err := openProfile(c, flags.name)
	defer cleanup()
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return fmt.Errorf("profile '%s' not found", flags.name)
		}
		if profile == nil {
			return fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
		} else {
			clog.Errorf("failed loading profile '%s': %s", flags.name, err)
		}
	}

	// Show global
	err = config.ShowStruct(output, global, constants.SectionConfigurationGlobal)
	if err != nil {
		clog.Errorf("cannot show global section: %s", err.Error())
	}
	_, _ = fmt.Fprintln(output)

	// Show profile
	err = config.ShowStruct(output, profile, "profile "+flags.name)
	if err != nil {
		clog.Errorf("cannot show profile '%s': %s", flags.name, err.Error())
	}
	_, _ = fmt.Fprintln(output)

	// Show schedules
	showSchedules(output, profile.Schedules())

	// Show deprecation notice
	displayProfileDeprecationNotices(profile)

	// Show config issues
	c.DisplayConfigurationIssues()

	return nil
}

func showSchedules(output io.Writer, schedules []*config.Schedule) {
	for _, schedule := range schedules {
		err := config.ShowStruct(output, schedule, "schedule "+schedule.CommandName+"@"+schedule.Profiles[0])
		if err != nil {
			fmt.Fprintln(output, err)
		}
		fmt.Fprintln(output, "")
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

// selectProfiles returns a list with length >= 1, containing profile names that have been selected in flags or extra args.
// With "--all" set in args names of all profiles are returned, else profile or profile group referenced in flags.name returns names.
// If no match, flags.name is returned as-is.
func selectProfiles(c *config.Config, flags commandLineFlags, args []string) []string {
	profiles := make([]string, 0, 1)

	// Check for --all or groups
	if slices.Contains(args, "--all") {
		profiles = c.GetProfileNames()

	} else if !c.HasProfile(flags.name) {
		if names, err := c.GetProfileGroup(flags.name); err == nil && names != nil {
			profiles = names.Profiles
		}
	}

	// Fallback add profile name from flags
	if len(profiles) == 0 {
		profiles = append(profiles, flags.name)
	}

	return profiles
}

// flagsForProfile returns a copy of flags with name set to profileName.
func flagsForProfile(flags commandLineFlags, profileName string) commandLineFlags {
	flags.name = profileName
	return flags
}

// createSchedule accepts one argument from the commandline: --no-start
func createSchedule(_ io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	type profileJobs struct {
		scheduler schedule.SchedulerConfig
		profile   string
		jobs      []*config.Schedule
	}

	allJobs := make([]profileJobs, 0, 1)

	// Step 1: Collect all jobs of all selected profiles
	for _, profileName := range selectProfiles(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)

		scheduler, profile, jobs, err := getScheduleJobs(c, profileFlags)
		if err == nil {
			err = requireScheduleJobs(jobs, profileFlags)

			// Skip profile with no schedules when "--all" option is set.
			if err != nil && slices.Contains(args, "--all") {
				continue
			}
		}
		if err != nil {
			return err
		}

		displayProfileDeprecationNotices(profile)

		// add the no-start flag to all the jobs
		if slices.Contains(args, "--no-start") {
			for id := range jobs {
				jobs[id].SetFlag("no-start", "")
			}
		}

		allJobs = append(allJobs, profileJobs{scheduler: scheduler, profile: profileName, jobs: jobs})
	}

	// Step 2: Schedule all collected jobs
	for _, j := range allJobs {
		err := scheduleJobs(schedule.NewHandler(j.scheduler), j.profile, j.jobs)
		if err != nil {
			return retryElevated(err, flags)
		}
	}

	return nil
}

func removeSchedule(_ io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	// Unschedule all jobs of all selected profiles
	for _, profileName := range selectProfiles(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)

		scheduler, _, jobs, err := getRemovableScheduleJobs(c, profileFlags)
		if err != nil {
			return err
		}

		err = removeJobs(schedule.NewHandler(scheduler), profileName, jobs)
		if err != nil {
			return retryElevated(err, flags)
		}
	}

	return nil
}

func statusSchedule(w io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	if !slices.Contains(args, "--all") {
		// simple case of displaying status for one profile
		scheduler, profile, schedules, err := getScheduleJobs(c, flags)
		if err != nil {
			return err
		}
		if len(schedules) == 0 {
			clog.Warningf("profile %s has no schedule", flags.name)
			return nil
		}
		return statusScheduleProfile(scheduler, profile, schedules, flags)
	}

	for _, profileName := range selectProfiles(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)
		scheduler, profile, schedules, err := getScheduleJobs(c, profileFlags)
		if err != nil {
			return err
		}
		// it's all fine if this profile has no schedule
		if len(schedules) == 0 {
			continue
		}
		clog.Infof("Profile %q:", profileName)
		err = statusScheduleProfile(scheduler, profile, schedules, profileFlags)
		if err != nil {
			// display the error but keep going with the other profiles
			clog.Error(err)
		}
	}
	return nil
}

func statusScheduleProfile(scheduler schedule.SchedulerConfig, profile *config.Profile, schedules []*config.Schedule, flags commandLineFlags) error {
	displayProfileDeprecationNotices(profile)

	err := statusJobs(schedule.NewHandler(scheduler), flags.name, schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func getScheduleJobs(c *config.Config, flags commandLineFlags) (schedule.SchedulerConfig, *config.Profile, []*config.Schedule, error) {
	global, err := c.GetGlobalSection()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot load global section: %w", err)
	}

	profile, err := c.GetProfile(flags.name)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil, nil, nil, fmt.Errorf("profile '%s' not found", flags.name)
		}
		return nil, nil, nil, fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
	}

	return schedule.NewSchedulerConfig(global), profile, profile.Schedules(), nil
}

func requireScheduleJobs(schedules []*config.Schedule, flags commandLineFlags) error {
	if len(schedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", flags.name)
	}
	return nil
}

func getRemovableScheduleJobs(c *config.Config, flags commandLineFlags) (schedule.SchedulerConfig, *config.Profile, []*config.Schedule, error) {
	scheduler, profile, schedules, err := getScheduleJobs(c, flags)
	if err != nil {
		return nil, nil, nil, err
	}

	// Add all undeclared schedules as remove-only configs
	for _, command := range profile.SchedulableCommands() {
		declared := false
		for _, s := range schedules {
			if declared = s.CommandName == command; declared {
				break
			}
		}
		if !declared {
			schedules = append(schedules, config.NewEmptySchedule(profile.Name, command))
		}
	}

	return scheduler, profile, schedules, nil
}

func preRunSchedule(ctx *Context) error {
	if len(ctx.request.arguments) < 1 {
		return errors.New("run-schedule command expects one argument: schedule name")
	}
	scheduleName := ctx.request.arguments[0]
	// temporarily allow v2 configuration to run v1 schedules
	// if ctx.config.GetVersion() < config.Version02
	{
		// schedule name is in the form "command@profile"
		commandName, profileName, ok := strings.Cut(scheduleName, "@")
		if !ok {
			return errors.New("the expected format of the schedule name is <command>@<profile-name>")
		}
		ctx.request.profile = profileName
		ctx.request.schedule = scheduleName
		ctx.command = commandName
		// remove the parameter from the arguments
		ctx.request.arguments = ctx.request.arguments[1:]

		// don't save the profile in the context now, it's only loaded but not prepared
		profile, err := ctx.config.GetProfile(profileName)
		if err != nil || profile == nil {
			return fmt.Errorf("cannot load profile '%s': %w", profileName, err)
		}
		// get the list of all scheduled commands to find the current command
		schedules := profile.Schedules()
		for _, schedule := range schedules {
			if schedule.CommandName == ctx.command {
				ctx.schedule = schedule
				prepareScheduledProfile(ctx)
				break
			}
		}
	}
	return nil
}

func prepareScheduledProfile(ctx *Context) {
	clog.Debugf("preparing scheduled profile %q", ctx.request.schedule)
	// log file
	if len(ctx.schedule.Log) > 0 {
		ctx.logTarget = ctx.schedule.Log
	}
	// battery
	if ctx.schedule.IgnoreOnBatteryLessThan > 0 {
		ctx.stopOnBattery = ctx.schedule.IgnoreOnBatteryLessThan
	} else if ctx.schedule.IgnoreOnBattery {
		ctx.stopOnBattery = 100
	}
	// lock
	if ctx.schedule.GetLockWait() > 0 {
		ctx.lockWait = ctx.schedule.LockWait
	}
	if ctx.schedule.GetLockMode() == config.ScheduleLockModeDefault {
		if ctx.schedule.GetLockWait() > 0 {
			ctx.lockWait = ctx.schedule.GetLockWait()
		}
	} else if ctx.schedule.GetLockMode() == config.ScheduleLockModeIgnore {
		ctx.noLock = true
	}
}

func runSchedule(_ io.Writer, cmdCtx commandContext) error {
	err := startProfileOrGroup(&cmdCtx.Context)
	if err != nil {
		return err
	}
	return nil
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

	return elevated(ctx.flags)
}

func retryElevated(err error, flags commandLineFlags) error {
	if err == nil {
		return nil
	}
	// maybe can find a better way than searching for the word "denied"?
	if platform.IsWindows() && !flags.isChild && strings.Contains(err.Error(), "denied") {
		clog.Info("restarting resticprofile in elevated mode...")
		err := elevated(flags)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

func elevated(flags commandLineFlags) error {
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
		remote.StopServer()
		return err
	}

	// wait until the server is done
	<-done

	return nil
}
