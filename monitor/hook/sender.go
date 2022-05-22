package hook

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/creativeprojects/resticprofile/config"
)

type Sender struct {
	client *http.Client
}

func NewSender(timeout time.Duration) *Sender {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
	}
	return &Sender{
		client: client,
	}
}

func (s *Sender) Send(cfg config.SendMonitorSection) error {
	if cfg.URL == "" {
		return errors.New("URL field is empty")
	}
	method := cfg.Method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader = http.NoBody
	if cfg.Body != "" {
		body = bytes.NewBufferString(cfg.Body)
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
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
