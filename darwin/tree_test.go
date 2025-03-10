//go:build darwin

package darwin

import (
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTree(t *testing.T) {
	testData := []struct {
		input    string
		expected []*treeElement
	}{
		{"*-*-*", []*treeElement{
			{
				value:       0,
				elementType: calendar.TypeHour,
				subElements: []*treeElement{
					{
						value:       0,
						elementType: calendar.TypeMinute,
					},
				},
			},
		}},
		{"*:0,30", []*treeElement{
			{
				value:       0,
				elementType: calendar.TypeMinute,
				subElements: nil,
			},
			{
				value:       30,
				elementType: calendar.TypeMinute,
				subElements: nil,
			},
		}},
		{"0,12:20", []*treeElement{
			{
				value:       0,
				elementType: calendar.TypeHour,
				subElements: []*treeElement{
					{
						value:       20,
						elementType: calendar.TypeMinute,
					},
				},
			},
			{
				value:       12,
				elementType: calendar.TypeHour,
				subElements: []*treeElement{
					{
						value:       20,
						elementType: calendar.TypeMinute,
					},
				},
			},
		}},
		{"0,12:20,40", []*treeElement{
			{
				value:       0,
				elementType: calendar.TypeHour,
				subElements: []*treeElement{
					{
						value:       20,
						elementType: calendar.TypeMinute,
					},
					{
						value:       40,
						elementType: calendar.TypeMinute,
					},
				},
			},
			{
				value:       12,
				elementType: calendar.TypeHour,
				subElements: []*treeElement{
					{
						value:       20,
						elementType: calendar.TypeMinute,
					},
					{
						value:       40,
						elementType: calendar.TypeMinute,
					},
				},
			},
		}},
		{"Sun *-*-01..06 3:30", []*treeElement{
			{
				elementType: calendar.TypeDay,
				value:       1,
				subElements: []*treeElement{
					{
						elementType: calendar.TypeWeekDay,
						value:       0,
						subElements: []*treeElement{
							{
								elementType: calendar.TypeHour,
								value:       3,
								subElements: []*treeElement{
									{
										elementType: calendar.TypeMinute,
										value:       30,
									},
								},
							},
						},
					},
				},
			},
			{
				elementType: calendar.TypeDay,
				value:       2,
				subElements: []*treeElement{
					{
						elementType: calendar.TypeWeekDay,
						value:       0,
						subElements: []*treeElement{
							{
								elementType: calendar.TypeHour,
								value:       3,
								subElements: []*treeElement{
									{
										elementType: calendar.TypeMinute,
										value:       30,
									},
								},
							},
						},
					},
				},
			},
			{
				elementType: calendar.TypeDay,
				value:       3,
				subElements: []*treeElement{
					{
						elementType: calendar.TypeWeekDay,
						value:       0,
						subElements: []*treeElement{
							{
								elementType: calendar.TypeHour,
								value:       3,
								subElements: []*treeElement{
									{
										elementType: calendar.TypeMinute,
										value:       30,
									},
								},
							},
						},
					},
				},
			},
			{
				elementType: calendar.TypeDay,
				value:       4,
				subElements: []*treeElement{
					{
						elementType: calendar.TypeWeekDay,
						value:       0,
						subElements: []*treeElement{
							{
								elementType: calendar.TypeHour,
								value:       3,
								subElements: []*treeElement{
									{
										elementType: calendar.TypeMinute,
										value:       30,
									},
								},
							},
						},
					},
				},
			},
			{
				elementType: calendar.TypeDay,
				value:       5,
				subElements: []*treeElement{
					{
						elementType: calendar.TypeWeekDay,
						value:       0,
						subElements: []*treeElement{
							{
								elementType: calendar.TypeHour,
								value:       3,
								subElements: []*treeElement{
									{
										elementType: calendar.TypeMinute,
										value:       30,
									},
								},
							},
						},
					},
				},
			},
			{
				elementType: calendar.TypeDay,
				value:       6,
				subElements: []*treeElement{
					{
						elementType: calendar.TypeWeekDay,
						value:       0,
						subElements: []*treeElement{
							{
								elementType: calendar.TypeHour,
								value:       3,
								subElements: []*treeElement{
									{
										elementType: calendar.TypeMinute,
										value:       30,
									},
								},
							},
						},
					},
				},
			},
		}},
	}

	for _, testItem := range testData {
		t.Run(testItem.input, func(t *testing.T) {
			event := calendar.NewEvent()
			err := event.Parse(testItem.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, testItem.expected, generateTreeOfSchedules(event))
		})
	}
}
