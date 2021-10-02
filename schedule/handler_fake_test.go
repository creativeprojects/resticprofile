package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
)

const (
	errorMethodNotRegistered = "method not registered in mock"
)

type mockHandler struct {
	t                      *testing.T
	init                   func() error
	close                  func()
	parseSchedules         func(schedules []string) ([]*calendar.Event, error)
	displayParsedSchedules func(command string, events []*calendar.Event)
	displaySchedules       func(command string, schedules []string) error
	displayStatus          func(profileName string) error
	createJob              func(job JobConfig, schedules []*calendar.Event, permission string) error
	removeJob              func(job JobConfig, permission string) error
	displayJobStatus       func(job JobConfig) error
}

func (h mockHandler) Init() error {
	if h.init == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	return h.init()
}

func (h mockHandler) Close() {
	if h.close == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	h.close()
}

func (h mockHandler) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	if h.parseSchedules == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	return h.parseSchedules(schedules)
}

func (h mockHandler) DisplayParsedSchedules(command string, events []*calendar.Event) {
	if h.displayParsedSchedules == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	h.displayParsedSchedules(command, events)
}

func (h mockHandler) DisplaySchedules(command string, schedules []string) error {
	if h.displaySchedules == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	return h.displaySchedules(command, schedules)
}

func (h mockHandler) DisplayStatus(profileName string) error {
	if h.displayStatus == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	return h.displayStatus(profileName)
}

func (h mockHandler) CreateJob(job JobConfig, schedules []*calendar.Event, permission string) error {
	if h.createJob == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	return h.createJob(job, schedules, permission)
}

func (h mockHandler) RemoveJob(job JobConfig, permission string) error {
	if h.removeJob == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	return h.removeJob(job, permission)
}

func (h mockHandler) DisplayJobStatus(job JobConfig) error {
	if h.displayJobStatus == nil {
		h.t.Fatal(errorMethodNotRegistered)
	}
	return h.displayJobStatus(job)
}

var (
	_ Handler = &mockHandler{}
)
