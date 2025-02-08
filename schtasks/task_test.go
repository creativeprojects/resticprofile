package schtasks

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadXMLTask(t *testing.T) {
	filenames, err := filepath.Glob("examples/*.xml")
	require.NoError(t, err)
	for _, filename := range filenames {
		file, err := os.Open(filename)
		require.NoError(t, err)
		defer file.Close()

		decoder := xml.NewDecoder(file)
		decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
			// no need for character conversion
			return input, nil
		}
		task := Task{}
		err = decoder.Decode(&task)
		require.NoErrorf(t, err, "filename: %s", filename)
		assert.True(t, len(task.Triggers.CalendarTrigger) > 0 || len(task.Triggers.TimeTrigger) > 0)

		// t.Logf("%+v", task)
	}
}

func TestSaveXMLTask(t *testing.T) {
	file, err := os.Create("output.xml")
	require.NoError(t, err)
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	task := NewTask()
	task.RegistrationInfo.Author = constants.ApplicationName
	task.Principals.Principal.LogonType = LogonTypePassword
	task.Actions.Exec = []ExecAction{
		{
			Command:   "echo",
			Arguments: "Hello World!",
		},
	}
	task.Triggers.CalendarTrigger = []CalendarTrigger{
		{
			Enabled: true,
			ScheduleByDay: &ScheduleByDay{
				DaysInterval: 1,
			},
		},
	}
	err = encoder.Encode(&task)
	require.NoError(t, err)
}

func TestSaveXMLTaskUsingServiceAccount(t *testing.T) {
	file, err := os.Create("output.xml")
	require.NoError(t, err)
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	task := NewTask()
	task.RegistrationInfo.Author = constants.ApplicationName
	task.Principals.Principal.UserId = ServiceAccount
	task.Principals.Principal.RunLevel = RunLevelLeastPrivilege
	task.Principals.Principal.LogonType = ""
	task.Actions.Exec = []ExecAction{
		{
			Command:   "echo",
			Arguments: "Hello World!",
		},
	}
	err = encoder.Encode(&task)
	require.NoError(t, err)
}
