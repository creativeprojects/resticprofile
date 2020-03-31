package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/creativeprojects/resticprofile/filesearch"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
)

func main() {
	var err error

	flags := loadFlags()
	if flags.help {
		flag.Usage()
		return
	}
	clog.SetLevel(flags.quiet, flags.verbose)

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

	resticBinary, err := filesearch.FindResticBinary(global.ResticBinary)
	if err != nil {
		clog.Error("Cannot find restic:", err)
		clog.Warning("You can specify the path of the restic binary in the global section of the configuration file (restic-binary)")
		os.Exit(1)
	}
	fmt.Println(resticBinary)

	// fmt.Println(flags)
	// fmt.Println(flag.Args())

}

func testRestic() {
	path, err := exec.LookPath("restic")
	if err != nil {
		log.Fatal("restic is not available on your system")
	}
	fmt.Printf("restic is available at %s\n", path)

	commands := []commandDefinition{
		newCommand("restic", []string{"version"}, nil),
		newCommand("restic", []string{"snapshots"}, []string{"RESTIC_REPOSITORY=/tmp"}),
	}
	err = runCommands(commands)
	if err != nil {
		log.Fatal(err)
	}
}

func testNice() {
	command := newCommand("go", []string{"run", "./priority/check"}, nil)
	err := runCommand(command)
	if err != nil {
		log.Fatal(err)
	}
}
