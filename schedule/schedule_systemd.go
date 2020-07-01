//+build !darwin,!windows

package schedule

import "github.com/creativeprojects/resticprofile/config"

func CreateJob(configFile string, profile *config.Profile) error {
	return nil
}

func RemoveJob(profileName string) error {
	return nil
}
