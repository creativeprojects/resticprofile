package main

import (
	"flag"
	"os"
	"runtime"

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

	flags := loadFlags()
	if flags.help {
		flag.Usage()
		return
	}
	clog.SetLevel(flags.quiet, flags.verbose)
	if flags.theme != "" {
		clog.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		clog.Colorize(false)
	}

	banner()

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

	resticCommand := newCommand(resticBinary, flag.Args(), nil)
	err = runCommand(resticCommand)
	if err != nil {
		clog.Error(err)
		os.Exit(1)
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
