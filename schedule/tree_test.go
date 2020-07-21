package schedule

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

func BenchmarkManualSliceCopy(b *testing.B) {
	max := 10000
	testData := make([]int, max)
	for i := 0; i < max; i++ {
		testData[i] = i
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		newData := make([]int, len(testData))
		for i := 0; i < len(testData); i++ {
			newData[i] = testData[i]
		}
		assert.Len(b, newData, max)
	}
}

// This one performs much better, but notice how the slice should be pre-allocated
// Also doesn't work with allocation like newData := make([]int, 0, len(testData))
func BenchmarkBuiltinSliceCopy(b *testing.B) {
	max := 10000
	testData := make([]int, max)
	for i := 0; i < max; i++ {
		testData[i] = i
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		newData := make([]int, len(testData))
		copy(newData, testData)
		assert.Len(b, newData, max)
	}
}
