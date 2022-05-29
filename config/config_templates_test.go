package config

import (
	"bytes"
	"fmt"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
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
      run-before: ["mountpoint -q /backup"]
      run-after: ["touch /backup/lastrun"]
  t2:
    status-file: status-two
    backup:
      source: ["source-two"]
      run-before+0: ["mount /backup"]
      run-after++: ["unmount /backup"]
  t3:
    default-vars:
      named-source: "default-source"
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
		assert.Equal(t, []string{"mount /backup", "mountpoint -q /backup"}, p.Backup.RunBefore)
		assert.Equal(t, []string{"touch /backup/lastrun", "unmount /backup"}, p.Backup.RunAfter)
	})

	t.Run("vars", func(t *testing.T) {
		config := load(t, `
profiles:
  profile-simple:
    backup:
      templates: t3
  profile:
    backup:
      templates:
        - name: t3
          vars:
            named-source: my-source
`)
		p, err := config.getProfile("profile-simple")
		assert.NoError(t, err)
		assert.Equal(t, []string{"default-source", "${another-source}"}, p.Backup.Source)

		p, err = config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, []string{"my-source", "${another-source}"}, p.Backup.Source)
	})

	t.Run("inheritance-is-not-affected", func(t *testing.T) {
		config := load(t, `
profiles:
  default:
    templates: 
      - t2
  profile:
    inherit: default
    backup:
      run-before: ["echo profile-before"]
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo profile-before"}, p.Backup.RunBefore)
		assert.Equal(t, []string{"unmount /backup"}, p.Backup.RunAfter)
	})

	// TODO: More tests
}

func TestRevolveAppendToListKeys(t *testing.T) {
	load := func(t *testing.T, config map[string]interface{}) *viper.Viper {
		v := viper.New()
		require.NoError(t, v.MergeConfigMap(config))
		return v
	}

	mm := func(kv ...interface{}) (mapped map[string]interface{}) {
		mapped = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			mapped[cast.ToString(kv[i])] = kv[i+1]
		}
		return
	}

	ml := func(v ...interface{}) []interface{} {
		return v
	}

	tests := []struct {
		name                       string
		config, template, expected map[string]interface{}
	}{
		{
			name:     "append-string-to-string",
			config:   mm("key", "base"),
			template: mm("key++", "new"),
			expected: mm("key", ml("base", "new")),
		},
		{
			name:     "append-string-to-list",
			config:   mm("key", ml("base")),
			template: mm("key++", "new"),
			expected: mm("key", ml("base", "new")),
		},
		{
			name:     "append-list-to-list",
			config:   mm("key", ml("base")),
			template: mm("key++", ml("new")),
			expected: mm("key", ml("base", "new")),
		},
		{
			name:     "append-one-to-many",
			config:   mm("key", ml("base-1", "base-2")),
			template: mm("key++", "new"),
			expected: mm("key", ml("base-1", "base-2", "new")),
		},
		{
			name:     "append-many-to-one",
			config:   mm("key", "base"),
			template: mm("key++", ml("new-1", "new-2")),
			expected: mm("key", ml("base", "new-1", "new-2")),
		},
		{
			name:     "prepend-one-to-many",
			config:   mm("key", ml("base-1", "base-2")),
			template: mm("key+0", "new"),
			expected: mm("key", ml("new", "base-1", "base-2")),
		},
		{
			name:     "prepend-many-to-one",
			config:   mm("key", "base"),
			template: mm("key+0", ml("new-1", "new-2")),
			expected: mm("key", ml("new-1", "new-2", "base")),
		},
		{
			name:     "prepend-list-to-list",
			config:   mm("key", ml("base-1", "base-2")),
			template: mm("key+0", ml("new-1", "new-2")),
			expected: mm("key", ml("new-1", "new-2", "base-1", "base-2")),
		},
		{
			name:     "append-object-to-string",
			config:   mm("key", "base"),
			template: mm("key++", mm("newKey", "newValue")),
			expected: mm("key", ml("base", mm("newKey", "newValue"))),
		},
		{
			name:     "append-nested",
			config:   mm("key", mm("childKey", "childBase")),
			template: mm("key", mm("childKey++", "childNew")),
			expected: mm("key", mm("childKey", ml("childBase", "childNew"))),
		},
		{
			name:     "append-to-none",
			config:   mm(),
			template: mm("key", mm("childKey++", "childNew")),
			expected: mm("key", mm("childKey", ml("childNew"))),
		},
		{
			name:     "prepend-to-none",
			config:   mm(),
			template: mm("key", mm("childKey+0", "childNew")),
			expected: mm("key", mm("childKey", ml("childNew"))),
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("#%d_%s", i, test.name), func(t *testing.T) {
			v := load(t, test.config)
			revolveAppendToListKeys(v, test.template)
			assert.Equal(t, test.expected, test.template)
		})
	}
}
