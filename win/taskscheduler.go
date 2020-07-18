//+build windows

package win

import (
	"fmt"
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

// TaskScheduler wraps up a task scheduler service
type TaskScheduler struct {
	config Config
}

// NewTaskScheduler creates a new service to talk to windows task scheduler
func NewTaskScheduler(config Config) *TaskScheduler {
	return &TaskScheduler{
		config: config,
	}
}

// Create a task
func (s *TaskScheduler) Create(schedules []*calendar.Event, permission Permission) error {
	if permission == SystemAccount {
		return s.createSystemTask(schedules)
	}
	return s.createUserTask(schedules)
}

func (s *TaskScheduler) createUserTask(schedules []*calendar.Event) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	fmt.Printf("\nCreating task for user %s\n", currentUser.Username)
	fmt.Printf("Windows requires your password to validate the task: ")
	password, err := term.ReadPassword()
	if err != nil {
		return err
	}

	taskService, err := s.connect()
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	task := taskService.NewTaskDefinition()
	task.AddExecAction(
		s.config.Command(),
		strings.Join(s.config.Arguments(), " "),
		s.config.WorkingDirectory(),
		"")
	task.Principal.LogonType = taskmaster.TASK_LOGON_PASSWORD
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_LUA
	task.Principal.UserID = currentUser.Username
	task.RegistrationInfo.Author = "resticprofile"
	task.RegistrationInfo.Description = s.config.JobDescription()

	s.createSchedules(&task, schedules)

	_, _, err = taskService.CreateTaskEx(
		getTaskPath(s.config.Title(), s.config.SubTitle()),
		task,
		currentUser.Username,
		password,
		taskmaster.TASK_LOGON_PASSWORD,
		true)
	if err != nil {
		return err
	}
	return nil
}

func (s *TaskScheduler) createSystemTask(schedules []*calendar.Event) error {
	taskService, err := s.connect()
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	task := taskService.NewTaskDefinition()
	task.AddExecAction(
		s.config.Command(),
		strings.Join(s.config.Arguments(), " "),
		s.config.WorkingDirectory(),
		"")
	task.Principal.LogonType = taskmaster.TASK_LOGON_SERVICE_ACCOUNT
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_HIGHEST
	task.Principal.UserID = "SYSTEM"
	task.RegistrationInfo.Author = "resticprofile"
	task.RegistrationInfo.Description = s.config.JobDescription()

	s.createSchedules(&task, schedules)

	_, _, err = taskService.CreateTask(getTaskPath(s.config.Title(), s.config.SubTitle()), task, true)
	if err != nil {
		return err
	}
	return nil
}

func (s *TaskScheduler) createSchedules(task *taskmaster.Definition, schedules []*calendar.Event) {
	for _, schedule := range schedules {
		if once, ok := schedule.AsTime(); ok {
			// one time only
			task.AddTimeTrigger(period.Period{}, once)
			continue
		}
		if !schedule.WeekDay.HasValue() && !schedule.Month.HasValue() && !schedule.Day.HasValue() {
			// recurring daily
			s.createDailyTrigger(task, schedule)
			continue
		}
	}
}

func (s *TaskScheduler) createDailyTrigger(task *taskmaster.Definition, schedule *calendar.Event) {
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
}

// Update a task
func (s *TaskScheduler) Update() error {
	return nil
}

// Delete a task
func (s *TaskScheduler) Delete() error {
	taskService, err := s.connect()
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	err = taskService.DeleteTask(getTaskPath(s.config.Title(), s.config.SubTitle()))
	if err != nil {
		return err
	}
	return nil
}

// Status returns the status of a task
func (s *TaskScheduler) Status() error {
	taskService, err := s.connect()
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	taskName := getTaskPath(s.config.Title(), s.config.SubTitle())
	registeredTask, err := taskService.GetRegisteredTask(taskName)
	if err != nil {
		return err
	}
	if registeredTask == nil {
		return fmt.Errorf("%w: %s", ErrorNotRegistered, taskName)
	}
	writer := tabwriter.NewWriter(term.GetOutput(), 2, 2, 2, ' ', 0)
	fmt.Fprintf(writer, "Task\t%s\n", registeredTask.Path)
	fmt.Fprintf(writer, "User\t%s\n", registeredTask.Definition.Principal.UserID)
	if registeredTask.Definition.Actions != nil && len(registeredTask.Definition.Actions) > 0 {
		if action, ok := registeredTask.Definition.Actions[0].(taskmaster.ExecAction); ok {
			fmt.Fprintf(writer, "Working Dir\t%v\n", action.WorkingDir)
			fmt.Fprintf(writer, "Exec\t%v\n", action.Path+" "+action.Args)
		}
	}
	fmt.Fprintf(writer, "Enabled\t%v\n", registeredTask.Enabled)
	fmt.Fprintf(writer, "State\t%s\n", registeredTask.State.String())
	fmt.Fprintf(writer, "Missed runs\t%d\n", registeredTask.MissedRuns)
	fmt.Fprintf(writer, "Next Run Time\t%v\n", registeredTask.NextRunTime)
	fmt.Fprintf(writer, "Last Run Time\t%v\n", registeredTask.LastRunTime)
	fmt.Fprintf(writer, "Last Task Result\t%d\n", registeredTask.LastTaskResult)
	writer.Flush()
	return nil
}

func (s *TaskScheduler) connect() (*taskmaster.TaskService, error) {
	return taskmaster.Connect("", "", "", "")
}

func getTaskPath(profileName, commandName string) string {
	return fmt.Sprintf("%s%s %s", tasksPath, profileName, commandName)
}
