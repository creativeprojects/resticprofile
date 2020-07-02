package calendar

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyValue(t *testing.T) {
	min := 10
	max := 20
	value := NewValue(min, max)
	assert.False(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.False(t, value.HasRange())
	assert.False(t, value.HasContiguousRange())
	assert.ElementsMatch(t, []int{}, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{}, value.GetRanges())
	assert.Equal(t, "*", value.String())
}

func TestSingleValue(t *testing.T) {
	min := 0
	max := 10
	entry := 5
	value := NewValue(min, max)
	value.AddValue(entry)
	assert.True(t, value.HasValue())
	assert.True(t, value.HasSingleValue())
	assert.False(t, value.HasRange())
	assert.False(t, value.HasContiguousRange())
	assert.ElementsMatch(t, []int{entry}, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{entry, entry}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d", entry), value.String())
}

func TestSimpleRangeValue(t *testing.T) {
	min := 1
	max := 9
	entries := []int{min, max}
	value := NewValue(min, max)
	for _, entry := range entries {
		value.AddValue(entry)
	}
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.False(t, value.HasContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{min, min}, {max, max}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d,%02d", min, max), value.String())
}

func TestContiguousRangeValue(t *testing.T) {
	min := 10
	max := 20
	entries := []int{11, 12, 14}
	value := NewValue(min, max)
	for _, entry := range entries {
		value.AddValue(entry)
	}
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.True(t, value.HasContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{11, 12}, {14, 14}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d..%02d,%02d", entries[0], entries[1], entries[2]), value.String())
}

func TestComplexContiguousRanges(t *testing.T) {
	min := 10
	max := 20
	entries := []int{10, 11, 14, 15, 16, 19, 20}
	value := NewValue(min, max)
	for _, entry := range entries {
		value.AddValue(entry)
	}
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.True(t, value.HasContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{10, 11}, {14, 16}, {19, 20}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d..%02d,%02d..%02d,%02d..%02d", entries[0], entries[1], entries[2], entries[4], entries[5], entries[6]), value.String())
}

func TestAddRanges(t *testing.T) {
	min := 10
	max := 20
	entries := []int{11, 12, 15}
	value := NewValue(min, max)
	value.AddRange(11, 12)
	value.AddRange(15, 15)
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.True(t, value.HasContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{11, 12}, {15, 15}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d..%02d,%02d", entries[0], entries[1], entries[2]), value.String())
}

func TestAddValueOutOfRange(t *testing.T) {
	min := 10
	max := 20
	entries := []int{min - 1, max + 1}
	for _, entry := range entries {
		value := NewValue(min, max)
		assert.Panics(t, func() {
			value.AddValue(entry)
		})
	}
}

func TestAddRangeValuesOutOfRange(t *testing.T) {
	min := 10
	max := 20

	value := NewValue(min, max)
	assert.Panics(t, func() {
		value.AddRange(min-1, max-1)
	})

}
