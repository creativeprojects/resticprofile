package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/spf13/viper"
)

type ownCommand struct {
	name              string
	description       string
	action            func(*config.Config, commandLineFlags, []string) error
	needConfiguration bool // true if the action needs a configuration file loaded
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
			description:       "(debug only) show all keys from the configuration file",
			action:            allKeys,
			needConfiguration: true,
		},
	}
)

func displayOwnCommands() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, command := range ownCommands {
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
	profileSections := configuration.ProfileSections()
	if profileSections == nil || len(profileSections) == 0 {
		fmt.Println("\nThere's no available profile in the configuration")
	} else {
		fmt.Println("\nProfiles available:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for name, sections := range profileSections {
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
	groups := configuration.ProfileGroups()
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

func allKeys(_ *config.Config, flags commandLineFlags, args []string) error {
	keys := viper.AllKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, key := range keys {
		_, _ = fmt.Fprintf(w, "\t%s\t%+v\n", key, viper.Get(key))
	}
	_ = w.Flush()
	return nil
}
