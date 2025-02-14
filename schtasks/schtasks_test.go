//go:build windows

package schtasks

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const csvOutput = `"HostName","TaskName","Next Run Time","Status","Logon Mode","Last Run Time","Last Result","Author","Task To Run","Start In","Comment","Scheduled Task State","Idle Time","Power Management","Run As User","Delete Task If Not Rescheduled","Stop Task If Runs X Hours and X Mins","Schedule","Schedule Type","Start Time","Start Date","End Date","Days","Months","Repeat: Every","Repeat: Until: Time","Repeat: Until: Duration","Repeat: Stop If Still Running"
"WIN-7060","\resticprofile backup\self backup","13/02/2025 17:30:00","Ready","Interactive only","13/02/2025 17:15:00","-2147024894","resticprofile","R:\Temp\go-build4112622634\b001\exe\resticprofile.exe --no-ansi --config examples\windows.yaml run-schedule backup@self","M:\go\src\github.com\creativeprojects\resticprofile","resticprofile backup for profile self in examples\windows.yaml","Enabled","Disabled","Stop On Battery Mode, No Start On Batteries","user","Disabled","72:00:00","Scheduling data is not available in this format.","Weekly","17:15:00","13/02/2025","N/A","MON, TUE, WED, THU, FRI","Every 1 week(s)","0 Hour(s), 15 Minute(s)","None","23 Hour(s), 45 Minute(s)","Disabled"
"WIN-7060","\resticprofile backup\self backup","13/02/2025 17:30:00","Ready","Interactive only","13/02/2025 17:15:00","-2147024894","resticprofile","R:\Temp\go-build4112622634\b001\exe\resticprofile.exe --no-ansi --config examples\windows.yaml run-schedule backup@self","M:\go\src\github.com\creativeprojects\resticprofile","resticprofile backup for profile self in examples\windows.yaml","Enabled","Disabled","Stop On Battery Mode, No Start On Batteries","user","Disabled","72:00:00","Scheduling data is not available in this format.","Weekly","00:00:00","15/02/2025","N/A","SUN, SAT","Every 1 week(s)","12 Hour(s), 0 Minute(s)","None","12 Hour(s), 0 Minute(s)","Disabled"
`

func TestReadingCSVOutput(t *testing.T) {
	t.Parallel()

	output, err := getCSV(bytes.NewBufferString(csvOutput))
	require.NoError(t, err)
	assert.Len(t, output, 3)     // lines
	assert.Len(t, output[0], 28) // fields
}

func TestReadTaskInfoError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		taskname string
		expected error
	}{
		{``, ErrEmptyTaskName},
		{`\\something\invalid`, ErrInvalidTaskName},
		{`\most likely\does\not\exist`, ErrNotRegistered},
		{`most likely\does\not\exist`, ErrNotRegistered},
		{`\most likely does not exist`, ErrNotRegistered},
		{`most likely does not exist`, ErrNotRegistered},
	}

	for _, tc := range testCases {
		t.Run(tc.taskname, func(t *testing.T) {
			output := &bytes.Buffer{}
			err := readTaskInfo(tc.taskname, output)
			require.ErrorIs(t, err, tc.expected)
			assert.Empty(t, output.Len())
		})
	}
}

func TestSchTasksError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		message  string
		expected error
	}{
		{`ERROR: The system cannot find the file specified.\r\n`, ErrNotRegistered},
		{`ERROR: The system cannot find the path] specified.\r\n`, ErrNotRegistered},
		{`ERROR: The specified task name "\resticprofile backup\toto" does not exist in the system.\r\n`, ErrNotRegistered},
		{`ERROR: The filename, directory name, or volume label syntax is incorrect.\r\n`, ErrInvalidTaskName},
		{`ERROR: Access is denied.\r\n`, ErrAccessDenied},
		{`ERROR: Cannot create a file when that file already exists.\r\n`, ErrAlreadyExist},
	}
	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			require.ErrorIs(t, schTasksError(tc.message), tc.expected)
		})
	}
}
