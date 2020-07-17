package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/creativeprojects/clog"
)

// Client for sending messages back to the parent process
type Client struct {
	baseURL   string
	client    *http.Client
	logPrefix string
}

type remoteLog struct {
	Level   int    `json:"level"`
	Message string `json:"message"`
}

// NewClient creates a new client to connect to localhost and port in parameter
func NewClient(port int) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		client:  &http.Client{},
	}
}

// SetLogPrefix adds a prefix to all the log messages
func (c *Client) SetLogPrefix(logPrefix string) {
	c.logPrefix = logPrefix
}

// LogEntry logs messages back to the parent process
func (c *Client) LogEntry(logEntry clog.LogEntry) error {
	log := remoteLog{
		Level:   int(logEntry.Level),
		Message: c.logPrefix + logEntry.GetMessage(),
	}
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.Encode(log)
	resp, err := c.client.Post(c.baseURL+logPath, "application/json", buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Term sends plain text terminal output
func (c *Client) Term(p []byte) error {
	buffer := bytes.NewBuffer(p)
	resp, err := c.client.Post(c.baseURL+termPath, "text/plain", buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Done signals to the parent process that we're finished
func (c *Client) Done() error {
	resp, err := c.client.Get(c.baseURL + donePath)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	return nil
}
