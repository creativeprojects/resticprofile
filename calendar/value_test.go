package calendar

import (
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
	assert.ElementsMatch(t, []struct{ start, end int }{}, value.GetRanges())
}

func TestSingleValue(t *testing.T) {
	min := 10
	max := 20
	entry := 15
	value := NewValue(min, max)
	value.AddValue(entry)
	assert.True(t, value.HasValue())
	assert.True(t, value.HasSingleValue())
	assert.False(t, value.HasRange())
	assert.False(t, value.HasContiguousRange())
	assert.ElementsMatch(t, []int{entry}, value.GetRangeValues())
	assert.ElementsMatch(t, []struct{ start, end int }{{entry, entry}}, value.GetRanges())
}

func TestSimpleRangeValue(t *testing.T) {
	min := 10
	max := 20
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
	assert.ElementsMatch(t, []struct{ start, end int }{{min, min}, {max, max}}, value.GetRanges())
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
	assert.ElementsMatch(t, []struct{ start, end int }{{11, 12}, {14, 14}}, value.GetRanges())
}
