package ansi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCursorUpLeftN(t *testing.T) {
	assert.Equal(t, Escape+"[0F", CursorUpLeftN(-1))
	assert.Equal(t, Escape+"[0F", CursorUpLeftN(0))
	assert.Equal(t, Escape+"[1F", CursorUpLeftN(1))
	assert.Equal(t, Escape+"[2F", CursorUpLeftN(2))
}
