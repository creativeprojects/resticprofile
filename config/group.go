package config

// Group of profiles
type Group struct {
	Description     string   `mapstructure:"description" description:"Describe the group"`
	Profiles        []string `mapstructure:"profiles" description:"Names of the profiles belonging to this group"`
	ContinueOnError *bool    `mapstructure:"continue-on-error" default:"auto" description:"Continue with the next profile on a failure, overrides \"global.group-continue-on-error\""`
}
