package schedule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
)

func TestSimpleCombination(t *testing.T) {
	testData := []combinationItem{
		{value: 1},
		{value: 2},
		{value: 3},
		{value: 4},
		{value: 5},
	}
	expected := `00: [?01, ?02, ?03]
01: [?01, ?02, ?04]
02: [?01, ?02, ?05]
03: [?01, ?03, ?04]
04: [?01, ?03, ?05]
05: [?01, ?04, ?05]
06: [?02, ?03, ?04]
07: [?02, ?03, ?05]
08: [?02, ?04, ?05]
09: [?03, ?04, ?05]
`
	size := 3
	combinations := generateCombination(testData, size)
	assert.Equal(t, expected, combinationsToString(combinations))
}

func combinationsToString(combinations [][]combinationItem) string {
	if len(combinations) == 0 {
		return ""
	}
	builder := &strings.Builder{}
	values := make([]string, len(combinations[0]))
	for lineIndex, line := range combinations {
		for i, value := range line {
			shortType := "?"
			switch value.itemType {
			case calendar.TypeWeekDay:
				shortType = "W"
			case calendar.TypeMonth:
				shortType = "M"
			case calendar.TypeDay:
				shortType = "D"
			case calendar.TypeHour:
				shortType = "H"
			case calendar.TypeMinute:
				shortType = "m"
			}
			values[i] = fmt.Sprintf("%s%02d", shortType, value.value)
		}
		fmt.Fprintf(builder, "%02d: [%s]\n", lineIndex, strings.Join(values, ", "))
	}
	return builder.String()
}
