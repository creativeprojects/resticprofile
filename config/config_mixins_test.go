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

func mixinsTestTools() (
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

func TestMixin(t *testing.T) {
	mm, list := mixinsTestTools()

	tpl := mixin{
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

	mm, list := mixinsTestTools()

	tests := []struct {
		name                    string
		config, mixin, expected map[string]interface{}
	}{
		{
			name:     "append-string-to-string",
			config:   mm("key", "base"),
			mixin:    mm("key__APPEND", "new"),
			expected: mm("key", list("base", "new")),
		},
		{
			name:     "append-string-to-list",
			config:   mm("key", list("base")),
			mixin:    mm("key__APPEND", "new"),
			expected: mm("key", list("base", "new")),
		},
		{
			name:     "append-list-to-list",
			config:   mm("key", list("base")),
			mixin:    mm("key__APPEND", list("new")),
			expected: mm("key", list("base", "new")),
		},
		{
			name:     "append-one-to-many",
			config:   mm("key", list("base-1", "base-2")),
			mixin:    mm("key__APPEND", "new"),
			expected: mm("key", list("base-1", "base-2", "new")),
		},
		{
			name:     "append-many-to-one",
			config:   mm("key", "base"),
			mixin:    mm("key__APPEND", list("new-1", "new-2")),
			expected: mm("key", list("base", "new-1", "new-2")),
		},
		{
			name:     "prepend-one-to-many",
			config:   mm("key", list("base-1", "base-2")),
			mixin:    mm("key__PREPEND", "new"),
			expected: mm("key", list("new", "base-1", "base-2")),
		},
		{
			name:     "prepend-many-to-one",
			config:   mm("key", "base"),
			mixin:    mm("key__PREPEND", list("new-1", "new-2")),
			expected: mm("key", list("new-1", "new-2", "base")),
		},
		{
			name:     "prepend-list-to-list",
			config:   mm("key", list("base-1", "base-2")),
			mixin:    mm("key__PREPEND", list("new-1", "new-2")),
			expected: mm("key", list("new-1", "new-2", "base-1", "base-2")),
		},
		{
			name:     "prepend-and-append",
			config:   mm("key", "base"),
			mixin:    mm("key__PREPEND", "new-head", "key__APPEND", "new-tail"),
			expected: mm("key", list("new-head", "base", "new-tail")),
		},
		{
			name:     "prepend-and-append-not-case-sensitive",
			config:   mm("key", "base"),
			mixin:    mm("key__PrePend", "new-head", "key__aPpEnd", "new-tail"),
			expected: mm("key", list("new-head", "base", "new-tail")),
		},
		{
			name:     "append-object-to-string",
			config:   mm("key", "base"),
			mixin:    mm("key__APPEND", mm("newKey", "newValue")),
			expected: mm("key", list("base", mm("newKey", "newValue"))),
		},
		{
			name:     "append-nested",
			config:   mm("key", mm("childKey", "childBase")),
			mixin:    mm("key", mm("childKey__APPEND", "childNew")),
			expected: mm("key", mm("childKey", list("childBase", "childNew"))),
		},
		{
			name:     "append-to-none",
			config:   mm(),
			mixin:    mm("key", mm("childKey__APPEND", "childNew")),
			expected: mm("key", mm("childKey", list("childNew"))),
		},
		{
			name:     "prepend-to-none",
			config:   mm(),
			mixin:    mm("key", mm("childKey__PREPEND", "childNew")),
			expected: mm("key", mm("childKey", list("childNew"))),
		},
		{
			name:     "prepend-and-append-to-self",
			config:   mm(),
			mixin:    mm("key__PREPEND", "first", "key", "middle", "key__APPEND", "last"),
			expected: mm("key", list("first", "middle", "last")),
		},
		{
			name:     "short-syntax",
			config:   mm("key", "base"),
			mixin:    mm("key...", "new-tail", "...key", "new-head"),
			expected: mm("key", list("new-head", "base", "new-tail")),
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("#%d_%s", i, test.name), func(t *testing.T) {
			v := load(t, test.config)
			revolveAppendToListKeys(v, test.mixin)
			assert.Equal(t, test.expected, test.mixin)
		})
	}
}

func TestApplyMixins(t *testing.T) {
	base := `---
version: 2
mixins:
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
      run-before__PREPEND: ["mount /backup"]
      run-after__APPEND: ["unmount /backup"]
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
    use:
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

	t.Run("can-replace-string-with-list", func(t *testing.T) { // requires viper>=v1.11.0
		config := load(t, `
  new-before:
    run-before: [new-before-head, new-before-tail]

profiles:
  profile:
    backup:
      use: new-before
      run-before: "base"
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, []string{"new-before-head", "new-before-tail"}, p.Backup.RunBefore)
	})

	t.Run("list-append-does-not-inherit", func(t *testing.T) {
		config := load(t, `
profiles:
  default:
    use: t2
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

	t.Run("list-append-replaces-inherited-list", func(t *testing.T) {
		config := load(t, `
profiles:
  default:
    backup:
      run-before: ["echo profile-default"]
  profile:
    use: t2
    inherit: default
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, []string{"mount /backup"}, p.Backup.RunBefore)
		assert.Equal(t, []string{"unmount /backup"}, p.Backup.RunAfter)
	})

	t.Run("short-syntax-append", func(t *testing.T) {
		config := load(t, `
  short-append:
    ...run-before: "new-begin"
    run-before...: new-end

profiles:
  profile:
    backup:
      use: short-append
      run-before: ["base"]
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, []string{"new-begin", "base", "new-end"}, p.Backup.RunBefore)
	})

	t.Run("use-is-not-in-flags", func(t *testing.T) {
		config := load(t, `
profiles:
  profile:
    use: t1
    a: b
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, map[string][]string{"a": {"b"}}, p.GetCommonFlags().ToMap())
	})

	t.Run("vars", func(t *testing.T) {
		config := load(t, `
profiles:
  profile-simple:
    backup:
      use: t3
  profile:
    backup:
      use:
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

	t.Run("implicit-vars", func(t *testing.T) {
		config := load(t, `
profiles:
  profile:
    backup:
      use:
        - name: t3
          named-source: my-source
  profile-both:
    backup:
      use:
        - name: t3
          vars:
            named-source: my-source
          named-source: my-implicit-source
          another-source: my-other-implicit-source
`)
		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, []string{"my-source", "${another-source}"}, p.Backup.Source)

		p, err = config.getProfile("profile-both")
		assert.NoError(t, err)
		assert.Equal(t, []string{"my-source", "my-other-implicit-source"}, p.Backup.Source)
	})

	t.Run("invalid-mixin-does-not-fail", func(t *testing.T) {
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
    use: tvalid
`)
		mixins := parseMixins(config.viper)
		assert.Contains(t, mixins, "tvalid")
		assert.NotContains(t, mixins, "tinvalid-no-object1")
		assert.NotContains(t, mixins, "tinvalid-no-object2")

		p, err := config.getProfile("profile")
		assert.NoError(t, err)
		assert.Equal(t, "abc 1 false ${list} ${obj}", p.StatusFile)
	})

	t.Run("unknown-use-fails", func(t *testing.T) {
		buffer := bytes.NewBufferString(base + `
profiles:
  profile:
    use: 
      - t1
      - t2
      - tunknown
`)
		_, err := Load(buffer, FormatYAML)
		assert.EqualError(t, err, "failed applying profiles.profile.use: undefined mixin \"tunknown\"")
	})

	t.Run("invalid-use-fails", func(t *testing.T) {
		defaultError := "mixin use must be string or list of strings or list of use objects"
		invalidUses := map[string]string{
			"false":   defaultError,
			"[1]":     "cannot parse mixin use [1]: '' expected a map, got 'int'",
			"[false]": "cannot parse mixin use []: '' expected a map, got 'bool'",
			"1":       defaultError,
		}
		for use, errorMessage := range invalidUses {
			buffer := bytes.NewBufferString(fmt.Sprintf(`
%s
profiles:
  profile:
    use: %s 
`, base, use))
			_, err := Load(buffer, FormatYAML)
			assert.EqualError(t, err, "failed applying profiles.profile.use: "+errorMessage)
		}
	})
}
