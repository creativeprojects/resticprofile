package term

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadingFromNilReader(t *testing.T) {
	r := new(nilReader)
	buffer, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, 0, len(buffer))
}
