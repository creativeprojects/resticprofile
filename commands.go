package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/systemd"
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
			name:              "profiles",
			description:       "display profile names from the configuration file",
			action:            displayProfilesCommand,
			needConfiguration: true,
		},
		{
			name:              "self-update",
			description:       "update resticprofile to latest version (does not update restic)",
			action:            selfUpdate,
			needConfiguration: false,
		},
		{
			name:              "systemd-unit",
			description:       "create a user systemd timer",
			action:            createSystemdTimer,
			needConfiguration: true,
		},
		{
			name:              "show",
			description:       "show all the details of the current profile",
			action:            showProfile,
			needConfiguration: true,
			hide:              false,
		},
		{
			name:              "panic",
			description:       "(debug only) simulates a panic",
			action:            panicCommand,
			needConfiguration: false,
			hide:              true,
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

func displayProfiles(configuration *config.Config) {
	profileSections := configuration.GetProfileSections()
	keys := sortedMapKeys(profileSections)
	if profileSections == nil || len(profileSections) == 0 {
		fmt.Println("\nThere's no available profile in the configuration")
	} else {
		fmt.Println("\nProfiles available:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, name := range keys {
			sections := profileSections[name]
			sort.Strings(sections)
			if sections == nil || len(sections) == 0 {
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
	if groups == nil || len(groups) == 0 {
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

func createSystemdTimer(_ *config.Config, flags commandLineFlags, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("OnCalendar argument required")
	}
	systemd.Generate(flags.name, args[0])
	return nil
}

func panicCommand(_ *config.Config, _ commandLineFlags, _ []string) error {
	panic("you asked for it")
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
	fmt.Printf("\n%s:\n", flags.name)
	config.ShowStruct(os.Stdout, profile)
	return nil
}

func createSchedule(c *config.Config, flags commandLineFlags, args []string) error {
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		return fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
	}
	if profile == nil {
		return fmt.Errorf("profile '%s' not found", flags.name)
	}

	job := schedule.NewJob(flags.config, profile)
	err = job.Create()
	if err == nil {
		clog.Info("scheduled job created")
	}
	return err
}

func removeSchedule(c *config.Config, flags commandLineFlags, args []string) error {
	err := schedule.RemoveJob(flags.name)
	if err == nil {
		clog.Info("scheduled job removed")
	}
	return err
}
