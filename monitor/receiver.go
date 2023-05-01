package monitor

// Receiver is an interface for implementations interested in the restic command status and summary
//
//go:generate mockery --name=Receiver
type Receiver interface {
	// Start of a command
	Start(command string)
	// Status during the execution
	Status(status Status)
	// Summary at the end of a command
	Summary(command string, summary Summary, stderr string, result error)
}
