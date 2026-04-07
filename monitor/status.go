package monitor

type Status struct {
	PercentDone      float64
	SecondsElapsed   int
	SecondsRemaining int
	TotalFiles       int64
	FilesDone        int64
	TotalBytes       uint64
	BytesDone        uint64
	ErrorCount       int64
	CurrentFiles     []string
}
