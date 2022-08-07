package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/win"
)

type ownCommand struct {
	name              string
	description       string
	action            func(io.Writer, *config.Config, commandLineFlags, []string) error
	needConfiguration bool              // true if the action needs a configuration file loaded
	hide              bool              // don't display the command in help and completion
	hideInCompletion  bool              // don't display the command in completion
	flags             map[string]string // own command flags should be simple enough to be handled manually for now
}

var ownCommands []ownCommand

func init() {
	ownCommands = getOwnCommands()
}

func getOwnCommands() []ownCommand {
	return []ownCommand{
		{
			name:              "version",
			description:       "display version (run in verbose mode for detailed information)",
			action:            displayVersion,
			needConfiguration: false,
			flags:             map[string]string{"-v, --verbose": "display details information"},
		},
		{
			name:              "self-update",
			description:       "update to latest resticprofile (use -q/--quiet flag to update without confirmation)",
			action:            selfUpdate,
			needConfiguration: false,
			flags:             map[string]string{"-q, --quiet": "update without confirmation prompt"},
		},
		{
			name:              "profiles",
			description:       "display profile names from the configuration file",
			action:            displayProfilesCommand,
			needConfiguration: true,
		},
		{
			name:              "show",
			description:       "show all the details of the current profile",
			action:            showProfile,
			needConfiguration: true,
		},
		{
			name:              "random-key",
			description:       "generate a cryptographically secure random key to use as a restic keyfile",
			action:            randomKey,
			needConfiguration: false,
			hide:              true,
		},
		{
			name:              "schedule",
			description:       "schedule jobs from a profile (use --all flag to schedule all jobs of all profiles)",
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
			description:       "remove scheduled jobs of a profile (use --all flag to unschedule all profiles)",
			action:            removeSchedule,
			needConfiguration: true,
			hide:              false,
			flags:             map[string]string{"--all": "remove all scheduled jobs of all profiles"},
		},
		{
			name:              "status",
			description:       "display the status of scheduled jobs (use --all flag for all profiles)",
			action:            statusSchedule,
			needConfiguration: true,
			hide:              false,
			flags:             map[string]string{"--all": "display the status of all scheduled jobs of all profiles"},
		},
		{
			name:              "generate",
			description:       "generate resources (--random-key [size], --bash-completion & --zsh-completion)",
			action:            generateCommand,
			needConfiguration: false,
			hide:              false,
			flags: map[string]string{
				"--random-key":      "generate a cryptographically secure random key to use as a restic keyfile",
				"--bash-completion": "generate a shell completion script for bash",
				"--zsh-completion":  "generate a shell completion script for zsh",
			},
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
			needConfiguration: true,
			hide:              true,
		},
		{
			name:              "panic",
			description:       "(debug only) simulates a panic",
			action:            panicCommand,
			needConfiguration: false,
			hide:              true,
		},
		{
			name:              "test",
			description:       "placeholder for a quick test",
			action:            testCommand,
			needConfiguration: true,
			hide:              true,
		},
	}
}

func displayOwnCommands(output io.Writer) {
	commandsWriter := tabwriter.NewWriter(output, 0, 0, 3, ' ', 0)
	for _, command := range ownCommands {
		if command.hide {
			continue
		}
		_, _ = fmt.Fprintf(commandsWriter, "\t%s\t%s\n", command.name, command.description)
		// TODO: find a nice way to display command flags
	}
	_ = commandsWriter.Flush()
}

func isOwnCommand(command string, configurationLoaded bool) bool {
	for _, commandDef := range ownCommands {
		if commandDef.name == command && commandDef.needConfiguration == configurationLoaded {
			return true
		}
	}
	return false
}

func runOwnCommand(configuration *config.Config, command string, flags commandLineFlags, args []string) error {
	for _, commandDef := range ownCommands {
		if commandDef.name == command {
			return commandDef.action(os.Stdout, configuration, flags, args)
		}
	}
	return fmt.Errorf("command not found: %v", command)
}

func displayProfilesCommand(output io.Writer, configuration *config.Config, _ commandLineFlags, _ []string) error {
	displayProfiles(output, configuration)
	displayGroups(output, configuration)
	return nil
}

func displayVersion(output io.Writer, _ *config.Config, flags commandLineFlags, args []string) error {
	fmt.Fprintf(output, "resticprofile version %s commit %q\n", version, commit)

	// allow for the general verbose flag, or specified after the command
	if flags.verbose || (len(args) > 0 && (args[0] == "-v" || args[0] == "--verbose")) {
		w := tabwriter.NewWriter(output, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintf(w, "\n")
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "home", "https://github.com/creativeprojects/resticprofile")
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "os", runtime.GOOS)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "arch", runtime.GOARCH)
		if goarm > 0 {
			_, _ = fmt.Fprintf(w, "\t%s:\tv%d\n", "arm", goarm)
		}
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "version", version)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "commit", commit)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "compiled", date)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "by", builtBy)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "go version", runtime.Version())
		_, _ = fmt.Fprintf(w, "\n")
		_, _ = fmt.Fprintf(w, "\t%s:\n", "go modules")
		bi, _ := debug.ReadBuildInfo()
		for _, dep := range bi.Deps {
			_, _ = fmt.Fprintf(w, "\t\t%s\t%s\n", dep.Path, dep.Version)
		}
		_, _ = fmt.Fprintf(w, "\n")

		w.Flush()
	}
	return nil
}

func displayProfiles(output io.Writer, configuration *config.Config) {
	profiles := configuration.GetProfiles()
	keys := sortedProfileKeys(profiles)
	if len(profiles) == 0 {
		fmt.Fprintln(output, "\nThere's no available profile in the configuration")
	} else {
		fmt.Fprintln(output, "\nProfiles available (name, sections, description):")
		w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
		for _, name := range keys {
			sections := profiles[name].DefinedCommands()
			sort.Strings(sections)
			if len(sections) == 0 {
				_, _ = fmt.Fprintf(w, "\t%s:\t(n/a)\t%s\n", name, profiles[name].Description)
			} else {
				_, _ = fmt.Fprintf(w, "\t%s:\t(%s)\t%s\n", name, strings.Join(sections, ", "), profiles[name].Description)
			}
		}
		_ = w.Flush()
	}
	fmt.Fprintln(output, "")
}

func displayGroups(output io.Writer, configuration *config.Config) {
	groups := configuration.GetProfileGroups()
	if len(groups) == 0 {
		return
	}
	fmt.Fprintln(output, "Groups available (name, profiles, description):")
	w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	for name, groupList := range groups {
		_, _ = fmt.Fprintf(w, "\t%s:\t[%s]\t%s\n", name, strings.Join(groupList.Profiles, ", "), groupList.Description)
	}
	_ = w.Flush()
	fmt.Fprintln(output, "")
}

func selfUpdate(_ io.Writer, _ *config.Config, flags commandLineFlags, args []string) error {
	quiet := flags.quiet
	if !quiet && len(args) > 0 && (args[0] == "-q" || args[0] == "--quiet") {
		quiet = true
	}
	err := confirmAndSelfUpdate(quiet, flags.verbose, version, true)
	if err != nil {
		return err
	}
	return nil
}

func panicCommand(_ io.Writer, _ *config.Config, _ commandLineFlags, _ []string) error {
	panic("you asked for it")
}

func testCommand(_ io.Writer, _ *config.Config, _ commandLineFlags, _ []string) error {
	clog.Info("Nothing to test")
	return nil
}

func completeCommand(output io.Writer, _ *config.Config, _ commandLineFlags, args []string) error {
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

	completions := (&Completer{}).Complete(args)
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

func generateCommand(output io.Writer, config *config.Config, flags commandLineFlags, args []string) (err error) {
	if containsString(args, "--bash-completion") {
		_, err = fmt.Fprintln(output, bashCompletionScript)
	} else if containsString(args, "--random-key") {
		flags.resticArgs = args
		err = randomKey(output, config, flags, args)
	} else if containsString(args, "--zsh-completion") {
		_, err = fmt.Fprintln(output, zshCompletionScript)
	} else {
		err = fmt.Errorf("nothing to generate for: %s", strings.Join(args, ", "))
	}
	return
}

func sortedProfileKeys(data map[string]*config.Profile) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// containsString returns true if arg is contained in args.
func containsString(args []string, arg string) bool {
	for _, a := range args {
		if a == arg {
			return true
		}
	}
	return false
}

func showProfile(output io.Writer, c *config.Config, flags commandLineFlags, args []string) error {
	defer c.DisplayConfigurationIssues()

	// Show global section first
	global, err := c.GetGlobalSection()
	if err != nil {
		return fmt.Errorf("cannot load global section: %w", err)
	}
	err = config.ShowStruct(output, global, constants.SectionConfigurationGlobal)
	if err != nil {
		return fmt.Errorf("cannot show global section: %w", err)
	}
	fmt.Fprintln(output, "")

	// Show profile
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return fmt.Errorf("profile '%s' not found", flags.name)
		}
		return fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
	}

	// Display deprecation notice
	displayProfileDeprecationNotices(profile)

	err = config.ShowStruct(output, profile, "profile "+flags.name)
	if err != nil {
		return fmt.Errorf("cannot show profile '%s': %w", flags.name, err)
	}
	fmt.Fprintln(output, "")

	showSchedules(output, profile.Schedules())
	return nil
}

func showSchedules(output io.Writer, schedulesConfig []*config.ScheduleConfig) {
	for _, schedule := range schedulesConfig {
		export := schedule.Export()
		err := config.ShowStruct(output, export, "schedule "+export.Profiles[0]+"-"+export.Command)
		if err != nil {
			fmt.Fprintln(output, err)
		}
		fmt.Fprintln(output, "")
	}
}

// randomKey simply display a base64'd random key to the console
func randomKey(output io.Writer, c *config.Config, flags commandLineFlags, args []string) error {
	var err error
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
	if containsString(args, "--all") {
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
func createSchedule(_ io.Writer, c *config.Config, flags commandLineFlags, args []string) error {
	defer c.DisplayConfigurationIssues()

	type profileJobs struct {
		scheduler schedule.SchedulerConfig
		profile   string
		jobs      []*config.ScheduleConfig
	}

	allJobs := make([]profileJobs, 0, 1)

	// Step 1: Collect all jobs of all selected profiles
	for _, profileName := range selectProfiles(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)

		scheduler, profile, jobs, err := getScheduleJobs(c, profileFlags)
		if err == nil {
			err = requireScheduleJobs(jobs, profileFlags)

			// Skip profile with no schedules when "--all" option is set.
			if err != nil && containsString(args, "--all") {
				continue
			}
		}
		if err != nil {
			return err
		}

		displayProfileDeprecationNotices(profile)

		// add the no-start flag to all the jobs
		if containsString(args, "--no-start") {
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

func removeSchedule(_ io.Writer, c *config.Config, flags commandLineFlags, args []string) error {
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

func statusSchedule(w io.Writer, c *config.Config, flags commandLineFlags, args []string) error {
	defer c.DisplayConfigurationIssues()

	if !containsString(args, "--all") {
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

func statusScheduleProfile(scheduler schedule.SchedulerConfig, profile *config.Profile, schedules []*config.ScheduleConfig, flags commandLineFlags) error {
	displayProfileDeprecationNotices(profile)

	err := statusJobs(schedule.NewHandler(scheduler), flags.name, schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func getScheduleJobs(c *config.Config, flags commandLineFlags) (schedule.SchedulerConfig, *config.Profile, []*config.ScheduleConfig, error) {
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

func requireScheduleJobs(schedules []*config.ScheduleConfig, flags commandLineFlags) error {
	if len(schedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", flags.name)
	}
	return nil
}

func getRemovableScheduleJobs(c *config.Config, flags commandLineFlags) (schedule.SchedulerConfig, *config.Profile, []*config.ScheduleConfig, error) {
	scheduler, profile, schedules, err := getScheduleJobs(c, flags)
	if err != nil {
		return nil, nil, nil, err
	}

	// Add all undeclared schedules as remove-only configs
	for _, command := range profile.SchedulableCommands() {
		declared := false
		for _, s := range schedules {
			if declared = s.SubTitle == command; declared {
				break
			}
		}
		if !declared {
			schedules = append(schedules, config.NewRemoveOnlyConfig(profile.Name, command))
		}
	}

	return scheduler, profile, schedules, nil
}

func testElevationCommand(_ io.Writer, c *config.Config, flags commandLineFlags, args []string) error {
	if flags.isChild {
		client := remote.NewClient(flags.parentPort)
		term.Print("first line", "\n")
		term.Println("second", "one")
		term.Printf("value = %d", 11)
		err := client.Done()
		if err != nil {
			return err
		}
		return nil
	}

	return elevated(flags)
}

func retryElevated(err error, flags commandLineFlags) error {
	if err == nil {
		return nil
	}
	// maybe can find a better way than searching for the word "denied"?
	if runtime.GOOS == "windows" && !flags.isChild && strings.Contains(err.Error(), "denied") {
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
	if runtime.GOOS != "windows" {
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
