//go:build windows

package schtasks

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const listOutput = `
Folder: \resticprofile backup
HostName:                             WIN-5NMJF0VS8OR
TaskName:                             \resticprofile backup\self backup
Next Run Time:                        04/08/2025 21:15:00
Status:                               Ready
Logon Mode:                           Interactive only
Last Run Time:                        04/08/2025 21:00:01
Last Result:                          1
Author:                               resticprofile
Task To Run:                          G:\go\src\github.com\creativeprojects\resticprofile\resticprofile.exe --no-ansi --config "examples\othere windows.yaml" run-schedule backup@self
Start In:                             G:\go\src\github.com\creativeprojects\resticprofile
Comment:                              resticprofile backup for profile self in examples\othere windows.yaml
Scheduled Task State:                 Enabled
Idle Time:                            Disabled
Power Management:                     Stop On Battery Mode, No Start On Batteries
Run As User:                          Fred
Delete Task If Not Rescheduled:       Disabled
Stop Task If Runs X Hours and X Mins: Disabled
Schedule:                             Scheduling data is not available in this format.
Schedule Type:                        Weekly
Start Time:                           20:15:00
Start Date:                           04/08/2025
End Date:                             N/A
Days:                                 MON, TUE, WED, THU, FRI
Months:                               Every 1 week(s)
Repeat: Every:                        0 Hour(s), 15 Minute(s)
Repeat: Until: Time:                  None
Repeat: Until: Duration:              23 Hour(s), 45 Minute(s)
Repeat: Stop If Still Running:        Disabled

HostName:                             WIN-5NMJF0VS8OR
TaskName:                             \resticprofile backup\self backup
Next Run Time:                        04/08/2025 21:15:00
Status:                               Ready
Logon Mode:                           Interactive only
Last Run Time:                        04/08/2025 21:00:01
Last Result:                          1
Author:                               resticprofile
Task To Run:                          G:\go\src\github.com\creativeprojects\resticprofile\resticprofile.exe --no-ansi --config "examples\othere windows.yaml" run-schedule backup@self
Start In:                             G:\go\src\github.com\creativeprojects\resticprofile
Comment:                              resticprofile backup for profile self in examples\othere windows.yaml
Scheduled Task State:                 Enabled
Idle Time:                            Disabled
Power Management:                     Stop On Battery Mode, No Start On Batteries
Run As User:                          Fred
Delete Task If Not Rescheduled:       Disabled
Stop Task If Runs X Hours and X Mins: Disabled
Schedule:                             Scheduling data is not available in this format.
Schedule Type:                        Weekly
Start Time:                           00:00:00
Start Date:                           09/08/2025
End Date:                             N/A
Days:                                 SUN, SAT
Months:                               Every 1 week(s)
Repeat: Every:                        12 Hour(s), 0 Minute(s)
Repeat: Until: Time:                  None
Repeat: Until: Duration:              12 Hour(s), 0 Minute(s)
Repeat: Stop If Still Running:        Disabled
`

func TestReadingListOutput(t *testing.T) {
	t.Parallel()

	output, err := getTaskInfoFromList(bytes.NewBufferString(listOutput))
	require.NoError(t, err)
	assert.NotNil(t, output)

	expected := []map[string]string{
		{
			"Folder":                               "\\resticprofile backup",
			"HostName":                             "WIN-5NMJF0VS8OR",
			"TaskName":                             "\\resticprofile backup\\self backup",
			"Next Run Time":                        "04/08/2025 21:15:00",
			"Status":                               "Ready",
			"Logon Mode":                           "Interactive only",
			"Last Run Time":                        "04/08/2025 21:00:01",
			"Last Result":                          "1",
			"Author":                               "resticprofile",
			"Task To Run":                          "G:\\go\\src\\github.com\\creativeprojects\\resticprofile\\resticprofile.exe --no-ansi --config \"examples\\othere windows.yaml\" run-schedule backup@self",
			"Start In":                             "G:\\go\\src\\github.com\\creativeprojects\\resticprofile",
			"Comment":                              "resticprofile backup for profile self in examples\\othere windows.yaml",
			"Scheduled Task State":                 "Enabled",
			"Idle Time":                            "Disabled",
			"Power Management":                     "Stop On Battery Mode, No Start On Batteries",
			"Run As User":                          "Fred",
			"Delete Task If Not Rescheduled":       "Disabled",
			"Stop Task If Runs X Hours and X Mins": "Disabled",
			"Schedule":                             "Scheduling data is not available in this format.",
			"Schedule Type":                        "Weekly",
			"Start Time":                           "20:15:00",
			"Start Date":                           "04/08/2025",
			"End Date":                             "N/A",
			"Days":                                 "MON, TUE, WED, THU, FRI",
			"Months":                               "Every 1 week(s)",
			"Repeat: Every":                        "0 Hour(s), 15 Minute(s)",
			"Repeat: Until: Time":                  "None",
			"Repeat: Until: Duration":              "23 Hour(s), 45 Minute(s)",
			"Repeat: Stop If Still Running":        "Disabled",
		},
		{
			"Folder":                               "\\resticprofile backup",
			"HostName":                             "WIN-5NMJF0VS8OR",
			"TaskName":                             "\\resticprofile backup\\self backup",
			"Next Run Time":                        "04/08/2025 21:15:00",
			"Status":                               "Ready",
			"Logon Mode":                           "Interactive only",
			"Last Run Time":                        "04/08/2025 21:00:01",
			"Last Result":                          "1",
			"Author":                               "resticprofile",
			"Task To Run":                          "G:\\go\\src\\github.com\\creativeprojects\\resticprofile\\resticprofile.exe --no-ansi --config \"examples\\othere windows.yaml\" run-schedule backup@self",
			"Start In":                             "G:\\go\\src\\github.com\\creativeprojects\\resticprofile",
			"Comment":                              "resticprofile backup for profile self in examples\\othere windows.yaml",
			"Scheduled Task State":                 "Enabled",
			"Idle Time":                            "Disabled",
			"Power Management":                     "Stop On Battery Mode, No Start On Batteries",
			"Run As User":                          "Fred",
			"Delete Task If Not Rescheduled":       "Disabled",
			"Stop Task If Runs X Hours and X Mins": "Disabled",
			"Schedule":                             "Scheduling data is not available in this format.",
			"Schedule Type":                        "Weekly",
			"Start Time":                           "00:00:00",
			"Start Date":                           "09/08/2025",
			"End Date":                             "N/A",
			"Days":                                 "SUN, SAT",
			"Months":                               "Every 1 week(s)",
			"Repeat: Every":                        "12 Hour(s), 0 Minute(s)",
			"Repeat: Until: Time":                  "None",
			"Repeat: Until: Duration":              "12 Hour(s), 0 Minute(s)",
			"Repeat: Stop If Still Running":        "Disabled",
		},
	}
	assert.Equal(t, expected, output)
}
