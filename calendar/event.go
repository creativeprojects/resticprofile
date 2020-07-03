package calendar

import (
	"fmt"
	"strings"
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

func NewEvent(initValues ...func(*Event)) *Event {
	event := &Event{
		WeekDay: NewValue(1, 7),
		Year:    NewValue(2000, 2200),
		Month:   NewValue(1, 12),
		Day:     NewValue(1, 31),
		Hour:    NewValue(0, 23),
		Minute:  NewValue(0, 59),
		Second:  NewValue(0, 59),
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
			break
		}
	}

	return nil
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
