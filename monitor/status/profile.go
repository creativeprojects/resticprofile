package status

import (
	"math"
	"time"

	"github.com/creativeprojects/resticprofile/monitor"
)

// Profile status
type Profile struct {
	Backup    *BackupStatus  `json:"backup,omitempty"`
	Retention *CommandStatus `json:"retention,omitempty"`
	Check     *CommandStatus `json:"check,omitempty"`
}

func newProfile() *Profile {
	return &Profile{}
}

// CommandStatus is the last command status
type CommandStatus struct {
	Success  bool      `json:"success"`
	Time     time.Time `json:"time"`
	Error    string    `json:"error"`
	Stderr   string    `json:"stderr"`
	Duration int64     `json:"duration"`
}

// BackupStatus contains the last backup status
type BackupStatus struct {
	CommandStatus
	FilesNew        int    `json:"files_new"`
	FilesChanged    int    `json:"files_changed"`
	FilesUnmodified int    `json:"files_unmodified"`
	DirsNew         int    `json:"dirs_new"`
	DirsChanged     int    `json:"dirs_changed"`
	DirsUnmodified  int    `json:"dirs_unmodified"`
	FilesTotal      int    `json:"files_total"`
	BytesAdded      uint64 `json:"bytes_added"`
	BytesTotal      uint64 `json:"bytes_total"`
}

// BackupSuccess indicates the last backup was successful
func (p *Profile) BackupSuccess(summary monitor.Summary, stderr string) *Profile {
	p.Backup = &BackupStatus{
		CommandStatus: CommandStatus{
			Success:  true,
			Time:     time.Now(),
			Duration: int64(math.Ceil(summary.Duration.Seconds())),
			Stderr:   stderr,
		},
		FilesNew:        summary.FilesNew,
		FilesChanged:    summary.FilesChanged,
		FilesUnmodified: summary.FilesUnmodified,
		DirsNew:         summary.DirsNew,
		DirsChanged:     summary.DirsChanged,
		DirsUnmodified:  summary.DirsUnmodified,
		FilesTotal:      summary.FilesTotal,
		BytesAdded:      summary.BytesAdded,
		BytesTotal:      summary.BytesTotal,
	}
	return p
}

// BackupError sets the error of the last backup
func (p *Profile) BackupError(err error, summary monitor.Summary, stderr string) *Profile {
	p.Backup = &BackupStatus{
		CommandStatus: CommandStatus{
			Success:  false,
			Time:     time.Now(),
			Error:    err.Error(),
			Duration: int64(math.Ceil(summary.Duration.Seconds())),
			Stderr:   stderr,
		},
		FilesNew:        0,
		FilesChanged:    0,
		FilesUnmodified: 0,
		DirsNew:         0,
		DirsChanged:     0,
		DirsUnmodified:  0,
		FilesTotal:      0,
		BytesAdded:      0,
		BytesTotal:      0,
	}
	return p
}

// RetentionSuccess indicates the last retention was successful
func (p *Profile) RetentionSuccess(summary monitor.Summary, stderr string) *Profile {
	p.Retention = newSuccess(summary.Duration, stderr)
	return p
}

// RetentionError sets the error of the last retention
func (p *Profile) RetentionError(err error, summary monitor.Summary, stderr string) *Profile {
	p.Retention = newError(err, summary.Duration, stderr)
	return p
}

// CheckSuccess indicates the last check was successful
func (p *Profile) CheckSuccess(summary monitor.Summary, stderr string) *Profile {
	p.Check = newSuccess(summary.Duration, stderr)
	return p
}

// CheckError sets the error of the last check
func (p *Profile) CheckError(err error, summary monitor.Summary, stderr string) *Profile {
	p.Check = newError(err, summary.Duration, stderr)
	return p
}

func newSuccess(duration time.Duration, stderr string) *CommandStatus {
	return &CommandStatus{
		Success:  true,
		Time:     time.Now(),
		Duration: int64(math.Ceil(duration.Seconds())),
		Stderr:   stderr,
	}
}

func newError(err error, duration time.Duration, stderr string) *CommandStatus {
	return &CommandStatus{
		Success:  false,
		Time:     time.Now(),
		Error:    err.Error(),
		Duration: int64(math.Ceil(duration.Seconds())),
		Stderr:   stderr,
	}
}
