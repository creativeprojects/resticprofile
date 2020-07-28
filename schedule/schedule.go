package schedule

import (
	"os"
	"runtime"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

var (
	// ScheduledSections are the command that can be scheduled (backup, retention, check)
	ScheduledSections = []string{
		constants.CommandBackup,
		constants.SectionConfigurationRetention,
		constants.CommandCheck,
	}
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
	Environment() map[string]string
	Nice() int
	Logfile() string
}

// Job scheduler
type Job struct {
	config Config
}

// NewJob instantiates a Job object to schedule jobs
func NewJob(config Config) *Job {
	return &Job{
		config: config,
	}
}

// Create a new job
func (j *Job) Create() error {
	schedules, err := loadSchedules(j.config.SubTitle(), j.config.Schedules())
	if err != nil {
		return err
	}

	err = j.createJob(schedules)
	if err != nil {
		return err
	}

	return nil
}

// Update an existing job
func (j *Job) Update() error {
	return nil
}

// Remove a job
func (j *Job) Remove() error {
	err := j.removeJob()
	if err != nil {
		return err
	}
	return nil
}

// Status of a job
func (j *Job) Status() error {
	_, err := loadSchedules(j.config.SubTitle(), j.config.Schedules())
	if err != nil {
		return err
	}

	err = j.displayStatus(j.config.SubTitle())
	if err != nil {
		return err
	}
	return nil
}

// getSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
//
// This method is for Unixes only
func (j *Job) getSchedulePermission() string {
	const message = "you have not specified the permission for your schedule (system or user): assuming "
	if j.config.Permission() == constants.SchedulePermissionSystem ||
		j.config.Permission() == constants.SchedulePermissionUser {
		// well defined
		return j.config.Permission()
	}
	// best guess is depending on the user being root or not:
	if os.Geteuid() == 0 {
		if runtime.GOOS != "darwin" {
			// darwin can backup protected files without the need of a system task; no need to bother the user then
			clog.Warning(message, "system")
		}
		return constants.SchedulePermissionSystem
	}
	if runtime.GOOS != "darwin" {
		// darwin can backup protected files without the need of a system task; no need to bother the user then
		clog.Warning(message, "user")
	}
	return constants.SchedulePermissionUser
}

// checkPermission returns true if the user is allowed.
//
// This method is for Unixes only
func (j *Job) checkPermission(permission string) bool {
	if permission == constants.SchedulePermissionUser {
		// user mode is always available
		return true
	}
	if os.Geteuid() == 0 {
		// user has sudoed
		return true
	}
	// last case is system (or undefined) + no sudo
	return false
}
