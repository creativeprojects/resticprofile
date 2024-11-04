package crond

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
)

func parseEvent(source string) (*calendar.Event, error) {
	event := calendar.NewEvent()
	source = strings.ReplaceAll(source, "\t", " ")
	source = strings.ReplaceAll(source, " ", " ")
	source = strings.TrimSpace(source)
	parts := strings.Split(source, " ")
	if len(parts) != 5 {
		return nil, fmt.Errorf("expected 5 fields but found %d: %q", len(parts), source)
	}

	err := event.Second.AddValue(0)
	if err != nil {
		return nil, err
	}

	for index, eventField := range []*calendar.Value{
		event.Minute,
		event.Hour,
		event.Day,
		event.Month,
		event.WeekDay,
	} {
		err := parseField(parts[index], eventField)
		if err != nil {
			return event, fmt.Errorf("error parsing %q: %w", parts[index], err)
		}
	}
	return event, nil
}

func parseField(field string, eventField *calendar.Value) error {
	if field == "*" {
		return nil
	}
	// list of values
	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		for _, part := range parts {
			err := parseField(part, eventField)
			if err != nil {
				return err
			}
		}
		return nil
	}
	// range of values
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) != 2 {
			return fmt.Errorf("expecting 2 values, found %d: %q", len(parts), field)
		}
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		err = eventField.AddRange(start, end)
		if err != nil {
			return err
		}
		return nil
	}
	// single value
	value, err := strconv.Atoi(field)
	if err != nil {
		return err
	}
	err = eventField.AddValue(value)
	if err != nil {
		return err
	}
	return nil
}
