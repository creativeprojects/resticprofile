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
	"slices"
	"text/tabwriter"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/term"
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
	taskName := getTaskPath(title, subtitle)
	_, err := deleteTask(taskName)
	return err
}

// Status returns the status of a task
func Status(title, subtitle string) error {
	taskName := getTaskPath(title, subtitle)
	info, err := getTaskInfo(taskName)
	if err != nil {
		return err
	}
	if len(info) < 2 {
		return ErrNotRegistered
	}
	writer := tabwriter.NewWriter(term.GetOutput(), 2, 2, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(writer, "Task:\t %s\n", getFirstField(info, "TaskName"))
	fmt.Fprintf(writer, "User:\t %s\n", getFirstField(info, "Run As User"))
	fmt.Fprintf(writer, "Working Dir:\t %v\n", getFirstField(info, "Start In"))
	fmt.Fprintf(writer, "Command:\t %v\n", getFirstField(info, "Task To Run"))
	fmt.Fprintf(writer, "Status:\t %s\n", getFirstField(info, "Status"))
	fmt.Fprintf(writer, "Last Run Time:\t %v\n", getFirstField(info, "Last Run Time"))
	fmt.Fprintf(writer, "Last Result:\t %s\n", getFirstField(info, "Last Result"))
	fmt.Fprintf(writer, "Next Run Time:\t %v\n", getFirstField(info, "Next Run Time"))
	writer.Flush()
	return nil
}

func getTaskPath(profileName, commandName string) string {
	return fmt.Sprintf("%s%s %s", tasksPath, profileName, commandName)
}

func getFirstField(data [][]string, fieldName string) string {
	index := slices.Index(data[0], fieldName)
	if index < 0 {
		return ""
	}
	return data[1][index]
}
