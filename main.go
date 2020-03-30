package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/viper"
)

func main() {
	loadConfiguration()
	testRestic()
}

func loadConfiguration() {
	viper.SetConfigType("toml")
	viper.SetConfigName("profiles.conf")
	viper.AddConfigPath("./examples")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(err)
	}
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
