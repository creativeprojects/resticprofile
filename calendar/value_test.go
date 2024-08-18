package calendar

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyValue(t *testing.T) {
	minValue := 10
	maxValue := 20
	value := NewValue(minValue, maxValue)
	assert.False(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.False(t, value.HasRange())
	assert.False(t, value.HasContiguousRange())
	assert.False(t, value.HasLongContiguousRange())
	assert.ElementsMatch(t, []int{}, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{}, value.GetRanges())
	assert.Equal(t, "*", value.String())
}

func TestSingleValue(t *testing.T) {
	minValue := 0
	maxValue := 10
	entry := 5
	value := NewValue(minValue, maxValue)
	err := value.AddValue(entry)
	require.NoError(t, err)
	assert.True(t, value.HasValue())
	assert.True(t, value.HasSingleValue())
	assert.False(t, value.HasRange())
	assert.False(t, value.HasContiguousRange())
	assert.False(t, value.HasLongContiguousRange())
	assert.ElementsMatch(t, []int{entry}, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{entry, entry}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d", entry), value.String())
}

func TestSimpleRangeValue(t *testing.T) {
	minValue := 1
	maxValue := 9
	entries := []int{minValue, maxValue}
	value := NewValue(minValue, maxValue)
	for _, entry := range entries {
		err := value.AddValue(entry)
		require.NoError(t, err)
	}
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.False(t, value.HasContiguousRange())
	assert.False(t, value.HasLongContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{minValue, minValue}, {maxValue, maxValue}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d,%02d", minValue, maxValue), value.String())
}

func TestContiguousRangeValue(t *testing.T) {
	minValue := 10
	maxValue := 20
	entries := []int{11, 12, 14}
	value := NewValue(minValue, maxValue)
	for _, entry := range entries {
		err := value.AddValue(entry)
		require.NoError(t, err)
	}
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.True(t, value.HasContiguousRange())
	assert.False(t, value.HasLongContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{11, 12}, {14, 14}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d,%02d,%02d", entries[0], entries[1], entries[2]), value.String())
}

func TestComplexContiguousRanges(t *testing.T) {
	minValue := 10
	maxValue := 20
	entries := []int{10, 11, 14, 15, 16, 19, 20}
	value := NewValue(minValue, maxValue)
	for _, entry := range entries {
		err := value.AddValue(entry)
		require.NoError(t, err)
	}
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.True(t, value.HasContiguousRange())
	assert.True(t, value.HasLongContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{10, 11}, {14, 16}, {19, 20}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d..%02d,%02d..%02d,%02d..%02d", entries[0], entries[1], entries[2], entries[4], entries[5], entries[6]), value.String())
}

func TestAddRanges(t *testing.T) {
	minValue := 10
	maxValue := 20
	entries := []int{11, 12, 15}
	value := NewValue(minValue, maxValue)
	err := value.AddRange(12, 11) // wrong order on purpose
	require.NoError(t, err)
	err = value.AddRange(15, 15)
	require.NoError(t, err)

	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
	assert.True(t, value.HasContiguousRange())
	assert.False(t, value.HasLongContiguousRange())
	assert.ElementsMatch(t, entries, value.GetRangeValues())
	assert.ElementsMatch(t, []Range{{11, 12}, {15, 15}}, value.GetRanges())
	assert.Equal(t, fmt.Sprintf("%02d,%02d,%02d", entries[0], entries[1], entries[2]), value.String())
}

func TestAddValueOutOfRange(t *testing.T) {
	minValue := 10
	maxValue := 20
	entries := []int{minValue - 1, maxValue + 1}
	for _, entry := range entries {
		value := NewValue(minValue, maxValue)
		assert.Panics(t, func() {
			value.MustAddValue(entry)
		})
	}
}

func TestAddRangeValuesOutOfRange(t *testing.T) {
	minValue := 10
	maxValue := 20

	value := NewValue(minValue, maxValue)
	assert.Panics(t, func() {
		value.MustAddRange(minValue-1, maxValue-1)
	})

}

func TestParseString(t *testing.T) {
	testData := []struct {
		min, max        int
		input, expected string
	}{
		{0, 0, "*", "*"},
		{0, 0, "", "*"},
		{1, 12, "1", "01"},
		{1, 12, "12", "12"},
		{1, 12, "1,12", "01,12"},
		{1, 12, "1,2,3", "01..03"},
		{1, 12, "1..3", "01..03"},
		{1, 12, "1..3,5..6,10..12", "01..03,05..06,10..12"},
		{1, 12, "1..3,5..5,10..12", "01..03,05,10..12"},
	}

	for _, testItem := range testData {
		t.Run(testItem.input, func(t *testing.T) {
			value := NewValue(testItem.min, testItem.max)
			err := value.Parse(testItem.input)
			assert.NoError(t, err)
			assert.Equal(t, testItem.expected, value.String())
		})
	}
}

func TestParseStringError(t *testing.T) {
	testData := []struct {
		min, max int
		input    string
	}{
		{1, 12, ".."},
		{1, 12, "1.."},
		{1, 12, "..1"},
		{1, 12, "0..10"},
		{1, 12, "1..13"},
	}

	for _, testItem := range testData {
		t.Run(testItem.input, func(t *testing.T) {
			value := NewValue(testItem.min, testItem.max)
			err := value.Parse(testItem.input)
			assert.Error(t, err)
		})
	}
}
