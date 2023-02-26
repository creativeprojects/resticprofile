package collect

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func even(i int) bool        { return i%2 == 0 }
func toInt(s string) (i int) { i, _ = strconv.Atoi(s); return }

func TestAll(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	expected := []int{2, 4}
	expectedInverse := []int{1, 3, 5}

	actual := All(input, even)
	assert.Equal(t, expected, actual)

	actual = All(input, Not(even))
	assert.Equal(t, expectedInverse, actual)

	assert.Nil(t, All([]int(nil), even))
	assert.Nil(t, All([]int{}, even))
	assert.Nil(t, All([]int{3}, even))
}

func TestIn(t *testing.T) {
	isIn := In("a", "b", "c")
	assert.False(t, isIn(""))
	assert.False(t, isIn("d"))
	assert.False(t, isIn("A"))
	assert.True(t, isIn("a"))
	assert.True(t, isIn("b"))
	assert.True(t, isIn("c"))
}

func TestFrom(t *testing.T) {
	input := []string{"1", "2", "3"}
	expected := []int{1, 2, 3}

	actual := From(input, toInt)
	assert.Equal(t, expected, actual)

	assert.Nil(t, From([]string(nil), toInt))
	assert.Nil(t, From([]string{}, toInt))
}
