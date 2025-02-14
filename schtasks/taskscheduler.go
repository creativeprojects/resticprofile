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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/term"
)

const (
	tasksPath  = `\resticprofile backup\`
	binaryPath = "schtasks.exe"
	// From: https://learn.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-string-format
	// O:owner_sid
	// G:group_sid
	// D:dacl_flags(string_ace1)(string_ace2)... (string_acen)  <---
	// S:sacl_flags(string_ace1)(string_ace2)... (string_acen)
	// With flag:
	// "AI"	SDDL_AUTO_INHERITED
	// From: https://learn.microsoft.com/en-us/windows/win32/secauthz/ace-strings
	// - first field:
	// "A" 	SDDL_ACCESS_ALLOWED
	// - third field
	// "FA" 	SDDL_FILE_ALL 	FILE_GENERIC_ALL
	// "FR" 	SDDL_FILE_READ 	FILE_GENERIC_READ
	// "FW" 	SDDL_FILE_WRITE 	FILE_GENERIC_WRITE
	// "FX" 	SDDL_FILE_EXECUTE 	FILE_GENERIC_EXECUTE
	// From: https://learn.microsoft.com/en-us/windows/win32/secauthz/sid-strings
	// "AU" 	SDDL_AUTHENTICATED_USERS
	// "BA" 	SDDL_BUILTIN_ADMINISTRATORS
	// "LS" 	SDDL_LOCAL_SERVICE
	// "SY" 	SDDL_LOCAL_SYSTEM
	securityDescriptor = "D:AI(A;;FA;;;BA)(A;;FA;;;SY)(A;;FRFX;;;LS)(A;;FR;;;AU)"
)

// Init only checks the schtask.exe tool is available
func Init() error {
	found, err := exec.LookPath(binaryPath)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like %s is installed on your system", binaryPath)
	}
	return nil
}

// Create or update a task (if the name already exists in the Task Scheduler)
func Create(config *Config, schedules []*calendar.Event, permission Permission) error {
	var (
		// username, password string
		err error
	)
	task := createTaskDefinition(config, schedules)
	taskPath := getTaskPath(config.ProfileName, config.CommandName)
	task.RegistrationInfo.URI = taskPath

	switch permission {
	case SystemAccount:
		task.Principals.Principal.LogonType = LogonTypeServiceForUser
		task.Principals.Principal.RunLevel = RunLevelHighest
		task.Principals.Principal.UserId = serviceAccount
		task.RegistrationInfo.SecurityDescriptor = securityDescriptor // allow authenticated users to read the task status

	case UserLoggedOnAccount:
		task.Principals.Principal.LogonType = LogonTypeInteractiveToken

	default:
		task.Principals.Principal.LogonType = LogonTypePassword
		// username, password, err = userCredentials()
		// if err != nil {
		// 	return fmt.Errorf("cannot get user name or password: %w", err)
		// }
	}

	file, err := os.CreateTemp("", "*.xml")
	if err != nil {
		return fmt.Errorf("cannot create XML task file: %w", err)
	}
	filename := file.Name()
	defer func() {
		file.Close()
		os.Remove(filename)
	}()

	err = createTaskFile(task, file)
	if err != nil {
		return fmt.Errorf("cannot write XML task file: %w", err)
	}
	file.Close()

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/create", "/tn", taskPath, "/xml", filename)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		return schTasksError(stderr.String())
	}
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

func Registered() ([]Config, error) {
	raw, err := listRegisteredTasks()
	if err != nil {
		return nil, err
	}
	all, err := getCSV(bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	for _, taskLine := range all {
		if strings.HasPrefix(taskLine[0], tasksPath) {
			list = append(list, taskLine[0])
		}
	}
	slices.Sort(list)
	list = slices.Compact(list)

	configs := make([]Config, 0, len(list))
	for _, taskPath := range list {
		info, err := getTaskInfo(taskPath)
		if err != nil {
			clog.Errorf("loading task %q: %w", taskPath, err)
			continue
		}
		if len(info) < 2 {
			continue
		}
		taskName := strings.TrimPrefix(taskPath, tasksPath)
		parts := strings.Split(taskName, " ")
		if len(parts) < 2 {
			clog.Warningf("cannot parse task path: %s", taskPath)
			continue
		}
		profileName := strings.Join(parts[:len(parts)-1], " ")
		commandName := parts[len(parts)-1]
		commandLine := getFirstField(info, "Task To Run")
		command, args, _ := strings.Cut(commandLine, " ")
		config := Config{
			ProfileName:      profileName,
			CommandName:      commandName,
			JobDescription:   getFirstField(info, "Comment"),
			WorkingDirectory: getFirstField(info, "Start In"),
			Command:          command,
			Arguments:        args,
		}
		configs = append(configs, config)
	}
	return configs, nil
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
