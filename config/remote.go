package config

type Remote struct {
	name              string
	config            *Config
	Connection        string   `mapstructure:"connection" default:"ssh" description:"Connection type to use to connect to the remote client"`
	Host              string   `mapstructure:"host" description:"Address of the remote client. Format: <host>:<port>"`
	Username          string   `mapstructure:"username" description:"User to connect to the remote client"`
	PrivateKeyPath    string   `mapstructure:"private-key" description:"Path to the private key to use for authentication"`
	KnownHostsPath    string   `mapstructure:"known-hosts" description:"Path to the known hosts file"`
	BinaryPath        string   `mapstructure:"binary-path" description:"Path to the resticprofile binary to use on the remote client"`
	ConfigurationFile string   `mapstructure:"configuration-file" description:"Path to the configuration file to transfer to the remote client"`
	ProfileName       string   `mapstructure:"profile-name" description:"Name of the profile to use on the remote client"`
	SendFiles         []string `mapstructure:"send-files" description:"Other configuration files to transfer to the remote client"`
}

func NewRemote(config *Config, name string) *Remote {
	remote := &Remote{
		name:   name,
		config: config,
	}
	return remote
}

// SetRootPath changes the path of all the relative paths and files in the configuration
func (r *Remote) SetRootPath(rootPath string) {
	r.PrivateKeyPath = fixPath(r.PrivateKeyPath, expandEnv, absolutePrefix(rootPath))
	r.KnownHostsPath = fixPath(r.KnownHostsPath, expandEnv, absolutePrefix(rootPath))
	r.ConfigurationFile = fixPath(r.ConfigurationFile, expandEnv, absolutePrefix(rootPath))

	for i := range r.SendFiles {
		r.SendFiles[i] = fixPath(r.SendFiles[i], expandEnv, absolutePrefix(rootPath))
	}
}
