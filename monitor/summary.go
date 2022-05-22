package monitor

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
	OutputAnalysis  OutputAnalysis
}

// OutputAnalysis of the profile run
type OutputAnalysis interface {
	// ContainsRemoteLockFailure returns true if the output indicates that remote locking failed.
	ContainsRemoteLockFailure() bool

	// GetRemoteLockedSince returns the time duration since the remote lock was created.
	// If no remote lock is held or the time cannot be determined, the second parameter is false.
	GetRemoteLockedSince() (time.Duration, bool)

	// GetRemoteLockedBy returns who locked the remote lock, if available.
	GetRemoteLockedBy() (string, bool)
}
