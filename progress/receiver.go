package progress

type Receiver interface {
	Status(status Status)
	Summary(command string, summary Summary, stderr string, result error)
}
