package config

import "github.com/creativeprojects/resticprofile/util/maybe"

// Group of profiles
type Group struct {
	config           *Config
	Name             string                    `show:"noshow"`
	Description      string                    `mapstructure:"description" description:"Describe the group"`
	Profiles         []string                  `mapstructure:"profiles" description:"Names of the profiles belonging to this group"`
	ContinueOnError  maybe.Bool                `mapstructure:"continue-on-error" default:"auto" description:"Continue with the next profile on a failure, overrides \"global.group-continue-on-error\""`
	CommandSchedules map[string]ScheduleConfig `mapstructure:"schedules" description:"Allows to run the group on schedule for the specified command name."`
}

func NewGroup(c *Config, name string) (g *Group) {
	g = &Group{
		Name:   name,
		config: c,
	}
	return
}

func (g *Group) ResolveConfiguration() {
	global := g.config.mustGetGlobalSection()
	for command, cfg := range g.CommandSchedules {
		cfg.init(global.ScheduleDefaults)
		cfg.origin = ScheduleOrigin(g.Name, command, ScheduleOriginGroup)
	}
}

func (g *Group) Schedules() map[string]*Schedule {
	schedules := make(map[string]*Schedule)
	for command, cfg := range g.CommandSchedules {
		schedules[command] = NewSchedule(g.config, &cfg)
	}
	return schedules
}

// Implements Schedulable
var _ Schedulable = new(Group)
