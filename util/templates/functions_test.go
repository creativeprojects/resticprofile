package templates

import (
	"fmt"
	"github.com/creativeprojects/resticprofile/platform"
	"io/fs"
	"math/rand"
	"os"
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
		{template: `{{ "some string" | contains "some" }}`, expected: `true`},
		{template: `{{ "some string" | contains "else" }}`, expected: `false`},
		{template: `{{ "1 2 3 5" | split " " | contains "3" }}`, expected: `true`},
		{template: `{{ "1 2 3 5" | split " " | contains 03.0 }}`, expected: `true`},
		{template: `{{ "1 2 3 5" | split " " | contains "23" }}`, expected: `false`},
		{template: `{{ "some string" | matches "^.+str.+$" }}`, expected: `true`},
		{template: `{{ "some string" | matches "^.+else.+$" }}`, expected: `false`},
		{template: `{{ "some old string" | replace "old" "new" }}`, expected: `some new string`},
		{template: `{{ "some old string" | replaceR "(old)" "$1 and new" }}`, expected: `some old and new string`},
		{template: `{{ "some old string" | regex "(old)" "$1 and new" }}`, expected: `some old and new string`},
		{template: `{{ "ABC" | lower }}`, expected: `abc`},
		{template: `{{ "abc" | upper }}`, expected: `ABC`},
		{template: `{{ "  A " | trim }}`, expected: `A`},
		{template: `{{ "--A-" | trimPrefix "--" }}`, expected: `A-`},
		{template: `{{ "--A-" | trimSuffix "-" }}`, expected: `--A`},
		{template: `{{ "A,B,C" | split "," | join ";" }}`, expected: `A;B;C`},
		{template: `{{ "A , B,C" | splitR "\\s*,\\s*" | join ";" }}`, expected: `A;B;C`},
		{template: `{{ list "A" "B" "C" | join ";" }}`, expected: `A;B;C`},
		{template: `{{ "1 2 3 5" | split " " | list | join "-" }}`, expected: `[1,2,3,5]`},
		{template: `{{ range $v := "A,B,C" | split "," }} {{ $v }} {{ end }}`, expected: ` A  B  C `},
		{template: `{{ range $v := list "A" "B" "C" }} {{ $v }} {{ end }}`, expected: ` A  B  C `},
		{template: `{{ with map "k1" "v1" "k2" "v2" }} {{ .k1 }}-{{ .k2 }} {{ end }}`, expected: ` v1-v2 `},
		{template: `{{ with map "k1" nil nil "v2" }} {{ .k1 }}-{{ .k2 }} {{ end }}`, expected: ` <no value>-<no value> `},
		{template: `{{ with list "A" "B" nil "D" | map }} {{ ._0 }}-{{ ._1 }}-{{ ._2 }}-{{ ._3 }} {{ end }}`, expected: ` A-B-<no value>-D `},
		{template: `{{ with list "A" "B" nil "D" | map "key" }} {{ .key | join "-" }} {{ end }}`, expected: ` A-B-<nil>-D `},
		{template: `{{ tempDir }}`, expected: dir},
		{template: `{{ tempDir }}`, expected: dir}, // constant results when repeated
		{template: `{{ tempFile "test.txt" }}`, expected: file},
		{template: `{{ tempFile "test.txt" }}`, expected: file}, // constant results when repeated
		{template: `{{ env }}`, expected: TempFile(".env.none")},
		{template: `{{ "a & b\n" | html }}`, expected: "a &amp; b\n"},
		{template: `{{ "a & b\n" | urlquery }}`, expected: "a+%26+b%0A"},
		{template: `{{ "a & b\n" | js }}`, expected: "a \\u0026 b\\u000A"},
		{template: `{{ "plain" | hex }}`, expected: "706c61696e"},
		{template: `{{ "plain" | base64 }}`, expected: "cGxhaW4="},
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

	t.Run("envFileFunc", func(t *testing.T) {
		profileKey := fmt.Sprintf("prof-%d", int(rand.Uint64()))
		expectedFile := TempFile(fmt.Sprintf("%s.env", profileKey))

		var received []string
		receiveFunc := func(f string) { received = append(received, f) }
		extras := EnvFileFunc(func() (string, func(string)) { return profileKey, receiveFunc })

		tpl, err := New("test-template", extras).Parse(`{{ env }}`)
		require.NoError(t, err)
		require.NotNil(t, tpl)
		assert.Nil(t, received)

		for i := 0; i < 3; i++ {
			received = nil
			buffer.Reset()
			err = tpl.Execute(buffer, nil)
			assert.NoError(t, err)

			assert.Equal(t, expectedFile, buffer.String())
			assert.Equal(t, []string{expectedFile}, received)
		}

		stat, err := os.Stat(expectedFile)
		require.NoError(t, err)
		if platform.IsWindows() {
			assert.Equal(t, fs.FileMode(0666), stat.Mode().Perm()) // adjust when go for Windows supports perms
		} else {
			assert.Equal(t, fs.FileMode(0600), stat.Mode().Perm())
		}
	})
}
