package schedule

import "github.com/creativeprojects/resticprofile/calendar"

type treeElement struct {
	value       int
	elementType calendar.TypeValue
	subElements []*treeElement
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

func generateTreeOfSchedules(event *calendar.Event) []*treeElement {
	elements := make([]*treeElement, 0)
	currentElements := &elements
	for _, currentTypeValue := range valuesOrder {
		field := event.Field(currentTypeValue)
		values := field.GetRangeValues()
		if len(values) == 0 {
			continue
		}
		subTree := getElements(values, currentTypeValue)
		if len(*currentElements) > 0 {
			for _, element := range *currentElements {
				element.subElements = make([]*treeElement, len(subTree))
				copy(element.subElements, subTree)
			}
		} else {
			*currentElements = make([]*treeElement, len(subTree))
			copy(*currentElements, subTree)
		}
	}
	return elements
}

func getElements(values []int, typeValue calendar.TypeValue) []*treeElement {
	elements := make([]*treeElement, len(values))
	for i, value := range values {
		elements[i] = &treeElement{
			value:       value,
			elementType: typeValue,
		}
	}
	return elements
}
