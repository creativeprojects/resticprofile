package schedule

import (
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
)

func TestCreateJobHappyPathSystemd(t *testing.T) {
	counter := 0
	handler := mockHandler{
		t: t,
		parseSchedules: func(schedules []string) ([]*calendar.Event, error) {
			counter |= 1
			return nil, nil
		},
		displaySchedules: func(command string, schedules []string) error {
			counter |= 2
			return nil
		},
		createJob: func(job *Config, schedules []*calendar.Event, permission string) error {
			counter |= 4
			return nil
		},
	}
	job := Job{
		config:  &Config{},
		handler: handler,
	}
	err := job.Create()
	assert.NoError(t, err)

	assert.Equal(t, 1|2|4, counter)
}

func TestCreateJobHappyPathOther(t *testing.T) {
	counter := 0
	handler := mockHandler{
		t: t,
		parseSchedules: func(schedules []string) ([]*calendar.Event, error) {
			counter |= 1
			return []*calendar.Event{calendar.NewEvent()}, nil
		},
		displayParsedSchedules: func(command string, events []*calendar.Event) {
			counter |= 2
		},
		createJob: func(job *Config, schedules []*calendar.Event, permission string) error {
			counter |= 4
			return nil
		},
	}
	job := Job{
		config:  &Config{},
		handler: handler,
	}
	err := job.Create()
	assert.NoError(t, err)

	assert.Equal(t, 1|2|4, counter)
}

func TestCreateJobSadPath1(t *testing.T) {
	counter := 0
	handler := mockHandler{
		t: t,
		parseSchedules: func(schedules []string) ([]*calendar.Event, error) {
			counter |= 1
			return nil, errors.New("test!")
		},
	}
	job := Job{
		config:  &Config{},
		handler: handler,
	}
	err := job.Create()
	assert.Error(t, err)

	assert.Equal(t, 1, counter)
}

func TestCreateJobSadPath2(t *testing.T) {
	counter := 0
	handler := mockHandler{
		t: t,
		parseSchedules: func(schedules []string) ([]*calendar.Event, error) {
			counter |= 1
			return nil, nil
		},
		displaySchedules: func(command string, schedules []string) error {
			counter |= 2
			return errors.New("test!")
		},
	}
	job := Job{
		config:  &Config{},
		handler: handler,
	}
	err := job.Create()
	assert.Error(t, err)

	assert.Equal(t, 1|2, counter)
}

func TestCreateJobSadPath3(t *testing.T) {
	counter := 0
	handler := mockHandler{
		t: t,
		parseSchedules: func(schedules []string) ([]*calendar.Event, error) {
			counter |= 1
			return nil, nil
		},
		displaySchedules: func(command string, schedules []string) error {
			counter |= 2
			return nil
		},
		createJob: func(job *Config, schedules []*calendar.Event, permission string) error {
			counter |= 4
			return errors.New("test!")
		},
	}
	job := Job{
		config:  &Config{},
		handler: handler,
	}
	err := job.Create()
	assert.Error(t, err)

	assert.Equal(t, 1|2|4, counter)
}
