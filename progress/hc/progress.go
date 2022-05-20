package hc

import (
	"net/http"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/progress"
)

type Progress struct {
	profile *config.Profile
	client  *http.Client
}

func NewProgress(profile *config.Profile) *Progress {
	return &Progress{
		profile: profile,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *Progress) Start(command string) {
	if p.profile.HealthChecksURL == "" {
		return
	}
	uuid := p.getUUID(command)
	if uuid == "" {
		return
	}
	url := join(p.profile.HealthChecksURL, uuid, "start")
	clog.Debugf("healthchecks.io: send signal to %s", url)
	_, err := p.client.Head(url)
	if err != nil {
		clog.Errorf("error contacting healthchecks.io service: %s", err)
	}
}

func (p *Progress) Status(status progress.Status) {
	// we don't report any progress
}

func (p *Progress) Summary(command string, summary progress.Summary, stderr string, result error) {
	//
}

func (p *Progress) getUUID(command string) string {
	switch command {
	case constants.CommandBackup:
		if p.profile.Backup != nil {
			return p.profile.Backup.HealthChecksUUID
		}
		return ""
	}
	return ""
}

func join(host, uuid, cmd string) string {
	host = strings.TrimSuffix(host, "/")
	uuid = strings.TrimPrefix(uuid, "/")
	cmd = strings.TrimPrefix(cmd, "/")
	output := host + "/" + uuid
	if cmd != "" {
		output += "/" + cmd
	}
	return output
}

// Verify interface
var _ progress.Receiver = &Progress{}
