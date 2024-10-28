package schedule

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEmptySchedules(t *testing.T) {
	_, err := parseSchedules([]string{})
	require.NoError(t, err)
}

func TestParseSchedulesWithEmpty(t *testing.T) {
	_, err := parseSchedules([]string{""})
	require.Error(t, err)
}

func TestParseSchedulesWithError(t *testing.T) {
	_, err := parseSchedules([]string{"parse error"})
	require.Error(t, err)
}

func TestParseScheduleDaily(t *testing.T) {
	events, err := parseSchedules([]string{"daily"})
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "daily", events[0].Input())
	assert.Equal(t, "*-*-* 00:00:00", events[0].String())
}

func TestDisplayParseSchedules(t *testing.T) {
	events, err := parseSchedules([]string{"daily"})
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	defer term.SetOutput(os.Stdout)

	displayParsedSchedules("profile", "command", events)
	output := buffer.String()
	assert.Contains(t, output, "Original form: daily\n")
	assert.Contains(t, output, "Normalized form: *-*-* 00:00:00\n")
}

func TestDisplayParseSchedulesIndexAndTotal(t *testing.T) {
	events, err := parseSchedules([]string{"daily", "monthly", "yearly"})
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	defer term.SetOutput(os.Stdout)

	displayParsedSchedules("profile", "command", events)
	output := buffer.String()
	assert.Contains(t, output, "schedule 1/3")
	assert.Contains(t, output, "schedule 2/3")
	assert.Contains(t, output, "schedule 3/3")
}

func TestDisplaySystemdSchedulesWithEmpty(t *testing.T) {
	err := displaySystemdSchedules("profile", "command", []string{""})
	require.Error(t, err)
}

func TestDisplaySystemdSchedules(t *testing.T) {
	_, err := exec.LookPath("systemd-analyze")
	if err != nil {
		t.Skip("systemd-analyze not available")
	}

	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	defer term.SetOutput(os.Stdout)

	err = displaySystemdSchedules("profile", "command", []string{"daily"})
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "Original form: daily")
	assert.Contains(t, output, "Normalized form: *-*-* 00:00:00")
}

func TestDisplaySystemdSchedulesError(t *testing.T) {
	if platform.IsWindows() || platform.IsDarwin() {
		t.Skip("test only runs on linux")
	}
	err := displaySystemdSchedules("profile", "command", []string{"daily"})
	require.Error(t, err)
}
