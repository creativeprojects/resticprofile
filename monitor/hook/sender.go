package hook

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
)

type Sender struct {
	client         *http.Client
	insecureClient *http.Client
	userAgent      string
}

func NewSender(certificates []string, userAgent string, timeout time.Duration) *Sender {
	if userAgent == "" {
		userAgent = "resticprofile/1.0"
	}

	// normal client
	client := &http.Client{
		Timeout: timeout,
	}

	if len(certificates) > 0 {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{
			RootCAs: getRootCAs(certificates),
		}
		client.Transport = transport
	}

	// another client for insecure requests
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	insecureClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &Sender{
		client:         client,
		insecureClient: insecureClient,
		userAgent:      userAgent,
	}
}

func (s *Sender) Send(cfg config.SendMonitoringSection, ctx Context) error {
	if cfg.URL == "" {
		return errors.New("URL field is empty")
	}
	url := resolve(cfg.URL, ctx)
	method := cfg.Method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader = http.NoBody
	if cfg.BodyTemplate != "" {
		bodyTemplate, err := loadBodyTemplate(cfg.BodyTemplate, ctx)
		if err != nil {
			return err
		}
		body = bytes.NewBufferString(bodyTemplate)
	}
	if cfg.Body != "" {
		body = bytes.NewBufferString(resolve(cfg.Body, ctx))
	}
	req, err := http.NewRequest(method, url, body)
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

	client := s.client
	if cfg.SkipTLS {
		client = s.insecureClient
	}

	clog.Debugf("calling %q", req.URL.String())
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	return nil
}

func (s *Sender) setUserAgent(req *http.Request) {
	userAgentKey := "User-Agent"
	if req.Header.Get(userAgentKey) == "" {
		req.Header.Add(userAgentKey, s.userAgent)
	}
}

func getRootCAs(certificates []string) *x509.CertPool {
	caCertPool := x509.NewCertPool()

	for _, filename := range certificates {
		caCert, err := os.ReadFile(filename)
		if err != nil {
			clog.Warningf("cannot load CA certificate: %s", err)
			continue
		}
		if !caCertPool.AppendCertsFromPEM(caCert) {
			clog.Warningf("invalid certificate: %q", filename)
		}
	}
	return caCertPool
}

func resolve(body string, ctx Context) string {
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

func loadBodyTemplate(filename string, ctx Context) (string, error) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		return "", err
	}
	buffer := &bytes.Buffer{}
	err = tmpl.Execute(buffer, ctx)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
