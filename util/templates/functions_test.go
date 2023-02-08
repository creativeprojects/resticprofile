package templates

import (
	"path"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateFuncs(t *testing.T) {
	defer util.ClearTempDir()
	dir := TempDir()
	file := TempFile("test.txt")

	tests := []struct {
		template, expected string
	}{
		{template: `{{ "some old string" | replace "old" "new" }}`, expected: `some new string`},
		{template: `{{ "some old string" | regex "(old)" "$1 and new" }}`, expected: `some old and new string`},
		{template: `{{ "ABC" | lower }}`, expected: `abc`},
		{template: `{{ "abc" | upper }}`, expected: `ABC`},
		{template: `{{ "  A " | trim }}`, expected: `A`},
		{template: `{{ "--A-" | trimPrefix "--" }}`, expected: `A-`},
		{template: `{{ "--A-" | trimSuffix "-" }}`, expected: `--A`},
		{template: `{{ "A,B,C" | split "," | join ";" }}`, expected: `A;B;C`},
		{template: `{{ list "A" "B" "C" | join ";" }}`, expected: `A;B;C`},
		{template: `{{ range $v := "A,B,C" | split "," }} {{ $v }} {{ end }}`, expected: ` A  B  C `},
		{template: `{{ range $v := list "A" "B" "C" }} {{ $v }} {{ end }}`, expected: ` A  B  C `},
		{template: `{{ tempDir }}`, expected: dir},
		{template: `{{ tempDir }}`, expected: dir}, // constant results when repeated
		{template: `{{ tempFile "test.txt" }}`, expected: file},
		{template: `{{ tempFile "test.txt" }}`, expected: file}, // constant results when repeated
		{template: `{{ hello }}`, expected: `Hello World`},
	}

	extraFuncs := map[string]any{
		"hello": func() string { return "Hello World" },
	}

	buffer := &strings.Builder{}

	for _, test := range tests {
		t.Run(test.template, func(t *testing.T) {
			tpl, err := New("test-template", extraFuncs).Parse(test.template)
			require.NoError(t, err)
			require.NotNil(t, tpl)

			buffer.Reset()
			err = tpl.Execute(buffer, nil)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, buffer.String())
		})
	}

	t.Run("tempFile", func(t *testing.T) {
		expected := path.Join(TempDir(), "tf.txt")
		assert.NoFileExists(t, expected)

		file := TempFile("tf.txt")
		assert.Equal(t, expected, file)
		assert.FileExists(t, file)
	})
}
