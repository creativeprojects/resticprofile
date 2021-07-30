package prom

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/progress"
)

type Progress struct {
	profile *config.Profile
	metrics *Metrics
}

func NewProgress(profile *config.Profile, metrics *Metrics) *Progress {
	return &Progress{
		profile: profile,
		metrics: metrics,
	}
}

func (p *Progress) Status(status progress.Status) {
	// we don't report any progress here yet
}

func (p *Progress) Summary(command string, summary progress.Summary, stderr string, result error) {
	if p.profile.PrometheusPush == "" && p.profile.PrometheusSaveToFile == "" {
		return
	}
	var status Status
	switch {
	case progress.IsSuccess(result):
		status = StatusSuccess

	case progress.IsWarning(result):
		status = StatusWarning

	case progress.IsError(result):
		status = StatusFailed
	}
	if command != constants.CommandBackup {
		return
	}
	p.metrics.BackupResults(p.profile.Name, status, summary)

	if p.profile.PrometheusSaveToFile != "" {
		err := p.metrics.SaveTo(p.profile.PrometheusSaveToFile)
		if err != nil {
			// not important enough to throw an error here
			clog.Warningf("saving prometheus file %q: %v", p.profile.PrometheusSaveToFile, err)
		}
	}
	if p.profile.PrometheusPush != "" {
		err := p.metrics.Push(p.profile.PrometheusPush, command)
		if err != nil {
			// not important enough to throw an error here
			clog.Warningf("pushing prometheus metrics to %q: %v", p.profile.PrometheusPush, err)
		}
	}
}

// Verify interface
var _ progress.Receiver = &Progress{}
