package profile

type Profile struct {
	Quiet        bool                   `mapstructure:"quiet"`
	Verbose      bool                   `mapstructure:"verbose"`
	Repository   string                 `mapstructure:"repository"`
	Initialize   bool                   `mapstructure:"initialize"`
	ForgetBefore bool                   `mapstructure:"forget-before"`
	ForgetAfter  bool                   `mapstructure:"forget-after"`
	CheckBefore  bool                   `mapstructure:"check-before"`
	CheckAfter   bool                   `mapstructure:"check-after"`
	RunBefore    []string               `mapstructure:"run-before"`
	RunAfter     []string               `mapstructure:"run-after"`
	UseStdin     bool                   `mapstructure:"stdin"`
	Inherit      string                 `mapstructure:"inherit"`
	Source       []string               `mapstructure:"source"`
	OtherFlags   map[string]interface{} `mapstructure:",remain"`

	Name        string
	Environment map[string]string
}

func NewProfile(name string) *Profile {
	return &Profile{
		Name: name,
	}
}
