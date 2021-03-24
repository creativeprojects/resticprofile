package config

import (
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
)

// ScheduleConfig contains all information to schedule a profile command
type ScheduleConfig struct {
	profileName      string
	commandName      string
	schedules        []string
	permission       string
	wd               string
	command          string
	arguments        []string
	environment      map[string]string
	jobDescription   string
	timerDescription string
	priority         string
	logfile          string
	configfile       string
	flags            map[string]string
}

func (s *ScheduleConfig) SetCommand(wd, command string, args []string) {
	s.wd = wd
	s.command = command
	s.arguments = args
}

func (s *ScheduleConfig) SetJobDescription(description string) {
	s.jobDescription = description
}

func (s *ScheduleConfig) SetTimerDescription(description string) {
	s.timerDescription = description
}

func (s *ScheduleConfig) Title() string {
	return s.profileName
}

func (s *ScheduleConfig) SubTitle() string {
	return s.commandName
}

func (s *ScheduleConfig) JobDescription() string {
	return s.jobDescription
}

func (s *ScheduleConfig) TimerDescription() string {
	return s.timerDescription
}

func (s *ScheduleConfig) Schedules() []string {
	return s.schedules
}

func (s *ScheduleConfig) Permission() string {
	return s.permission
}

func (s *ScheduleConfig) Command() string {
	return s.command
}

func (s *ScheduleConfig) WorkingDirectory() string {
	return s.wd
}

func (s *ScheduleConfig) Arguments() []string {
	return s.arguments
}

func (s *ScheduleConfig) Environment() map[string]string {
	return s.environment
}

// Priority is either "background" or "standard"
func (s *ScheduleConfig) Priority() string {
	s.priority = strings.ToLower(s.priority)
	// default value for priority is "background"
	if s.priority != constants.SchedulePriorityBackground && s.priority != constants.SchedulePriorityStandard {
		s.priority = constants.SchedulePriorityBackground
	}
	return s.priority
}

func (s *ScheduleConfig) Logfile() string {
	return s.logfile
}

func (s *ScheduleConfig) Configfile() string {
	return s.configfile
}

func (s *ScheduleConfig) GetFlag(name string) (string, bool) {
	if len(s.flags) == 0 {
		return "", false
	}
	// we can't do a direct return, technically the map returns only one value
	value, found := s.flags[name]
	return value, found
}

func (s *ScheduleConfig) SetFlag(name, value string) {
	if s.flags == nil {
		s.flags = make(map[string]string)
	}
	s.flags[name] = value
}
