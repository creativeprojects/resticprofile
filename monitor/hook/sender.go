package hook

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
)

type Sender struct {
	client    *http.Client
	userAgent string
}

func NewSender(userAgent string, timeout time.Duration) *Sender {
	if userAgent == "" {
		userAgent = "resticprofile/1.0"
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
	}
	return &Sender{
		client:    client,
		userAgent: userAgent,
	}
}

func (s *Sender) Send(cfg config.SendMonitorSection, ctx Context) error {
	if cfg.URL == "" {
		return errors.New("URL field is empty")
	}
	method := cfg.Method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader = http.NoBody
	if cfg.Body != "" {
		body = bytes.NewBufferString(resolveBody(cfg.Body, ctx))
	}
	req, err := http.NewRequest(method, cfg.URL, body)
	if err != nil {
		return err
	}
	for _, header := range cfg.Headers {
		if header.Name == "" {
			continue
		}
		req.Header.Add(header.Name, header.Value)
	}
	s.setUserAgent(req)

	clog.Debugf("calling %q", req.URL.String())
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (s *Sender) setUserAgent(req *http.Request) {
	userAgentKey := "User-Agent"
	if req.Header.Get(userAgentKey) == "" {
		req.Header.Add(userAgentKey, s.userAgent)
	}
}

func resolveBody(body string, ctx Context) string {
	body = os.Expand(body, func(s string) string {
		switch s {
		case constants.EnvProfileName:
			return ctx.ProfileName

		case constants.EnvProfileCommand:
			return ctx.ProfileCommand

		case constants.EnvError:
			return ctx.Error.Message

		case constants.EnvErrorCommandLine:
			return ctx.Error.CommandLine

		case constants.EnvErrorExitCode:
			return ctx.Error.ExitCode

		case constants.EnvErrorStderr:
			return ctx.Error.Stderr

		default:
			return os.Getenv(s)
		}
	})
	return body
}
