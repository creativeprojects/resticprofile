package remote

const (
	ManifestFilename = ".manifest.json"
)

type Manifest struct {
	Version              string
	ConfigurationFile    string
	ProfileName          string
	CommandLineArguments string
}
