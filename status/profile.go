package status

import (
	"math"
	"time"

	"github.com/creativeprojects/resticprofile/shell"
)

// Profile status
type Profile struct {
	Backup    *CommandStatus `json:"backup,omitempty"`
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
	Duration int64     `json:"duration"`
}

// BackupSuccess indicates the last backup was successful
func (p *Profile) BackupSuccess(summary shell.Summary) *Profile {
	p.Backup = newSuccess(summary.Duration)
	return p
}

// BackupError sets the error of the last backup
func (p *Profile) BackupError(err error, summary shell.Summary) *Profile {
	p.Backup = newError(err, summary.Duration)
	return p
}

// RetentionSuccess indicates the last retention was successful
func (p *Profile) RetentionSuccess(summary shell.Summary) *Profile {
	p.Retention = newSuccess(summary.Duration)
	return p
}

// RetentionError sets the error of the last retention
func (p *Profile) RetentionError(err error, summary shell.Summary) *Profile {
	p.Retention = newError(err, summary.Duration)
	return p
}

// CheckSuccess indicates the last check was successful
func (p *Profile) CheckSuccess(summary shell.Summary) *Profile {
	p.Check = newSuccess(summary.Duration)
	return p
}

// CheckError sets the error of the last check
func (p *Profile) CheckError(err error, summary shell.Summary) *Profile {
	p.Check = newError(err, summary.Duration)
	return p
}

func newSuccess(duration time.Duration) *CommandStatus {
	return &CommandStatus{
		Success:  true,
		Time:     time.Now(),
		Duration: int64(math.Ceil(duration.Seconds())),
	}
}

func newError(err error, duration time.Duration) *CommandStatus {
	return &CommandStatus{
		Success:  false,
		Time:     time.Now(),
		Error:    err.Error(),
		Duration: int64(math.Ceil(duration.Seconds())),
	}
}
