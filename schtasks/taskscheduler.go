//+build windows

package schtasks

import (
	"errors"
	"fmt"
	"math"
	"os/user"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/creativeprojects/clog"

	"github.com/capnspacehook/taskmaster"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/rickb777/date/period"
)

// Schedule types on Windows:
// ==========================
// 1. one time:
//    - at a specific date
// 2. daily:
//    - 1 start date
//    - recurring every n days
// 3. weekly:
//    - 1 start date
//    - recurring every n weeks
//    - on specific weekdays
// 4. monthly:
//    - 1 start date
//    - on specific months
//    - on specific days (1 to 31)

const (
	tasksPath = `\resticprofile backup\`
)

// Permission is a choice between User and System
type Permission int

// Permission available
const (
	UserAccount Permission = iota
	SystemAccount
)

// Config contains all the information needed to schedule a Job
type Config interface {
	Title() string
	SubTitle() string
	JobDescription() string
	TimerDescription() string
	Schedules() []string
	Permission() string
	WorkingDirectory() string
	Command() string
	Arguments() []string
}

var (
	// no need to recreate the service every time
	taskService *taskmaster.TaskService
	// current user
	userName = ""
	// ask the user password only once
	userPassword = ""
)

// Connect initializes a connection to the local task scheduler
func Connect() error {
	var err error

	if taskService == nil || !taskService.IsConnected() {
		taskService, err = taskmaster.Connect("", "", "", "")
	}
	return err
}

// Close releases the ressources used by the task service
func Close() {
	if taskService == nil {
		return
	}
	taskService.Disconnect()
	taskService = nil
}

// Create or update a task (if the name already exists in the Task Scheduler)
func Create(config Config, schedules []*calendar.Event, permission Permission) error {
	if permission == SystemAccount {
		return createSystemTask(config, schedules)
	}
	return createUserTask(config, schedules)
}

// createUserTask creates a new user task. Will update an existing task instead of overwritting
func createUserTask(config Config, schedules []*calendar.Event) error {
	taskName := getTaskPath(config.Title(), config.SubTitle())
	registeredTask, err := taskService.GetRegisteredTask(taskName)
	if err != nil {
		return err
	}
	if registeredTask != nil {
		return updateUserTask(registeredTask, config, schedules)
	}

	username, password, err := userCredentials()
	if err != nil {
		return fmt.Errorf("cannot get user name or password: %w", err)
	}

	task := taskService.NewTaskDefinition()

	task.AddExecAction(
		config.Command(),
		strings.Join(config.Arguments(), " "),
		config.WorkingDirectory(),
		"")
	task.Principal.LogonType = taskmaster.TASK_LOGON_PASSWORD
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_LUA
	task.Principal.UserID = username
	task.RegistrationInfo.Author = "resticprofile"
	task.RegistrationInfo.Description = config.JobDescription()

	createSchedules(&task, schedules)

	_, created, err := taskService.CreateTaskEx(
		taskName,
		task,
		username,
		password,
		taskmaster.TASK_LOGON_PASSWORD,
		false)
	if err != nil {
		return err
	}
	if !created {
		return errors.New("cannot create user task")
	}
	return nil
}

// updateUserTask updates an existing task
func updateUserTask(task *taskmaster.RegisteredTask, config Config, schedules []*calendar.Event) error {
	taskName := getTaskPath(config.Title(), config.SubTitle())

	username, password, err := userCredentials()
	if err != nil {
		return fmt.Errorf("cannot get user name or password: %w", err)
	}

	// clear up all actions and put ours back
	task.Definition.Actions = make([]taskmaster.Action, 0, 1)
	task.Definition.AddExecAction(
		config.Command(),
		strings.Join(config.Arguments(), " "),
		config.WorkingDirectory(),
		"")
	task.Definition.Principal.LogonType = taskmaster.TASK_LOGON_PASSWORD
	task.Definition.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_LUA
	task.Definition.Principal.UserID = username

	// clear up all schedules and put them back
	task.Definition.Triggers = []taskmaster.Trigger{}
	createSchedules(&task.Definition, schedules)

	_, err = taskService.UpdateTaskEx(
		taskName,
		task.Definition,
		username,
		password,
		taskmaster.TASK_LOGON_PASSWORD)
	if err != nil {
		return err
	}
	return nil
}

// userCredentials asks for the user password only once, and keeps it in cache
func userCredentials() (string, string, error) {
	if userName != "" {
		// we've been here already: we don't check for blank password as it's a valid password
		return userName, userPassword, nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return "", "", err
	}
	userName = currentUser.Username

	fmt.Printf("\nCreating task for user %s\n", userName)
	fmt.Printf("Task Scheduler requires your Windows password to validate the task: ")
	userPassword, err = term.ReadPassword()
	if err != nil {
		return "", "", err
	}
	return userName, userPassword, nil
}

// createSystemTask creates a new system task. Will update an existing task instead of overwritting
func createSystemTask(config Config, schedules []*calendar.Event) error {
	taskName := getTaskPath(config.Title(), config.SubTitle())
	registeredTask, err := taskService.GetRegisteredTask(taskName)
	if err != nil {
		return err
	}
	if registeredTask != nil {
		return updateSystemTask(registeredTask, config, schedules)
	}

	task := taskService.NewTaskDefinition()
	task.AddExecAction(
		config.Command(),
		strings.Join(config.Arguments(), " "),
		config.WorkingDirectory(),
		"")
	task.Principal.LogonType = taskmaster.TASK_LOGON_SERVICE_ACCOUNT
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_HIGHEST
	task.Principal.UserID = "SYSTEM"
	task.RegistrationInfo.Author = "resticprofile"
	task.RegistrationInfo.Description = config.JobDescription()

	createSchedules(&task, schedules)

	_, created, err := taskService.CreateTask(taskName, task, false)
	if err != nil {
		return err
	}
	if !created {
		return errors.New("cannot create system task")
	}
	return nil
}

// updateSystemTask updates an existing task
func updateSystemTask(task *taskmaster.RegisteredTask, config Config, schedules []*calendar.Event) error {
	taskName := getTaskPath(config.Title(), config.SubTitle())

	// clear up all actions and put ours back
	task.Definition.Actions = make([]taskmaster.Action, 0, 1)
	task.Definition.AddExecAction(
		config.Command(),
		strings.Join(config.Arguments(), " "),
		config.WorkingDirectory(),
		"")
	task.Definition.Principal.LogonType = taskmaster.TASK_LOGON_SERVICE_ACCOUNT
	task.Definition.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_HIGHEST
	task.Definition.Principal.UserID = "SYSTEM"

	// clear up all schedules and put them back
	task.Definition.Triggers = []taskmaster.Trigger{}
	createSchedules(&task.Definition, schedules)

	_, err := taskService.UpdateTask(taskName, task.Definition)
	if err != nil {
		return err
	}
	return nil
}

func createSchedules(task *taskmaster.Definition, schedules []*calendar.Event) {
	for _, schedule := range schedules {
		if once, ok := schedule.AsTime(); ok {
			// one time only
			task.AddTimeTrigger(period.Period{}, once)
			continue
		}
		if schedule.IsDaily() {
			// recurring daily
			createDailyTrigger(task, schedule)
			continue
		}
		if schedule.IsWeekly() {
			createWeeklyTrigger(task, schedule)
			continue
		}
		if schedule.IsMonthly() {
			createMonthlyTrigger(task, schedule)
			continue
		}
		clog.Warningf("cannot convert schedule '%s' into a task scheduler equivalent", schedule.String())
	}
}

func createDailyTrigger(task *taskmaster.Definition, schedule *calendar.Event) {
	emptyPeriod := period.Period{}
	start := schedule.Next(time.Now())
	// get all recurrences in the same day
	recurrences := schedule.GetAllInBetween(start, start.Add(24*time.Hour))
	if len(recurrences) == 0 {
		clog.Warningf("cannot convert schedule '%s' into a daily trigger", schedule.String())
		return
	}
	// Is it only once a day?
	if len(recurrences) == 1 {
		task.AddDailyTrigger(1, emptyPeriod, recurrences[0])
		return
	}
	// now calculate the difference in between each
	differences := make([]time.Duration, len(recurrences)-1)
	for i := 0; i < len(recurrences)-1; i++ {
		differences[i] = recurrences[i+1].Sub(recurrences[i])
	}
	// check if they're all the same
	compactDifferences := make([]time.Duration, 0, len(differences))
	var previous time.Duration = 0
	for _, difference := range differences {
		if difference.Seconds() != previous.Seconds() {
			compactDifferences = append(compactDifferences, difference)
			previous = difference
		}
	}

	if len(compactDifferences) == 1 {
		// easy case
		interval, _ := period.NewOf(compactDifferences[0])
		task.AddDailyTriggerEx(
			1,
			emptyPeriod,
			"",
			start,
			time.Time{},
			emptyPeriod,
			period.NewYMD(0, 0, 1),
			interval,
			false,
			true)
		return
	}
	clog.Warning("createDailyTrigger fallback")
}

func createWeeklyTrigger(task *taskmaster.Definition, schedule *calendar.Event) {
	emptyPeriod := period.Period{}
	start := schedule.Next(time.Now())
	// get all recurrences in the same day
	recurrences := schedule.GetAllInBetween(start, start.Add(24*time.Hour))
	if len(recurrences) == 0 {
		clog.Warningf("cannot convert schedule '%s' into a weekly trigger", schedule.String())
		return
	}
	// Is it only once per 24h?
	if len(recurrences) == 1 {
		task.AddWeeklyTrigger(
			taskmaster.Day(convertWeekdaysToBitmap(schedule.WeekDay.GetRangeValues())),
			1, emptyPeriod, recurrences[0])
		return
	}
	// now calculate the difference in between each
	differences := make([]time.Duration, len(recurrences)-1)
	for i := 0; i < len(recurrences)-1; i++ {
		differences[i] = recurrences[i+1].Sub(recurrences[i])
	}
	// check if they're all the same
	compactDifferences := make([]time.Duration, 0, len(differences))
	var previous time.Duration = 0
	for _, difference := range differences {
		if difference.Seconds() != previous.Seconds() {
			compactDifferences = append(compactDifferences, difference)
			previous = difference
		}
	}

	if len(compactDifferences) == 1 {
		// easy case
		interval, _ := period.NewOf(compactDifferences[0])
		task.AddWeeklyTriggerEx(
			taskmaster.Day(convertWeekdaysToBitmap(schedule.WeekDay.GetRangeValues())),
			1,
			emptyPeriod,
			"",
			start,
			time.Time{},
			emptyPeriod,
			period.NewYMD(0, 0, 1),
			interval,
			false,
			true)
		return
	}
	clog.Warning("createWeeklyTrigger fallback")
}

func createMonthlyTrigger(task *taskmaster.Definition, schedule *calendar.Event) {
	emptyPeriod := period.Period{}
	start := schedule.Next(time.Now())
	// get all recurrences in the same day
	recurrences := schedule.GetAllInBetween(start, start.Add(24*time.Hour))
	if len(recurrences) == 0 {
		clog.Warningf("cannot convert schedule '%s' into a monthly trigger", schedule.String())
		return
	}
	// Is it only once per 24h?
	if len(recurrences) == 1 {
		if schedule.WeekDay.HasValue() && schedule.Day.HasValue() {
			clog.Warningf("task scheduler does not support a day of the month and a day of the week in the same trigger: %s", schedule.String())
			return
		}
		if schedule.WeekDay.HasValue() {
			task.AddMonthlyDOWTrigger(
				taskmaster.Day(convertWeekdaysToBitmap(schedule.WeekDay.GetRangeValues())),
				taskmaster.First|taskmaster.Second|taskmaster.Third|taskmaster.Fourth|taskmaster.Last,
				taskmaster.Month(convertMonthsToBitmap(schedule.Month.GetRangeValues())),
				true,
				emptyPeriod,
				recurrences[0],
			)
			return
		}
		task.AddMonthlyTrigger(
			convertDaysToBitmap(schedule.Day.GetRangeValues()),
			taskmaster.Month(convertMonthsToBitmap(schedule.Month.GetRangeValues())),
			emptyPeriod,
			recurrences[0],
		)
		return
	}
	clog.Warning("createMonthlyTrigger fallback")
}

// Delete a task
func Delete(title, subtitle string) error {
	taskName := getTaskPath(title, subtitle)
	err := taskService.DeleteTask(taskName)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return fmt.Errorf("%w: %s", ErrorNotRegistered, taskName)
		}
		return err
	}
	return nil
}

// Status returns the status of a task
func Status(title, subtitle string) error {
	taskName := getTaskPath(title, subtitle)
	registeredTask, err := taskService.GetRegisteredTask(taskName)
	if err != nil {
		return err
	}
	if registeredTask == nil {
		return fmt.Errorf("%w: %s", ErrorNotRegistered, taskName)
	}
	writer := tabwriter.NewWriter(term.GetOutput(), 2, 2, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(writer, "Task:\t %s\n", registeredTask.Path)
	fmt.Fprintf(writer, "User:\t %s\n", registeredTask.Definition.Principal.UserID)
	if registeredTask.Definition.Actions != nil && len(registeredTask.Definition.Actions) > 0 {
		if action, ok := registeredTask.Definition.Actions[0].(taskmaster.ExecAction); ok {
			fmt.Fprintf(writer, "Working Dir:\t %v\n", action.WorkingDir)
			fmt.Fprintf(writer, "Exec:\t %v\n", action.Path+" "+action.Args)
		}
	}
	fmt.Fprintf(writer, "Enabled:\t %v\n", registeredTask.Enabled)
	fmt.Fprintf(writer, "State:\t %s\n", registeredTask.State.String())
	fmt.Fprintf(writer, "Missed runs:\t %d\n", registeredTask.MissedRuns)
	fmt.Fprintf(writer, "Last Run Time:\t %v\n", registeredTask.LastRunTime)
	fmt.Fprintf(writer, "Last Result:\t %d\n", registeredTask.LastTaskResult)
	fmt.Fprintf(writer, "Next Run Time:\t %v\n", registeredTask.NextRunTime)
	writer.Flush()
	return nil
}

func getTaskPath(profileName, commandName string) string {
	return fmt.Sprintf("%s%s %s", tasksPath, profileName, commandName)
}

func convertWeekdaysToBitmap(weekdays []int) int {
	if weekdays == nil || len(weekdays) == 0 {
		return 0
	}
	bitmap := 0
	for _, weekday := range weekdays {
		bitmap |= getWeekdayBit(weekday)
	}
	return bitmap
}

func getWeekdayBit(weekday int) int {
	switch weekday {
	case 0:
		return 1
	case 1:
		return 2
	case 2:
		return 4
	case 3:
		return 8
	case 4:
		return 16
	case 5:
		return 32
	case 6:
		return 64
	case 7:
		// Sunday is the first day of the week
		return 1
	}
	return 0
}

func convertMonthsToBitmap(months []int) int {
	if months == nil {
		return 0
	}
	if len(months) == 0 {
		// all values
		return int(math.Pow(2, 12)) - 1
	}
	bitmap := 0
	for _, month := range months {
		bitmap |= int(math.Pow(2, float64(month-1)))
	}
	return bitmap
}

func convertDaysToBitmap(days []int) int {
	if days == nil {
		return 0
	}
	if len(days) == 0 {
		// every day
		return int(math.Pow(2, 31)) - 1
	}
	bitmap := 0
	for _, day := range days {
		bitmap |= int(math.Pow(2, float64(day-1)))
	}
	return bitmap
}
