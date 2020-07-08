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
	// start from time and increment of 1 second each time
	next := from
	// should stop in 2 years time to avoid an infinite loop
	endYear := from.Year() + 2
	for next.Year() <= endYear {
		if e.match(next) {
			return next
		}
		// increment 1 second
		next = next.Add(time.Second)
	}
	return time.Time{}
}

// AsTime returns a time.Time representation of the event if possible,
// if not possible the second value will be false
func (e *Event) AsTime() (time.Time, bool) {
	// Let's not bother about seconds
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
	recurrences := []time.Time{}
	next := e.Next(start)
	for next.Before(end) {
		recurrences = append(recurrences, next)
		next = e.Next(next.Add(time.Minute))
	}
	return recurrences
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
		{e.Second, currentTime.Second()},
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
	for day := 1; day <= 7; day++ {
		weekdays = strings.ReplaceAll(weekdays, fmt.Sprintf("%02d", day), capitalize(shortWeekDay[day-1]))
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
