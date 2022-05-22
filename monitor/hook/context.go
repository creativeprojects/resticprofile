package hook

type Context struct {
	ProfileName    string
	ProfileCommand string
	Error          ErrorContext
}

type ErrorContext struct {
	Message     string
	CommandLine string
	ExitCode    int
	Stderr      string
}
