package config

import (
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/maybe"
)

// Group of profiles
type Group struct {
	config           *Config
	Name             string                     `show:"noshow"`
	Description      string                     `mapstructure:"description" description:"Describe the group"`
	Profiles         []string                   `mapstructure:"profiles" description:"Names of the profiles belonging to this group"`
	ContinueOnError  maybe.Bool                 `mapstructure:"continue-on-error" default:"auto" description:"Continue with the next profile on a failure, overrides \"global.group-continue-on-error\""`
	CommandSchedules map[string]*ScheduleConfig `mapstructure:"schedules" show:"noshow" description:"Allows to run the group on schedule for the specified command name (backup, copy, check, forget, prune)."`
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
		if cfg.HasSchedules() {
			cfg.init(global.ScheduleDefaults)
			cfg.origin = ScheduleOrigin(g.Name, command, ScheduleOriginGroup)
		} else {
			delete(g.CommandSchedules, command)
		}
	}
}

func (g *Group) Schedules() map[string]*Schedule {
	schedules := make(map[string]*Schedule)
	for command, cfg := range g.CommandSchedules {
		if cfg.HasSchedules() {
			schedules[command] = NewSchedule(g.config, cfg)
		}
	}
	return schedules
}

// SchedulableCommands returns the list of commands that can be scheduled (whether they have schedules or not)
func (g *Group) SchedulableCommands() []string {
	// once the deprecated retention schedule is removed, we can use the list from profiles
	// return NewProfile(g.config, "").SchedulableCommands()
	return []string{
		constants.CommandBackup,
		constants.CommandCheck,
		constants.CommandForget,
		constants.CommandPrune,
		constants.CommandCopy,
	}
}

func (g *Group) Kind() string {
	return constants.SchedulableKindGroup
}

// Implements Schedulable
var _ Schedulable = new(Group)
