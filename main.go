package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"

	"github.com/creativeprojects/resticprofile/constants"

	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/creativeprojects/resticprofile/priority"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
)

const (
	resticProfileVersion = "0.6.0"
)

func main() {
	var err error

	flagset, flags := loadFlags()
	if flags.help {
		flagset.Usage()
		return
	}
	setLoggerFlags(flags)
	banner()

	if flags.selfUpdate {
		err = confirmAndSelfUpdate(flags.verbose)
		if err != nil {
			clog.Error(err)
			os.Exit(1)
		}
		return
	}

	configFile, err := filesearch.FindConfigurationFile(flags.config)
	if err != nil {
		clog.Error(err)
		os.Exit(1)
	}

	err = config.LoadConfiguration(configFile)
	if err != nil {
		clog.Error("Cannot load configuration file:", err)
		os.Exit(1)
	}
	global, err := config.GetGlobalSection()
	if err != nil {
		clog.Error("Cannot load global configuration:", err)
		os.Exit(1)
	}

	err = setPriority(global.Nice, global.Priority)
	if err != nil {
		clog.Warning(err)
	}

	resticBinary, err := filesearch.FindResticBinary(global.ResticBinary)
	if err != nil {
		clog.Error("Cannot find restic:", err)
		clog.Warning("You can specify the path of the restic binary in the global section of the configuration file (restic-binary)")
		os.Exit(1)
	}

	// The remaining arguments are sent to the restic command line
	resticArguments := flags.resticArgs
	resticCommand := global.DefaultCommand
	if len(resticArguments) > 0 {
		resticCommand = resticArguments[0]
		resticArguments = resticArguments[1:]
	}

	if resticCommand == constants.CommandProfiles {
		displayProfiles()
		displayGroups()
		return
	}

	profile, err := config.LoadProfile(flags.name)
	if err != nil {
		clog.Warning(err)
	}
	if profile == nil {
		clog.Errorf("Profile '%s' not found", flags.name)
		os.Exit(1)
	}

	// All files in the configuration are relative to the configuration file, NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(viper.ConfigFileUsed())
	clog.Debugf("File in configuration are relative to '%s'", rootPath)
	profile.SetRootPath(rootPath)

	// resticBinary = "/Users/gouarfig/go/bin/showenv"
	wrapper := newResticWrapper(resticBinary, profile, resticArguments)
	if (global.Initialize || profile.Initialize) && resticCommand != constants.CommandInit {
		wrapper.runInitialize()
		// it's ok for the initialize to error out when the repository exists
	}
	err = wrapper.runCommand(resticCommand)
	if err != nil {
		clog.Error(err)
		os.Exit(1)
	}
}

func setLoggerFlags(flags commandLineFlags) {
	if flags.theme != "" {
		clog.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		clog.Colorize(false)
	}

	if flags.quiet && flags.verbose {
		clog.Warning("You specified -quiet (-q) and -verbose (-v) at the same time. Selection is verbose.")
		flags.quiet = false
	}
	if flags.quiet {
		clog.Quiet()
	}
	if flags.verbose {
		clog.Verbose()
	}
}

func banner() {
	clog.Infof("resticprofile %s compiled with %s", resticProfileVersion, runtime.Version())
}

func setPriority(nice int, class string) error {
	var err error

	if class != "" {
		if classID, ok := constants.PriorityValues[class]; ok {
			err = priority.SetClass(classID)
			if err != nil {
				return err
			}
		} else {
			clog.Warningf("Incorrect value '%s' for priority in global section", class)
		}
		return nil
	}
	if nice != 0 {
		err = priority.SetNice(nice)
		if err != nil {
			return err
		}
	}
	return nil
}

func displayProfiles() {
	profileSections := config.ProfileSections()
	if profileSections == nil || len(profileSections) == 0 {
		fmt.Println("\nThere's no available profile in the configuration")
	} else {
		fmt.Println("\nProfiles available:")
		for name, sections := range profileSections {
			if sections == nil || len(sections) == 0 {
				fmt.Printf("\t%s\n", name)
			} else {
				fmt.Printf("\t%s\t(%s)\n", name, strings.Join(sections, ", "))
			}
		}
	}
	fmt.Println("")
}

func displayGroups() {
	groups := config.ProfileGroups()
	if groups == nil || len(groups) == 0 {
		return
	}
	fmt.Println("Groups available:")
	for name, groupList := range groups {
		fmt.Printf("\t%s: %s\n", name, strings.Join(groupList, ", "))
	}
	fmt.Println("")
}

func displayStruct(name string, value interface{}) {
	s, _ := json.MarshalIndent(value, "", "\t")
	fmt.Printf("%s:\n%s\n\n", name, s)
}
