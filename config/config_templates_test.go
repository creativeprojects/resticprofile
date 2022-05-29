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

func inlineTemplateTestTools() (
	makeMap func(kv ...interface{}) (mapped map[string]interface{}),
	makeList func(v ...interface{}) []interface{},
) {
	makeMap = func(kv ...interface{}) (mapped map[string]interface{}) {
		mapped = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			mapped[cast.ToString(kv[i])] = kv[i+1]
		}
		return
	}

	makeList = func(v ...interface{}) []interface{} {
		return v
	}
	return
}

func TestInlineTemplate(t *testing.T) {
	mm, list := inlineTemplateTestTools()

	tpl := inlineTemplate{
		DefaultVariables: keysToUpper(mm("MyVar", "MyDefault")),
		Source: mm(
			"string", "string-val",
			"int", 321,
			"bool", true,
			"list", list(1, "2", 3),
			"list_with_vars", list("${var-1}", "$VAR_2", "[$MyVar]", "-${myvar}-", "-${MYVAR}-"),
			"string_with_vars", "${var-1} $Var_2 [$MyVar] -${myvar}- -${MYVAR}-",
			"nested_with_vars", list(
				mm(
					"list_with_vars", list("${var-1}"),
					"string_with_vars", "${var-1}",
					"nested", mm("key", "$MyVar"),
				),
				"$var_2",
			),
		),
	}

	t.Run("default-vars", func(t *testing.T) {
		assert.Equal(t, mm(
			"string", "string-val",
			"int", 321,
			"bool", true,
			"list", list(1, "2", 3),
			"list_with_vars", list("${var-1}", "${VAR_2}", "[MyDefault]", "-MyDefault-", "-MyDefault-"),
			"string_with_vars", "${var-1} ${Var_2} [MyDefault] -MyDefault- -MyDefault-",
			"nested_with_vars", list(
				mm(
					"list_with_vars", list("${var-1}"),
					"string_with_vars", "${var-1}",
					"nested", mm("key", "MyDefault"),
				),
				"${var_2}",
			),
		), tpl.Resolve(nil))
	})

	t.Run("non-default-vars", func(t *testing.T) {
		assert.Equal(t, mm(
			"string", "string-val",
			"int", 321,
			"bool", true,
			"list", list(1, "2", 3),
			"list_with_vars", list("${var-1}", "${VAR_2}", "[MySpecific]", "-MySpecific-", "-MySpecific-"),
			"string_with_vars", "${var-1} ${Var_2} [MySpecific] -MySpecific- -MySpecific-",
			"nested_with_vars", list(
				mm(
					"list_with_vars", list("${var-1}"),
					"string_with_vars", "${var-1}",
					"nested", mm("key", "MySpecific"),
				),
				"${var_2}",
			),
		), tpl.Resolve(keysToUpper(mm("myvar", "MySpecific"))))
	})

	t.Run("all-resolved", func(t *testing.T) {
		assert.Equal(t, mm(
			"string", "string-val",
			"int", 321,
			"bool", true,
			"list", list(1, "2", 3),
			"list_with_vars", list("val1", "val2", "[MySpecific]", "-MySpecific-", "-MySpecific-"),
			"string_with_vars", "val1 val2 [MySpecific] -MySpecific- -MySpecific-",
			"nested_with_vars", list(
				mm(
					"list_with_vars", list("val1"),
					"string_with_vars", "val1",
					"nested", mm("key", "MySpecific"),
				),
				"val2",
			),
		), tpl.Resolve(
			keysToUpper(mm(
				"myvar", "MySpecific",
				"var-1", "val1",
				"var_2", "val2",
			)),
		))
	})
}

func TestRevolveAppendToListKeys(t *testing.T) {
	load := func(t *testing.T, config map[string]interface{}) *viper.Viper {
		v := viper.New()
		require.NoError(t, v.MergeConfigMap(config))
		return v
	}

	mm, list := inlineTemplateTestTools()

	tests := []struct {
		name                       string
		config, template, expected map[string]interface{}
	}{
		{
			name:     "append-string-to-string",
			config:   mm("key", "base"),
			template: mm("key++", "new"),
			expected: mm("key", list("base", "new")),
		},
		{
			name:     "append-string-to-list",
			config:   mm("key", list("base")),
			template: mm("key++", "new"),
			expected: mm("key", list("base", "new")),
		},
		{
			name:     "append-list-to-list",
			config:   mm("key", list("base")),
			template: mm("key++", list("new")),
			expected: mm("key", list("base", "new")),
		},
		{
			name:     "append-one-to-many",
			config:   mm("key", list("base-1", "base-2")),
			template: mm("key++", "new"),
			expected: mm("key", list("base-1", "base-2", "new")),
		},
		{
			name:     "append-many-to-one",
			config:   mm("key", "base"),
			template: mm("key++", list("new-1", "new-2")),
			expected: mm("key", list("base", "new-1", "new-2")),
		},
		{
			name:     "prepend-one-to-many",
			config:   mm("key", list("base-1", "base-2")),
			template: mm("key+0", "new"),
			expected: mm("key", list("new", "base-1", "base-2")),
		},
		{
			name:     "prepend-many-to-one",
			config:   mm("key", "base"),
			template: mm("key+0", list("new-1", "new-2")),
			expected: mm("key", list("new-1", "new-2", "base")),
		},
		{
			name:     "prepend-list-to-list",
			config:   mm("key", list("base-1", "base-2")),
			template: mm("key+0", list("new-1", "new-2")),
			expected: mm("key", list("new-1", "new-2", "base-1", "base-2")),
		},
		{
			name:     "append-object-to-string",
			config:   mm("key", "base"),
			template: mm("key++", mm("newKey", "newValue")),
			expected: mm("key", list("base", mm("newKey", "newValue"))),
		},
		{
			name:     "append-nested",
			config:   mm("key", mm("childKey", "childBase")),
			template: mm("key", mm("childKey++", "childNew")),
			expected: mm("key", mm("childKey", list("childBase", "childNew"))),
		},
		{
			name:     "append-to-none",
			config:   mm(),
			template: mm("key", mm("childKey++", "childNew")),
			expected: mm("key", mm("childKey", list("childNew"))),
		},
		{
			name:     "prepend-to-none",
			config:   mm(),
			template: mm("key", mm("childKey+0", "childNew")),
			expected: mm("key", mm("childKey", list("childNew"))),
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

func TestApplyInlineTemplates(t *testing.T) {
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

	t.Run("merging-order-with-append", func(t *testing.T) {
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

	t.Run("list-append-does-not-inherit", func(t *testing.T) {
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

	t.Run("invalid-template-does-not-fail", func(t *testing.T) {
		config := load(t, `
  tvalid:
    default-vars:
      string: abc
      number: 1
      bool: false
      list: ["a", "b"]
      obj: {a: b}
    status-file: "$string $number $bool $list $obj"

  tinvalid-no-object1: "-"

  tinvalid-no-object2:
    - key: value

profiles:
  profile:
    templates: tvalid
`)
		templates := parseInlineTemplates(config.viper)
		assert.Contains(t, templates, "tvalid")
		assert.NotContains(t, templates, "tinvalid-no-object1")
		assert.NotContains(t, templates, "tinvalid-no-object2")

		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, "abc 1 false ${list} ${obj}", p.StatusFile)
	})

	t.Run("unknown-call-fails", func(t *testing.T) {
		buffer := bytes.NewBufferString(base + `
profiles:
  profile:
    templates: 
      - t1
      - t2
      - tunknown
`)
		_, err := Load(buffer, FormatYAML)
		assert.EqualError(t, err, "failed applying templates profiles.profile.templates: undefined template \"tunknown\"")
	})

	t.Run("invalid-call-fails", func(t *testing.T) {
		defaultError := "template call must be string or list of strings or list of call objects"
		invalidCalls := map[string]string{
			"templates: false":   defaultError,
			"templates: [1]":     "cannot parse template call [1]: '' expected a map, got 'int'",
			"templates: [false]": "cannot parse template call []: '' expected a map, got 'bool'",
			"templates: 1":       defaultError,
		}
		for call, errorMessage := range invalidCalls {
			buffer := bytes.NewBufferString(fmt.Sprintf(`
%s
profiles:
  profile:
    %s 
`, base, call))
			_, err := Load(buffer, FormatYAML)
			assert.EqualError(t, err, "failed applying templates profiles.profile.templates: "+errorMessage)
		}
	})
}
