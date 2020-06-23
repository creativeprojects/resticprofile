package filesearch

import (
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
)

func TestSearchConfigFile(t *testing.T) {
	found, err := xdg.SearchConfigFile(filepath.Join("some_service", "some_path", "some_file"))
	t.Log(err)
	assert.Empty(t, found)
	assert.Error(t, err)
}
