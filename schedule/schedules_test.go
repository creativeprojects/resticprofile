package schedule

import (
	"bytes"
	"os"
	"runtime"
	"testing"

	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
)

func TestParseEmptySchedules(t *testing.T) {
	_, err := parseSchedules([]string{})
	assert.NoError(t, err)
}

func TestParseSchedulesWithEmpty(t *testing.T) {
	_, err := parseSchedules([]string{""})
	assert.Error(t, err)
}

func TestParseSchedulesWithError(t *testing.T) {
	_, err := parseSchedules([]string{"parse error"})
	assert.Error(t, err)
}

func TestParseScheduleDaily(t *testing.T) {
	events, err := parseSchedules([]string{"daily"})
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "daily", events[0].Input())
	assert.Equal(t, "*-*-* 00:00:00", events[0].String())
}

func TestDisplayParseSchedules(t *testing.T) {
	events, err := parseSchedules([]string{"daily"})
	assert.NoError(t, err)

	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	defer term.SetOutput(os.Stdout)

	displayParsedSchedules("command", events)
	output := buffer.String()
	assert.Contains(t, output, "Original form: daily\n")
	assert.Contains(t, output, "Normalized form: *-*-* 00:00:00\n")
}

func TestDisplaySystemdSchedulesWithEmpty(t *testing.T) {
	err := displaySystemdSchedules("command", []string{""})
	assert.Error(t, err)
}

func TestDisplaySystemdSchedules(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip()
	}

	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	defer term.SetOutput(os.Stdout)

	err := displaySystemdSchedules("command", []string{"daily"})
	assert.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "Original form: daily\n")
	assert.Contains(t, output, "Normalized form: *-*-* 00:00:00\n")
}

func TestDisplaySystemdSchedulesError(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Skip()
	}
	err := displaySystemdSchedules("command", []string{"daily"})
	assert.Error(t, err)
}
