//+build darwin windows

package schedule

//
// Common code for Mac OS and Windows
//

import "github.com/creativeprojects/resticprofile/calendar"

func (j *Job) loadSchedules(command string, schedules []string) ([]*calendar.Event, error) {
	return loadParsedSchedules(command, schedules)
}
