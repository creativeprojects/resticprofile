package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInlineTemplates(t *testing.T) {
	base := `---
version: 2
templates:
  t1:
    status-file: status-one
    backup:
      source: ["source-one"]
  t2:
    status-file: status-two
    backup:
      source: ["source-two"]
  t3:
    source: ["${named-source}", "${another-source}"]
`
	load := func(t *testing.T, content string) *Config {
		buffer := bytes.NewBufferString(base + content)
		cfg, err := Load(buffer, FormatYAML)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		return cfg
	}

	t.Run("merging-order", func(t *testing.T) {
		config := load(t, `
profiles:
  profile:
    templates:
      - t1
      - t2
    backup:
      source: "source-profile"
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, "status-two", p.StatusFile)
		assert.Equal(t, []string{"source-two"}, p.Backup.Source)
	})

	t.Run("vars", func(t *testing.T) {
		config := load(t, `
profiles:
  profile:
    backup:
      templates:
        - name: t3
          vars:
            named-source: my-source
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, []string{"my-source", "${another-source}"}, p.Backup.Source)
	})

	// TODO: More tests
}
