package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/creativeprojects/resticprofile/lock"
	"github.com/creativeprojects/resticprofile/priority"
	"github.com/spf13/viper"
)

// These fields are populated by the goreleaser build
var (
	version = "0.6.1"
	commit  = ""
	date    = ""
	builtBy = ""
)

func init() {
	rand.Seed(time.Now().UnixNano() - time.Now().Unix())
}

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

	if flags.saveConfigAs != "" {
		err = config.SaveAs(flags.saveConfigAs)
		if err != nil {
			clog.Error("Cannot save configuration file:", err)
			os.Exit(1)
		}
		return
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

	if global.IONice {
		err = priority.SetIONice(global.IONiceClass, global.IONiceLevel)
		if err != nil {
			clog.Warning(err)
		}
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

	if config.HasProfile(flags.name) {
		// Single profile run
		runProfile(global, flags, flags.name, resticBinary, resticArguments, resticCommand)

	} else if config.HasGroup(flags.name) {
		// Group run
		group, err := config.LoadGroup(flags.name)
		if err != nil {
			clog.Errorf("Cannot load group '%s': %v", flags.name, err)
		}
		if group != nil && len(group) > 0 {
			for i, profileName := range group {
				clog.Debugf("[%d/%d] Starting profile '%s' from group '%s'", i+1, len(group), profileName, flags.name)
				runProfile(global, flags, profileName, resticBinary, resticArguments, resticCommand)
			}
		}

	} else {
		clog.Errorf("Profile or group not found '%s'", flags.name)
		displayProfiles()
		displayGroups()
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
		coin := ""
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
		}
		clog.Warningf("You specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! and selection is ... %s.", coin)
	}
	if flags.quiet {
		clog.Quiet()
	}
	if flags.verbose {
		clog.Verbose()
	}
}

func banner() {
	clog.Infof("resticprofile %s compiled with %s", version, runtime.Version())
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
			return fmt.Errorf("incorrect value '%s' for priority in global section", class)
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

func displayGroups() {
	groups := config.ProfileGroups()
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

func runProfile(global *config.Global, flags commandLineFlags, profileName string, resticBinary string, resticArguments []string, resticCommand string) {
	var err error

	profile, err := config.LoadProfile(profileName)
	if err != nil {
		clog.Warning(err)
	}
	if profile == nil {
		clog.Errorf("Profile '%s' not found", profileName)
		os.Exit(1)
	}

	// Send the quiet/verbose down to restic as well (override profile configuration)
	if flags.quiet {
		profile.Quiet = true
		profile.Verbose = false
	}
	if flags.verbose {
		profile.Verbose = true
		profile.Quiet = false
	}

	// All files in the configuration are relative to the configuration file, NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(viper.ConfigFileUsed())
	if rootPath != "." {
		clog.Debugf("Files in configuration are relative to '%s'", rootPath)
	}
	profile.SetRootPath(rootPath)

	// Specific case for the "host" flag where an empty value should be replaced by the hostname
	hostname := "none"
	currentHost, err := os.Hostname()
	if err == nil {
		hostname = currentHost
	}
	profile.SetHost(hostname)

	// Catch CTR-C keypress
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	wrapper := newResticWrapper(resticBinary, profile, resticArguments, sigChan)
	if (global.Initialize || profile.Initialize) && resticCommand != constants.CommandInit {
		_ = wrapper.runInitialize()
		// it's ok for the initialize to error out when the repository exists
	}

	err = lockRun(profile.Lock, func() error {
		var err error

		// pre-commands (for backup)
		if resticCommand == constants.CommandBackup {
			// Shell commands
			err = wrapper.runPreCommand(resticCommand)
			if err != nil {
				return err
			}
			// Check
			if profile.Backup != nil && profile.Backup.CheckBefore {
				err = wrapper.runCheck()
				if err != nil {
					return err
				}
			}
			// Retention
			if profile.Retention != nil && profile.Retention.BeforeBackup {
				err = wrapper.runRetention()
				if err != nil {
					return err
				}
			}
		}

		// Main command
		err = wrapper.runCommand(resticCommand)
		if err != nil {
			return err
		}

		// post-commands (for backup)
		if resticCommand == constants.CommandBackup {
			// Retention
			if profile.Retention != nil && profile.Retention.AfterBackup {
				err = wrapper.runRetention()
				if err != nil {
					return err
				}
			}
			// Check
			if profile.Backup != nil && profile.Backup.CheckAfter {
				err = wrapper.runCheck()
				if err != nil {
					return err
				}
			}
			// Shell commands
			err = wrapper.runPostCommand(resticCommand)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		clog.Error(err)
		os.Exit(1)
	}
}

func lockRun(filename string, run func() error) error {
	if filename == "" {
		// No lock
		return run()
	}
	// Make sure the path to the lock exists
	dir := filepath.Dir(filename)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			clog.Warningf("The profile will run without a lockfile: %v", err)
			return run()
		}
	}
	runLock := lock.NewLock(filename)
	if !runLock.TryAcquire() {
		return fmt.Errorf("another process is already running this profile: %s", runLock.Who())
	}
	defer runLock.Release()
	return run()
}

// randomBool returns true for Heads and false for Tails
func randomBool() bool {
	return rand.Int31n(10000) < 5000
}
