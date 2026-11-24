package prom

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitCode(t *testing.T) {
	if os.Getenv("RESTICPROFILE_PROM_TEST_EXIT") != "" {
		os.Exit(11)
	}
	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestExitCode$")
	cmd.Env = append(os.Environ(), "RESTICPROFILE_PROM_TEST_EXIT=1")
	err := cmd.Run()
	require.Error(t, err)

	assert.Equal(t, 0, exitCode(nil))
	assert.Equal(t, 11, exitCode(err))
	assert.Equal(t, 11, exitCode(fmt.Errorf("check on profile 'test': %w", err)))
	assert.Equal(t, 3, exitCode(&monitor.InternalWarningError{}))
	assert.Equal(t, -1, exitCode(errors.New("not started")))
}

func TestProgressSummary(t *testing.T) {
	testCases := []struct {
		command  string
		expected string
	}{
		{command: "check", expected: "resticprofile_check_status"},
		{command: "copy", expected: "resticprofile_copy_status"},
		{command: "forget", expected: "resticprofile_forget_status"},
		{command: "retention", expected: "resticprofile_forget_status"},
		{command: "prune", expected: "resticprofile_prune_status"},
		{command: "backup", expected: "resticprofile_backup_status"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.command, func(t *testing.T) {
			profile := config.NewProfile(nil, "test")
			profile.PrometheusSaveToFile = filepath.Join(t.TempDir(), "test.prom")
			progress := NewProgress(profile, NewMetrics("test", "", "", "", nil))

			progress.Summary(testCase.command, monitor.Summary{Duration: time.Second}, "", nil)

			content := readFile(t, profile.PrometheusSaveToFile)
			assert.Contains(t, content, testCase.expected+`{profile="test"} 2`+"\n")
		})
	}
}

func TestProgressSummaryNotReported(t *testing.T) {
	profile := config.NewProfile(nil, "test")
	profile.PrometheusSaveToFile = filepath.Join(t.TempDir(), "test.prom")
	progress := NewProgress(profile, NewMetrics("test", "", "", "", nil))

	progress.Summary("unlock", monitor.Summary{}, "", nil)

	assert.NoFileExists(t, profile.PrometheusSaveToFile)
}
