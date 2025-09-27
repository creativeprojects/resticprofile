//go:build darwin

package darwin

import "github.com/creativeprojects/resticprofile/calendar"

const (
	intervalMinute  = "Minute"
	intervalHour    = "Hour"
	intervalWeekday = "Weekday"
	intervalDay     = "Day"
	intervalMonth   = "Month"
)

// CalendarInterval contains date and time trigger definition inside a map.
// keys of the map should be:
//
//	"Month"   Month of year (1..12, 1 being January)
//	"Day"     Day of month (1..31)
//	"Weekday" Day of week (0..7, 0 and 7 being Sunday)
//	"Hour"    Hour of day (0..23)
//	"Minute"  Minute of hour (0..59)
type CalendarInterval map[string]int

// newCalendarInterval creates a new map of 5 elements
func newCalendarInterval() *CalendarInterval {
	var value CalendarInterval = make(map[string]int, 5)
	return &value
}

func (c *CalendarInterval) clone() *CalendarInterval {
	clone := newCalendarInterval()
	for key, value := range *c {
		(*clone)[key] = value
	}
	return clone
}

// GetCalendarIntervalsFromSchedules converts schedules into launchd calendar events
// let's say we've setup these rules:
//
//	Mon-Fri *-*-* *:0,30:00  = every half hour
//	Sat     *-*-* 0,12:00:00 = twice a day on saturday
//	        *-*-01 *:*:*     = the first of each month
//
// it should translate as:
// 1st rule
//
//	Weekday = Monday, Minute = 0
//	Weekday = Monday, Minute = 30
//	... same from Tuesday to Thurday
//	Weekday = Friday, Minute = 0
//	Weekday = Friday, Minute = 30
//
// Total of 10 rules
// 2nd rule
//
//	Weekday = Saturday, Hour = 0
//	Weekday = Saturday, Hour = 12
//
// Total of 2 rules
// 3rd rule
//
//	Day = 1
//
// Total of 1 rule
func GetCalendarIntervalsFromSchedules(schedules []*calendar.Event) []CalendarInterval {
	entries := make([]CalendarInterval, 0, len(schedules))
	for _, schedule := range schedules {
		entries = append(entries, getCalendarIntervalsFromScheduleTree(generateTreeOfSchedules(schedule))...)
	}
	return entries
}

func getCalendarIntervalsFromScheduleTree(tree []*treeElement) []CalendarInterval {
	entries := make([]CalendarInterval, 0)
	for _, element := range tree {
		// creates a new calendar entry for each tip of the branch
		newEntry := newCalendarInterval()
		fillInValueFromScheduleTreeElement(newEntry, element, &entries)
	}
	return entries
}

func fillInValueFromScheduleTreeElement(currentEntry *CalendarInterval, element *treeElement, entries *[]CalendarInterval) {
	setCalendarIntervalValueFromType(currentEntry, element.value, element.elementType)
	if len(element.subElements) == 0 {
		// end of the line, this entry is finished
		*entries = append(*entries, *currentEntry)
		return
	}
	for _, subElement := range element.subElements {
		// new branch means new calendar entry
		fillInValueFromScheduleTreeElement(currentEntry.clone(), subElement, entries)
	}
}

func setCalendarIntervalValueFromType(entry *CalendarInterval, value int, typeValue calendar.TypeValue) {
	if entry == nil {
		// should never happen (keep in case we change the workflow later on)
		entry = newCalendarInterval()
	}
	switch typeValue {
	case calendar.TypeWeekDay:
		(*entry)[intervalWeekday] = value
	case calendar.TypeMonth:
		(*entry)[intervalMonth] = value
	case calendar.TypeDay:
		(*entry)[intervalDay] = value
	case calendar.TypeHour:
		(*entry)[intervalHour] = value
	case calendar.TypeMinute:
		(*entry)[intervalMinute] = value
	}
}

// ParseCalendarIntervals converts calendar intervals into a single calendar event.
// TODO: find a pattern on how to split into multiple events when needed
func ParseCalendarIntervals(intervals []CalendarInterval) []string {
	event := calendar.NewEvent(func(e *calendar.Event) {
		_ = e.Second.AddValue(0)
	})
	for _, interval := range intervals {
		for key, value := range interval {
			switch key {
			case intervalMinute:
				_ = event.Minute.AddValue(value)
			case intervalHour:
				_ = event.Hour.AddValue(value)
			case intervalWeekday:
				_ = event.WeekDay.AddValue(value)
			case intervalDay:
				_ = event.Day.AddValue(value)
			case intervalMonth:
				_ = event.Month.AddValue(value)
			}
		}
	}
	return []string{event.String()}
}
