package hook

type Context struct {
	ProfileName    string
	ProfileCommand string
	Error          ErrorContext
	Stdout         string
}

type ErrorContext struct {
	Message     string
	CommandLine string
	ExitCode    string
	Stderr      string
}
