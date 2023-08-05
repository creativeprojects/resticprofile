package schedule

import (
	"os/exec"
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

	term.StartRecording(term.RecordOutput)
	defer term.StopRecording()

	displayParsedSchedules("command", events)
	output := term.ReadRecording()
	assert.Contains(t, output, "Original form: daily\n")
	assert.Contains(t, output, "Normalized form: *-*-* 00:00:00\n")
}

func TestDisplayParseSchedulesIndexAndTotal(t *testing.T) {
	events, err := parseSchedules([]string{"daily", "monthly", "yearly"})
	assert.NoError(t, err)

	term.StartRecording(term.RecordOutput)
	defer term.StopRecording()

	displayParsedSchedules("command", events)
	output := term.ReadRecording()
	assert.Contains(t, output, "schedule 1/3")
	assert.Contains(t, output, "schedule 2/3")
	assert.Contains(t, output, "schedule 3/3")
}

func TestDisplaySystemdSchedulesWithEmpty(t *testing.T) {
	err := displaySystemdSchedules("command", []string{""})
	assert.Error(t, err)
}

func TestDisplaySystemdSchedules(t *testing.T) {
	_, err := exec.LookPath("systemd-analyze")
	if err != nil {
		t.Skip("systemd-analyze not available")
	}

	term.StartRecording(term.RecordOutput)
	defer term.StopRecording()

	err = displaySystemdSchedules("command", []string{"daily"})
	assert.NoError(t, err)

	output := term.ReadRecording()
	assert.Contains(t, output, "Original form: daily")
	assert.Contains(t, output, "Normalized form: *-*-* 00:00:00")
}

func TestDisplaySystemdSchedulesError(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Skip()
	}
	err := displaySystemdSchedules("command", []string{"daily"})
	assert.Error(t, err)
}
