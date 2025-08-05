package remote

type Manifest struct {
	Version              string // resticprofile version
	ConfigurationFile    string
	ProfileName          string
	Mountpoint           string // Mountpoint of the virtual FS if configured
	CommandLineArguments []string
}
