package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilHelpers(t *testing.T) {
	num := new(1.1)
	str := new("1")
	b := new(true)

	t.Run("NilOr", func(t *testing.T) {
		assert.True(t, NilOr(num, 1.1))
		assert.True(t, NilOr(str, "1"))
		assert.True(t, NilOr(b, true))
		assert.False(t, NilOr(num, 2.2))
		assert.False(t, NilOr(str, "2"))
		assert.False(t, NilOr(b, false))
		assert.True(t, NilOr(nil, 1.1))
		assert.True(t, NilOr(nil, "1"))
		assert.True(t, NilOr(nil, true))
	})

	t.Run("NotNilAnd", func(t *testing.T) {
		assert.True(t, NotNilAnd(num, 1.1))
		assert.True(t, NotNilAnd(str, "1"))
		assert.True(t, NotNilAnd(b, true))
		assert.False(t, NotNilAnd(num, 2.2))
		assert.False(t, NotNilAnd(str, "2"))
		assert.False(t, NotNilAnd(b, false))
		assert.False(t, NotNilAnd(nil, 1.1))
		assert.False(t, NotNilAnd(nil, "1"))
		assert.False(t, NotNilAnd(nil, true))
	})
}
