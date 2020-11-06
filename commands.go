package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/win"
)

type ownCommand struct {
	name              string
	description       string
	action            func(*config.Config, commandLineFlags, []string) error
	needConfiguration bool // true if the action needs a configuration file loaded
	hide              bool
}

var (
	ownCommands = []ownCommand{
		{
			name:              "version",
			description:       "display version (run in vebose mode for detailed information)",
			action:            displayVersion,
			needConfiguration: false,
		},
		{
			name:              "self-update",
			description:       "update resticprofile to latest version (does not update restic)",
			action:            selfUpdate,
			needConfiguration: false,
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
		},
		{
			name:              "schedule",
			description:       "schedule a backup",
			action:            createSchedule,
			needConfiguration: true,
			hide:              false,
		},
		{
			name:              "unschedule",
			description:       "remove a scheduled backup",
			action:            removeSchedule,
			needConfiguration: true,
			hide:              false,
		},
		{
			name:              "status",
			description:       "display the status of a scheduled backup job",
			action:            statusSchedule,
			needConfiguration: true,
			hide:              false,
		},
		// hidden commands
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
)

func displayOwnCommands() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, command := range ownCommands {
		if command.hide {
			continue
		}
		_, _ = fmt.Fprintf(w, "\t%s\t%s\n", command.name, command.description)
	}
	_ = w.Flush()
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
			return commandDef.action(configuration, flags, args)
		}
	}
	return fmt.Errorf("command not found: %v", command)
}

func displayProfilesCommand(configuration *config.Config, _ commandLineFlags, _ []string) error {
	displayProfiles(configuration)
	displayGroups(configuration)
	return nil
}

func displayVersion(_ *config.Config, flags commandLineFlags, _ []string) error {
	fmt.Printf("resticprofile version %s commit %s.\n", version, commit)

	if flags.verbose {
		w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintf(w, "\n")
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "home", "https://github.com/creativeprojects/resticprofile")
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "os", runtime.GOOS)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "arch", runtime.GOARCH)
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

func displayProfiles(configuration *config.Config) {
	profileSections := configuration.GetProfileSections()
	keys := sortedMapKeys(profileSections)
	if len(profileSections) == 0 {
		fmt.Println("\nThere's no available profile in the configuration")
	} else {
		fmt.Println("\nProfiles available:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, name := range keys {
			sections := profileSections[name]
			sort.Strings(sections)
			if len(sections) == 0 {
				_, _ = fmt.Fprintf(w, "\t%s:\t(n/a)\n", name)
			} else {
				_, _ = fmt.Fprintf(w, "\t%s:\t(%s)\n", name, strings.Join(sections, ", "))
			}
		}
		_ = w.Flush()
	}
	fmt.Println("")
}

func displayGroups(configuration *config.Config) {
	groups := configuration.GetProfileGroups()
	if len(groups) == 0 {
		return
	}
	fmt.Println("Groups available:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for name, groupList := range groups {
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", name, strings.Join(groupList, ", "))
	}
	_ = w.Flush()
	fmt.Println("")
}

func selfUpdate(_ *config.Config, flags commandLineFlags, args []string) error {
	err := confirmAndSelfUpdate(flags.verbose)
	if err != nil {
		return err
	}
	return nil
}

func panicCommand(_ *config.Config, _ commandLineFlags, _ []string) error {
	panic("you asked for it")
}

func testCommand(_ *config.Config, _ commandLineFlags, _ []string) error {
	return nil
}

func sortedMapKeys(data map[string][]string) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func showProfile(c *config.Config, flags commandLineFlags, args []string) error {
	// Show global section first
	global, err := c.GetGlobalSection()
	if err != nil {
		return fmt.Errorf("cannot show global: %w", err)
	}
	fmt.Printf("\n%s:\n", constants.SectionConfigurationGlobal)
	config.ShowStruct(os.Stdout, global)

	// Then show profile
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		return fmt.Errorf("cannot show profile '%s': %w", flags.name, err)
	}
	if profile == nil {
		return fmt.Errorf("profile '%s' not found", flags.name)
	}
	// All files in the configuration are relative to the configuration file, NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(c.GetConfigFile())
	if rootPath != "." {
		clog.Debugf("files in configuration are relative to '%s'", rootPath)
	}
	profile.SetRootPath(rootPath)

	fmt.Printf("\n%s:\n", flags.name)
	config.ShowStruct(os.Stdout, profile)
	return nil
}

// randomKey simply display a base64'd random key to the console
func randomKey(c *config.Config, flags commandLineFlags, args []string) error {
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
	encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)
	_, err = encoder.Write(buffer)
	encoder.Close()
	fmt.Println("")
	return err
}

func createSchedule(c *config.Config, flags commandLineFlags, args []string) error {
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		return fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
	}
	if profile == nil {
		return fmt.Errorf("profile '%s' not found", flags.name)
	}

	schedules := profile.Schedules()
	if len(schedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", flags.name)
	}

	err = scheduleJobs(flags.config, schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func removeSchedule(c *config.Config, flags commandLineFlags, args []string) error {
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		return fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
	}
	if profile == nil {
		return fmt.Errorf("profile '%s' not found", flags.name)
	}

	schedules := profile.Schedules()
	if len(schedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", flags.name)
	}

	err = removeJobs(schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func statusSchedule(c *config.Config, flags commandLineFlags, args []string) error {
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		return fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
	}
	if profile == nil {
		return fmt.Errorf("profile '%s' not found", flags.name)
	}

	schedules := profile.Schedules()
	if len(schedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", flags.name)
	}

	err = statusJobs(schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func testElevationCommand(c *config.Config, flags commandLineFlags, args []string) error {
	if flags.isChild {
		client := remote.NewClient(flags.parentPort)
		term.Print("first line", "\n")
		term.Println("second", "one")
		term.Printf("value = %d", 11)
		client.Done()
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
