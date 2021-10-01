package schedule

import (
	"errors"
	"io"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schtasks"
)

// HandlerWindows is using windows task manager
type HandlerWindows struct {
	config SchedulerConfig
}

func NewHandler(config SchedulerConfig) *HandlerWindows {
	return &HandlerWindows{
		config: config,
	}
}

// Init a connection to the task scheduler
func (h *HandlerWindows) Init() error {
	return schtasks.Connect()
}

// Close the connection to the task scheduler
func (h *HandlerWindows) Close() {
	schtasks.Close()
}

func (h *HandlerWindows) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	return parseSchedules(schedules)
}

func (h *HandlerWindows) DisplayParsedSchedules(command string, events []*calendar.Event) {
	displayParsedSchedules(command, events)
}

// DisplaySchedules does nothing on windows
func (h *HandlerWindows) DisplaySchedules(command string, schedules []string) error {
	return nil
}

// DisplayStatus does nothing on windows task manager
func (h *HandlerWindows) DisplayStatus(profileName string, w io.Writer) error {
	return nil
}

// CreateJob is creating the task scheduler job.
func (h *HandlerWindows) CreateJob(job JobConfig, schedules []*calendar.Event) error {
	// default permission will be system
	permission := schtasks.SystemAccount
	if p, _ := detectSchedulePermission(job.Permission()); p == constants.SchedulePermissionUser {
		permission = schtasks.UserAccount
	}
	err := schtasks.Create(job, schedules, permission)
	if err != nil {
		return err
	}
	return nil
}

// RemoveJob is deleting the task scheduler job
func (h *HandlerWindows) RemoveJob(job JobConfig) error {
	err := schtasks.Delete(job.Title(), job.SubTitle())
	if err != nil {
		if errors.Is(err, schtasks.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}

// DisplayStatus display some information about the task scheduler job
func (h *HandlerWindows) DisplayJobStatus(job JobConfig, w io.Writer) error {
	err := schtasks.Status(job.Title(), job.SubTitle())
	if err != nil {
		if errors.Is(err, schtasks.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}

// detectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// unsafe specifies whether a guess may lead to a too broad or too narrow file access permission.
func detectSchedulePermission(permission string) (detected string, unsafe bool) {
	if permission == constants.SchedulePermissionUser {
		return constants.SchedulePermissionUser, false
	}
	return constants.SchedulePermissionSystem, false
}

// checkPermission returns true if the user is allowed to access the job.
// This is always true on Windows
func checkPermission(permission string) bool {
	return true
}
