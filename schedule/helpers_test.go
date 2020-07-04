package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpread(t *testing.T) {
	size := 1 * 2 * 3 * 2
	testData := []struct {
		values   []int
		position int
		expected []int
	}{
		{[]int{1}, 0, []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		// {[]int{1, 2}, 1, []int{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2}},
		// {[]int{1, 2, 3}, 2, []int{1, 1, 2, 2, 3, 3, 1, 1, 2, 2, 3, 3}},
		// {[]int{1, 2}, 3, []int{1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2}},
	}

	for _, testItem := range testData {
		result := spread(testItem.values, size, testItem.position)
		assert.Equal(t, testItem.expected, result)
	}
}
