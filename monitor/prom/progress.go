package prom

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	if command == constants.SectionConfigurationRetention {
		command = constants.CommandForget
	}
	if command != constants.CommandBackup && !p.metrics.HasCommand(command) {
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
	if command == constants.CommandBackup {
		p.metrics.BackupResults(status, summary)
	} else {
		p.metrics.CommandResults(command, status, exitCode(result), summary)
	}

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
		err := p.metrics.Push(p.profile.PrometheusPush, p.profile.PrometheusPushFormat, jobName, command)
		if err != nil {
			// not important enough to throw an error here
			clog.Warningf("pushing prometheus metrics to %q: %v", p.profile.PrometheusPush, err)
		}
	}
}

// exitCode returns the exit code of restic from the result of the command
func exitCode(result error) int {
	if result == nil {
		return 0
	}
	exitErr := &exec.ExitError{}
	if errors.As(result, &exitErr) {
		return exitErr.ExitCode()
	}
	warn := &monitor.InternalWarningError{}
	if errors.As(result, &warn) {
		return constants.ResticExitCodeWarning
	}
	return -1
}

// Verify interface
var _ monitor.Receiver = &Progress{}
