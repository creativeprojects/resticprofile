package shell

import "time"

// Summary of the profile run
type Summary struct {
	Duration        time.Duration
	FilesNew        int
	FilesChanged    int
	FilesUnmodified int
	DirsNew         int
	DirsChanged     int
	DirsUnmodified  int
	FilesTotal      int
	BytesAdded      uint64
	BytesTotal      uint64
}
