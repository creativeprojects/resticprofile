package monitor

type Receiver interface {
	// Start of a command
	Start(command string)
	// Status during the execution
	Status(status Status)
	// Summary at the end of a command
	Summary(command string, summary Summary, stderr string, result error)
}
