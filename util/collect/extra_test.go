package collect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstAndLast(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}

	actual := First(input, even)
	require.NotNil(t, actual)
	assert.Equal(t, 2, *actual)

	actual = Last(input, even)
	require.NotNil(t, actual)
	assert.Equal(t, 4, *actual)

	assert.Nil(t, First([]int(nil), even))
	assert.Nil(t, First([]int{}, even))
	assert.Nil(t, First([]int{3}, even))
	assert.Nil(t, Last([]int(nil), even))
	assert.Nil(t, Last([]int{}, even))
	assert.Nil(t, Last([]int{3}, even))
}

func TestFromMap(t *testing.T) {
	input := map[string]int{"1": 1, "2": 2, "3": 3, "4": 4, "5": 5}
	expected := map[int]string{2: "2", 4: "4"}

	evenSwapped := func(k string, v int) (rk int, rv string, include bool) {
		rk = v
		rv = k
		include = even(v)
		return
	}

	actual := FromMap(input, evenSwapped)
	assert.Equal(t, expected, actual)

	assert.Nil(t, FromMap(map[string]int(nil), evenSwapped))
	assert.Nil(t, FromMap(map[string]int{}, evenSwapped))
}

func TestKVMapper(t *testing.T) {
	input := map[string]int{"1": 1, "2": 2, "3": 3, "4": 4, "5": 5}
	expected := map[string]int{"d1": 2, "d2": 4, "d3": 6, "d4": 8, "d5": 10}

	output := FromMap(input, KVMapper(CopyMapper[string], CopyMapper[int]))
	assert.Equal(t, input, output)

	output = FromMap(input, KVMapper(
		func(k string) string { return "d" + k },
		func(v int) int { return 2 * v }))
	assert.Equal(t, expected, output)
}
