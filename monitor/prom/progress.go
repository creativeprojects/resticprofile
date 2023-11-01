package prom

import (
	"fmt"
	"os"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
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

func (p *Progress) Start(command string) {
	// nothing to do here
}

func (p *Progress) Status(status monitor.Status) {
	// we don't report any progress here yet
}

func (p *Progress) Summary(command string, summary monitor.Summary, stderr string, result error) {
	if p.profile.PrometheusPush == "" && p.profile.PrometheusSaveToFile == "" {
		return
	}
	var status Status
	switch {
	case monitor.IsSuccess(result):
		status = StatusSuccess

	case monitor.IsWarning(result):
		status = StatusWarning

	case monitor.IsError(result):
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
		jobName := p.profile.PrometheusPushJob
		if jobName == "" {
			jobName = fmt.Sprintf("%s.%s", p.profile.Name, command)
		}
		jobName = os.Expand(jobName, func(name string) string {
			if strings.EqualFold(name, "command") {
				return command
			} else if name == "$" {
				return "$" // allow to escape "$" as "$$"
			}
			return ""
		})
		err := p.metrics.Push(p.profile.PrometheusPush, p.profile.PrometheusPushFormat, jobName)
		if err != nil {
			// not important enough to throw an error here
			clog.Warningf("pushing prometheus metrics to %q: %v", p.profile.PrometheusPush, err)
		}
	}
}

// Verify interface
var _ monitor.Receiver = &Progress{}
