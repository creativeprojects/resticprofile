package config

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
