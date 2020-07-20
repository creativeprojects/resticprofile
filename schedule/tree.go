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
	currentElements := elements
	for _, currentTypeValue := range valuesOrder {
		value := getCalendarValue(event, currentTypeValue)
		values := value.GetRangeValues()
		if len(values) > 0 {
			subTree := getElements(values, currentTypeValue)
			if len(currentElements) > 0 {
				for _, element := range currentElements {
					element.subElements = make([]treeElement, len(subTree))
					copy(element.subElements, subTree)
				}
			} else {
				currentElements = make([]treeElement, len(subTree))
				copy(currentElements, subTree)
			}
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

func getElements(values []int, typeValue calendar.TypeValue) []treeElement {
	elements := make([]treeElement, len(values))
	for i, value := range values {
		elements[i] = treeElement{
			value:       value,
			elementType: typeValue,
		}
	}
	return elements
}
