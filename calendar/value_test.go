package calendar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyValue(t *testing.T) {
	min := uint(10)
	max := uint(20)
	value := NewValue(min, max)
	assert.False(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.False(t, value.HasRange())
}

func TestSingleValue(t *testing.T) {
	min := uint(10)
	max := uint(20)
	value := NewValue(min, max)
	value.AddValue(15)
	assert.True(t, value.HasValue())
	assert.True(t, value.HasSingleValue())
	assert.False(t, value.HasRange())
}

func TestSimpleRangeValue(t *testing.T) {
	min := uint(10)
	max := uint(20)
	value := NewValue(min, max)
	value.AddValue(min)
	value.AddValue(max)
	assert.True(t, value.HasValue())
	assert.False(t, value.HasSingleValue())
	assert.True(t, value.HasRange())
}
