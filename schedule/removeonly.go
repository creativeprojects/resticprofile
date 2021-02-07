package schedule

// RemoveOnlyConfig implements Config for jobs that are Job.RemoveOnly()
type RemoveOnlyConfig struct {
	title, subTitle string
}

func (r *RemoveOnlyConfig) Title() string                  { return r.title }
func (r *RemoveOnlyConfig) SubTitle() string               { return r.subTitle }
func (r *RemoveOnlyConfig) JobDescription() string         { return "" }
func (r *RemoveOnlyConfig) TimerDescription() string       { return "" }
func (r *RemoveOnlyConfig) Schedules() []string            { return []string{} }
func (r *RemoveOnlyConfig) Permission() string             { return "" }
func (r *RemoveOnlyConfig) WorkingDirectory() string       { return "" }
func (r *RemoveOnlyConfig) Command() string                { return "" }
func (r *RemoveOnlyConfig) Arguments() []string            { return []string{} }
func (r *RemoveOnlyConfig) Environment() map[string]string { return map[string]string{} }
func (r *RemoveOnlyConfig) Priority() string               { return "" }
func (r *RemoveOnlyConfig) Logfile() string                { return "" }
func (r *RemoveOnlyConfig) Configfile() string             { return "" }
func (r *RemoveOnlyConfig) GetFlag(string) (string, bool)  { return "", false }

func isRemoveOnlyConfig(config Config) bool {
	_, ok := config.(*RemoveOnlyConfig)
	return ok
}

// NewRemoveOnlyConfig creates a job config that may be used to call Job.Remove() on a scheduled job
func NewRemoveOnlyConfig(profileName, commandName string) Config {
	return &RemoveOnlyConfig{
		title:    profileName,
		subTitle: commandName,
	}
}
