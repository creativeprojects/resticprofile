package systemd

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSystemUnit(t *testing.T) {
	testCases := []struct {
		config Config
	}{
		{
			config: Config{
				CommandLine:      "/bin/resticprofile --config profiles.yaml run-schedule backup@profile1",
				WorkingDirectory: "/workdir",
				Title:            "profile1",
				SubTitle:         "backup",
				JobDescription:   "job description",
				TimerDescription: "timer description",
				Schedules:        []string{"daily"},
				UnitType:         SystemUnit,
				Priority:         "low",
			},
		},
		{
			config: Config{
				CommandLine:      "/bin/resticprofile --no-ansi --config profiles.yaml run-schedule check@profile2",
				WorkingDirectory: "/workdir",
				Title:            "profile2",
				SubTitle:         "check",
				JobDescription:   "job description",
				TimerDescription: "timer description",
				Schedules:        []string{"weekly"},
				UnitType:         UserUnit,
				Priority:         "low",
			},
		},
	}

	fs = afero.NewMemMapFs()

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			baseUnit := fmt.Sprintf("resticprofile-%s@profile-%s", tc.config.SubTitle, tc.config.Title)
			serviceFile := baseUnit + ".service"

			err := Generate(tc.config)
			require.NoError(t, err)

			readCfg, err := Read(serviceFile, tc.config.UnitType)
			require.NoError(t, err)
			assert.NotNil(t, readCfg)

			expected := &Config{
				Title:            tc.config.Title,
				SubTitle:         tc.config.SubTitle,
				JobDescription:   tc.config.JobDescription,
				WorkingDirectory: tc.config.WorkingDirectory,
				CommandLine:      tc.config.CommandLine,
				UnitType:         tc.config.UnitType,
			}
			assert.Equal(t, expected, readCfg)
		})
	}
}
