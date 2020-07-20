package schedule

import "github.com/creativeprojects/resticprofile/calendar"

type treeElement struct {
	value       int
	elementType calendar.TypeValue
	subElements []treeElement
}

var (
	valuesOrder = []calendar.TypeValue{
		calendar.TypeMonth,
		calendar.TypeDay,
		calendar.TypeWeekDay,
		calendar.TypeHour,
		calendar.TypeMinute,
	}
)

func generateTreeOfSchedules(event *calendar.Event) []treeElement {
	elements := make([]treeElement, 0)
	for _, currentTypeValue := range valuesOrder {
		value := getCalendarValue(event, currentTypeValue)
		values := value.GetRangeValues()
		if len(values) > 0 {
			// do something
		}
	}
	return elements
}

func getCalendarValue(event *calendar.Event, typeValue calendar.TypeValue) *calendar.Value {
	switch typeValue {
	case calendar.TypeMonth:
		return event.Month
	case calendar.TypeDay:
		return event.Day
	case calendar.TypeWeekDay:
		return event.WeekDay
	case calendar.TypeHour:
		return event.Hour
	case calendar.TypeMinute:
		return event.Minute
	}
	return nil
}
