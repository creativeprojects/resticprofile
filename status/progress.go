package status

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/progress"
)

type Progress struct {
	profile *config.Profile
}

func NewProgress(profile *config.Profile) *Progress {
	return &Progress{
		profile: profile,
	}
}

func (p *Progress) Status(status progress.Status) {
	// we don't report any progress here
}

func (p *Progress) Summary(command string, summary progress.Summary, stderr string, result error) {
	if p.profile.StatusFile == "" {
		return
	}
	switch {
	case progress.IsSuccess(result):
		p.success(command, summary, stderr)

	case progress.IsWarning(result):
		p.success(command, summary, stderr)

	case progress.IsError(result):
		p.error(command, summary, stderr, result)
	}
}

func (p *Progress) success(command string, summary progress.Summary, stderr string) {
	var err error
	switch command {
	case constants.CommandBackup:
		status := NewStatus(p.profile.StatusFile).Load()
		status.Profile(p.profile.Name).BackupSuccess(summary, stderr)
		err = status.Save()
	case constants.CommandCheck:
		status := NewStatus(p.profile.StatusFile).Load()
		status.Profile(p.profile.Name).CheckSuccess(summary, stderr)
		err = status.Save()
	case constants.SectionConfigurationRetention, constants.CommandForget:
		status := NewStatus(p.profile.StatusFile).Load()
		status.Profile(p.profile.Name).RetentionSuccess(summary, stderr)
		err = status.Save()
	}
	if err != nil {
		// not important enough to throw an error here
		clog.Warningf("saving status file '%s': %v", p.profile.StatusFile, err)
	}
}

func (p *Progress) error(command string, summary progress.Summary, stderr string, fail error) {
	var err error
	switch command {
	case constants.CommandBackup:
		status := NewStatus(p.profile.StatusFile).Load()
		status.Profile(p.profile.Name).BackupError(fail, summary, stderr)
		err = status.Save()
	case constants.CommandCheck:
		status := NewStatus(p.profile.StatusFile).Load()
		status.Profile(p.profile.Name).CheckError(fail, summary, stderr)
		err = status.Save()
	case constants.SectionConfigurationRetention, constants.CommandForget:
		status := NewStatus(p.profile.StatusFile).Load()
		status.Profile(p.profile.Name).RetentionError(fail, summary, stderr)
		err = status.Save()
	}
	if err != nil {
		// not important enough to throw an error here
		clog.Warningf("saving status file '%s': %v", p.profile.StatusFile, err)
	}
}

// Verify interface
var _ progress.Receiver = &Progress{}
