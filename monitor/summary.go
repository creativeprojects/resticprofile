package monitor

import "time"

// Summary of the profile run
type Summary struct {
	Duration        time.Duration
	FilesNew        int64
	FilesChanged    int64
	FilesUnmodified int64
	DirsNew         int64
	DirsChanged     int64
	DirsUnmodified  int64
	FilesTotal      int64
	BytesAdded      uint64
	BytesStored     uint64
	BytesTotal      uint64
	Compression     float64
	SnapshotID      string
	Extended        bool
	DryRun          bool
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

	// GetFailedFiles returns a list of files that failed to get processed.
	GetFailedFiles() []string
}
