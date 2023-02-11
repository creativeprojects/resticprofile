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
	"regexp"
	"strings"
	"text/tabwriter"
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
	dryRun         bool
}

func NewSender(certificates []string, userAgent string, timeout time.Duration, dryRun bool) *Sender {
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
		dryRun:         dryRun,
	}
}

func (s *Sender) Send(cfg config.SendMonitoringSection, ctx Context) error {
	if cfg.URL.Value() == "" {
		return errors.New("URL field is empty")
	}
	url := resolve(cfg.URL.Value(), ctx)
	publicUrl := resolve(cfg.URL.String(), ctx)
	method := cfg.Method
	if method == "" {
		method = http.MethodGet
	}
	var (
		body       string    // only used in dry-run mode
		bodyReader io.Reader = http.NoBody
	)
	if cfg.BodyTemplate != "" {
		bodyTemplate, err := loadBodyTemplate(cfg.BodyTemplate, ctx)
		if err != nil {
			return err
		}
		body = bodyTemplate
		bodyReader = bytes.NewBufferString(body)
	}
	if cfg.Body != "" {
		body = resolve(cfg.Body, ctx)
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}
	for _, header := range cfg.Headers {
		if header.Name == "" {
			continue
		}
		req.Header.Add(header.Name, header.Value.Value())
	}
	s.setUserAgent(req)

	client := s.client
	if cfg.SkipTLS {
		client = s.insecureClient
	}

	if s.dryRun {
		clog.Infof("dry-run: webhook request method=%s url=%q headers:\n%s", method, publicUrl, s.stringifyHeaders(req.Header, cfg.Headers))
		if len(body) > 0 {
			clog.Infof("dry-run: webhook request body:\n%s", body)
		}
		return nil
	}

	clog.Debugf("calling: %s %q\n%s", method, publicUrl, s.stringifyHeaders(req.Header, cfg.Headers))
	if len(body) > 0 {
		clog.Debugf("request body:\n%s", body)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	s.logResponse(publicUrl, resp)

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	return nil
}

var responseContentSanitizer = regexp.MustCompile(`(?i)[^\d\w\s.,:;_*+\-=?!"'$%&§/\\\[\](){}<>]+`)

func (s *Sender) logResponse(url string, resp *http.Response) {
	clog.Debugf("%q returned: %s\n%s", url, resp.Status, s.stringifyHeaders(resp.Header, nil))

	if content, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024)); len(content) > 0 {
		content = responseContentSanitizer.ReplaceAll(content, []byte(" "))
		clog.Tracef("response body (sanitized):\n%s", string(content))
	}
}

func (s *Sender) stringifyHeaders(headers http.Header, config []config.SendMonitoringHeader) string {
	buf := &strings.Builder{}
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	for name, values := range headers {
		// Translate values to confidential replacement
		for i, value := range values {
			for _, ch := range config {
				if ch.Value.IsConfidential() && ch.Value.Value() == value {
					values[i] = ch.Value.String()
				}
			}
		}
		// Print header
		_, _ = fmt.Fprintf(w, "%s:\t%s\n", name, strings.Join(values, "; "))
	}
	_ = w.Flush()
	return buf.String()
}

func (s *Sender) setUserAgent(req *http.Request) {
	userAgentKey := "User-Agent"
	if req.Header.Get(userAgentKey) == "" {
		req.Header.Add(userAgentKey, s.userAgent)
	}
}

func getRootCAs(certificates []string) *x509.CertPool {
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		caCertPool = x509.NewCertPool()
	}

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
