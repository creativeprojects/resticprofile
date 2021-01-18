package calendar

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Should be able to read the same calendar events
// https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events

// Event represents a calendar event.
// It can be one specific point in time, or a recurring event
type Event struct {
	WeekDay *Value
	Year    *Value
	Month   *Value
	Day     *Value
	Hour    *Value
	Minute  *Value
	Second  *Value
}

// NewEvent instantiates a new event with all its default values
func NewEvent(initValues ...func(*Event)) *Event {
	event := &Event{
		WeekDay: NewValueFromType(TypeWeekDay),
		Year:    NewValueFromType(TypeYear),
		Month:   NewValueFromType(TypeMonth),
		Day:     NewValueFromType(TypeDay),
		Hour:    NewValueFromType(TypeHour),
		Minute:  NewValueFromType(TypeMinute),
		Second:  NewValueFromType(TypeSecond),
	}

	for _, initValue := range initValues {
		initValue(event)
	}
	return event
}

// String representation
func (e *Event) String() string {
	output := ""
	if e.WeekDay.HasValue() {
		output += numbersToWeekdays(e.WeekDay.String()) + " "
	}
	output += e.Year.String() + "-" +
		e.Month.String() + "-" +
		e.Day.String() + " " +
		e.Hour.String() + ":" +
		e.Minute.String() + ":" +
		e.Second.String()

	return output
}

// Parse a string into an event
func (e *Event) Parse(input string) error {
	if input == "" {
		return errors.New("calendar event cannot be an empty string")
	}

	// check for a keyword
	for keyword, setValues := range specialKeywords {
		if input == keyword {
			setValues(e)
			return nil
		}
	}

	// check for all variations one by one
	for _, rule := range parsingRules {
		if match := rule.expr.FindStringSubmatch(input); match != nil {
			for _, parseValue := range rule.parseValues {
				err := parseValue(e, match)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	return errors.New("calendar event doesn't match any well known pattern")
}

// Next returns the next schedule for this event
func (e *Event) Next(from time.Time) time.Time {
	// start from time and increment of 1 minute each time
	next := from.Truncate(time.Minute) // truncate all the seconds
	// should stop in 2 years time to avoid an infinite loop
	endYear := from.Year() + 2
	for next.Year() <= endYear {
		if e.match(next) {
			return next
		}
		// increment 1 minute
		next = next.Add(time.Minute)
	}
	return time.Time{}
}

// AsTime returns a time.Time representation of the event if possible,
// if not possible the second value will be false
func (e *Event) AsTime() (time.Time, bool) {
	// Let's not bother with seconds
	if !e.WeekDay.HasValue() &&
		e.Year.HasSingleValue() &&
		e.Month.HasSingleValue() &&
		e.Day.HasSingleValue() &&
		e.Hour.HasSingleValue() &&
		e.Minute.HasSingleValue() {
		event, err := time.Parse("2006-01-02 15:04:05", e.String())
		return event, err == nil
	}
	return time.Now(), false
}

// GetAllInBetween returns all activation times of the event in between these two dates.
// the minimum increment is 1 minute
func (e *Event) GetAllInBetween(start, end time.Time) []time.Time {
	// align time to the minute (zeroing the seconds)
	start = start.Truncate(time.Minute)
	end = end.Truncate(time.Minute)
	recurrences := []time.Time{}
	next := e.Next(start)
	for next.Before(end) {
		recurrences = append(recurrences, next)
		next = e.Next(next.Add(time.Minute))
	}
	return recurrences
}

// Field returns a calendar.Value from the type (year, month, day, etc.)
func (e *Event) Field(typeValue TypeValue) *Value {
	switch typeValue {
	case TypeYear:
		return e.Year
	case TypeMonth:
		return e.Month
	case TypeDay:
		return e.Day
	case TypeWeekDay:
		return e.WeekDay
	case TypeHour:
		return e.Hour
	case TypeMinute:
		return e.Minute
	case TypeSecond:
		return e.Second
	}
	return nil
}

// IsDaily means all events are within a day (from once to multiple times a day)
func (e *Event) IsDaily() bool {
	return !e.Year.HasValue() && !e.Month.HasValue() && !e.Day.HasValue() && !e.WeekDay.HasValue()
}

// IsWeekly means all events runs on specific days of the week (every week)
func (e *Event) IsWeekly() bool {
	return !e.Year.HasValue() && !e.Month.HasValue() && !e.Day.HasValue() && e.WeekDay.HasValue()
}

// IsMonthly means all events runs on specific days of the month, and/or specific months.
func (e *Event) IsMonthly() bool {
	return !e.Year.HasValue() && (e.Month.HasValue() || e.Day.HasValue())
}

// match returns true if the time in parameter would trigger the event
func (e *Event) match(currentTime time.Time) bool {
	values := []struct {
		ref     *Value
		current int
	}{
		{e.Year, currentTime.Year()},
		{e.Month, int(currentTime.Month())},
		{e.Day, currentTime.Day()},
		{e.WeekDay, int(currentTime.Weekday())},
		{e.Hour, currentTime.Hour()},
		{e.Minute, currentTime.Minute()},
		// Not really useful to check for the seconds
	}
	for _, value := range values {
		if !value.ref.HasValue() {
			continue
		}
		if value.ref.HasSingleValue() {
			if value.ref.singleValue != value.current {
				return false
			}
		}
		if !value.ref.IsInRange(value.current) {
			return false
		}
	}
	return true
}

func numbersToWeekdays(weekdays string) string {
	for day := minDay; day < maxDay; day++ {
		weekdays = strings.ReplaceAll(weekdays, fmt.Sprintf("%02d", day), capitalize(shortWeekDay[day]))
	}
	return weekdays
}

func capitalize(value string) string {
	if value == "" {
		return value
	}
	value = strings.ToUpper(value[0:1]) + value[1:]
	return value
}
