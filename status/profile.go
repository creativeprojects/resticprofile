package status

import "time"

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
func (p *Profile) BackupSuccess() *Profile {
	p.Backup = newSuccess()
	return p
}

// BackupError sets the error of the last backup
func (p *Profile) BackupError(err error) *Profile {
	p.Backup = newError(err)
	return p
}

// RetentionSuccess indicates the last retention was successful
func (p *Profile) RetentionSuccess() *Profile {
	p.Retention = newSuccess()
	return p
}

// RetentionError sets the error of the last retention
func (p *Profile) RetentionError(err error) *Profile {
	p.Retention = newError(err)
	return p
}

// CheckSuccess indicates the last check was successful
func (p *Profile) CheckSuccess() *Profile {
	p.Check = newSuccess()
	return p
}

// CheckError sets the error of the last check
func (p *Profile) CheckError(err error) *Profile {
	p.Check = newError(err)
	return p
}

func newSuccess() *CommandStatus {
	return &CommandStatus{
		Success: true,
		Time:    time.Now(),
	}
}

func newError(err error) *CommandStatus {
	return &CommandStatus{
		Success: false,
		Time:    time.Now(),
		Error:   err.Error(),
	}
}
