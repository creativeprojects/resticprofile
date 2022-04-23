package templates

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateFuncs(t *testing.T) {
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
		{template: `{{ hello }}`, expected: `Hello World`},
	}

	extraFuncs := map[string]any{
		"hello": func() string { return "Hello World" },
	}

	tpl := New("test-template", extraFuncs)
	buffer := &strings.Builder{}

	for _, test := range tests {

		t.Run(test.template, func(t *testing.T) {
			t2, err := tpl.Parse(test.template)
			require.NoError(t, err)
			require.NotNil(t, t2)

			buffer.Reset()
			err = t2.Execute(buffer, nil)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, buffer.String())
		})
	}
}
