package hc

import (
	"bytes"
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
	timeout := 10 * time.Second
	if profile.HealthChecksTimeout > 0 {
		timeout = profile.HealthChecksTimeout
	}
	return &Progress{
		profile: profile,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *Progress) Start(command string) {
	url := p.getURL(command, "start")
	if url == "" {
		return
	}
	req, err := p.newSimpleRequest(url)
	if err != nil {
		clog.Errorf("cannot send healthchecks.io request: %s", err)
		return
	}
	p.sendRequest(req)
}

func (p *Progress) Status(status progress.Status) {
	// we don't report any progress
}

func (p *Progress) Summary(command string, summary progress.Summary, stderr string, result error) {
	path := ""
	if result != nil {
		path = "fail"
	}
	url := p.getURL(command, path)
	if url == "" {
		return
	}
	if stderr == "" && result == nil {
		// no body to send
		req, err := p.newSimpleRequest(url)
		if err != nil {
			clog.Errorf("cannot send healthchecks.io request: %s", err)
			return
		}
		p.sendRequest(req)
		return
	}
	body := ""
	if result != nil {
		body = result.Error() + "\n\n"
	}
	body += stderr
	req, err := p.newRequestWithBody(url, body)
	if err != nil {
		clog.Errorf("cannot send healthchecks.io request: %s", err)
		return
	}
	p.sendRequest(req)
}

func (p *Progress) newSimpleRequest(url string) (*http.Request, error) {
	return http.NewRequest(http.MethodHead, url, http.NoBody)
}

func (p *Progress) newRequestWithBody(url, body string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/html; charset=UTF-8")
	return req, nil
}

func (p *Progress) sendRequest(req *http.Request) {
	clog.Debugf("healthchecks.io: send signal to %s", req.URL.String())
	resp, err := p.client.Do(req)
	if err != nil {
		clog.Errorf("error contacting healthchecks.io service: %s", err)
		return
	}
	resp.Body.Close()
}

func (p *Progress) getURL(command, path string) string {
	if p.profile.HealthChecksURL == "" {
		return ""
	}
	uuid := p.getUUID(command)
	if uuid == "" {
		return ""
	}
	url := join(p.profile.HealthChecksURL, uuid, path)
	return url
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
