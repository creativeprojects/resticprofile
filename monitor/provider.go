package monitor

import (
	"sync"
)

// Provider declares callbacks to provide status and final summary of restic commands
type Provider interface {
	// ProvideStatus is called by an output scanner when restic provided the current status.
	// The status update must happen inside the callback function which may run under a mutex
	ProvideStatus(func(status *Status))
	// UpdateSummary is called by an output scanner when restic provided an update to the final summary.
	// The summary update must happen inside the callback function which may run under a mutex
	UpdateSummary(func(summary *Summary))
	// CurrentSummary returns a copy of the current known summary
	CurrentSummary() Summary
	// ProvideSummary is called when the restic command finished to finalize the summary.
	ProvideSummary(command string, stderr string, result error)
}

// NewProvider creates a new provider that forwards status and summary to the specified receivers
func NewProvider(receiver ...Receiver) Provider {
	return &provider{receiver: receiver}
}

type provider struct {
	mutex    sync.Mutex
	summary  Summary
	receiver []Receiver
	final    bool
}

func (p *provider) ProvideStatus(fn func(status *Status)) {
	status := new(Status)
	fn(status)
	for _, receiver := range p.receiver {
		receiver.Status(*status)
	}
}

func (p *provider) UpdateSummary(fn func(summary *Summary)) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if !p.final {
		fn(&p.summary)
	}
}

func (p *provider) CurrentSummary() Summary {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.summary
}

func (p *provider) ProvideSummary(command string, stderr string, result error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if !p.final {
		p.final = true
		for _, receiver := range p.receiver {
			receiver.Summary(command, p.summary, stderr, result)
		}
	}
}
