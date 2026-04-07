package term

import (
	"bytes"
	"io"
	"sync"

	"github.com/creativeprojects/resticprofile/util"
)

type outputRecording struct {
	lock          sync.Mutex
	buffer        *bytes.Buffer
	writer        io.Writer
	output, error io.Writer
}

type RecordMode uint8

const (
	RecordOutput RecordMode = iota
	RecordError
	RecordBoth
)

func (r *outputRecording) StartRecording(mode RecordMode) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.buffer != nil {
		return
	}

	r.buffer = new(bytes.Buffer)
	r.writer = util.NewSyncWriterMutex(r.buffer, &r.lock)

	if mode != RecordError {
		r.output = GetOutput()
		setOutput(r.writer)
	}
	if mode != RecordOutput {
		r.error = GetErrorOutput()
		setErrorOutput(r.writer)
	}
}

func (r *outputRecording) ReadRecording() (content string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.buffer != nil {
		content = r.buffer.String()
		r.buffer.Reset()
	}
	return
}

func (r *outputRecording) StopRecording() (content string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.buffer != nil {
		if r.output != nil && r.writer == GetOutput() {
			setOutput(r.output)
			r.output = nil
		}

		if r.error != nil && r.writer == GetErrorOutput() {
			setErrorOutput(r.error)
			r.error = nil
		}

		content = r.buffer.String()
		r.writer = nil
		r.buffer = nil
	}
	return
}
