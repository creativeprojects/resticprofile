package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client for sending messages back to the parent process
type Client struct {
	baseURL string
	client  *http.Client
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

// Log messages back to the parent process
func (c *Client) Log(level int, message string) error {
	log := remoteLog{
		Level:   level,
		Message: message,
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

// Done signals to the parent process that we're finished
func (c *Client) Done() error {
	resp, err := c.client.Get(c.baseURL + donePath)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	return nil
}
