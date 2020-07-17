package term

import (
	"io"

	"github.com/creativeprojects/resticprofile/remote"
)

// RemoteTerm is used to send terminal output remotely
type RemoteTerm struct {
	client *remote.Client
}

// NewRemoteTerm creates a new RemoteTerm based on a remote.Client
func NewRemoteTerm(client *remote.Client) *RemoteTerm {
	return &RemoteTerm{
		client: client,
	}
}

// Write terminal data remotely
func (t *RemoteTerm) Write(p []byte) (n int, err error) {
	n = len(p)
	err = t.client.Term(p)
	return n, err
}

// Verify interface
var (
	_ io.Writer = &RemoteTerm{}
)
