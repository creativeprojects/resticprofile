package main

import (
	"github.com/spf13/viper"
)

func main() {
	loadConfiguration()
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
