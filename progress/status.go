package progress

type Status struct {
	PercentDone  float64
	TotalFiles   int
	FilesDone    int
	TotalBytes   int64
	BytesDone    int64
	ErrorCount   int
	CurrentFiles []string
}
