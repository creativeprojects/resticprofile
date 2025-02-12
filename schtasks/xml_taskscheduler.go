//go:build !taskmaster

// Schedule types on Windows:
// ==========================
// 1. one time:
//   - at a specific date
//
// 2. daily:
//   - 1 start date
//   - recurring every n days
//
// 3. weekly:
//   - 1 start date
//   - recurring every n weeks
//   - on specific weekdays
//
// 4. monthly:
//   - 1 start date
//   - on specific months
//   - on specific days (1 to 31)
package schtasks

import (
	"fmt"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
)

const (
	binaryPath = "schtasks.exe"
)

// Connect checks the schtask.exe tool is available
func Connect() error {
	found, err := exec.LookPath(binaryPath)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like %s is installed on your system", binaryPath)
	}
	return nil
}

// Close does nothing with this implementation
func Close() {
	// nothing to do
}

// Create or update a task (if the name already exists in the Task Scheduler)
func Create(config *Config, schedules []*calendar.Event, permission Permission) error {
	return nil
}

// Delete a task
func Delete(title, subtitle string) error {
	return nil
}

// Status returns the status of a task
func Status(title, subtitle string) error {
	return nil
}

func getTaskPath(profileName, commandName string) string {
	return fmt.Sprintf("%s%s %s", tasksPath, profileName, commandName)
}
