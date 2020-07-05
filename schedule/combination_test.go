package schedule

import "testing"

func TestSimpleCombination(t *testing.T) {
	testData := []combinationItem{
		{value: 1},
		{value: 2},
		{value: 3},
		{value: 4},
		{value: 5},
	}

	size := 3
	combinations := generateCombination(testData, size)
	t.Logf("%+v", combinations)
}
