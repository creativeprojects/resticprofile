package shell

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"text/tabwriter"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestArgumentEscape(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("Not running on Windows")
	}

	testData := []struct {
		input               string
		expectedEscape      string
		expectedNoGlobQuote string
	}{
		{"", "", ""},
		{"simple", "simple", "simple"},
		{"Côte d'Ivoire", `Côte\ d\'Ivoire`, `"Côte d'Ivoire"`},
		{"quo\"te", `quo\"te`, `"quo\"te"`},
		{"quo'te", `quo\'te`, `"quo'te"`},
		{"with space", `with\ space`, `"with space"`},
		{`quo\"te`, `quo\"te`, `quo\"te`},
		{`quo\'te`, `quo\'te`, `quo\'te`},
		{"/path/with space; echo foo", "/path/with\\ space\\;\\ echo\\ foo", "\"/path/with space; echo foo\""},
		{"**/.git/", "**/.git/", "\"**/.git/\""},
		{"[aA]*", "[aA]*", "\"[aA]*\""},
		{"/dir", "/dir", "/dir"},
		{"~/dir", "~/dir", "~/dir"},
		{"/\\*\\*/.git", "/\\*\\*/.git", "/\\*\\*/.git"},
		{`/?`, `/?`, `"/?"`},
		{`/\?`, `/\?`, `/\?`},
		{`/\\?`, `/\\?`, `/\\?`},
		{`/\\\?`, `/\\\?`, `/\\\?`},
		{`/\\\\?`, `/\\\\?`, `/\\\\?`},
		{`/ ?*`, `/\ ?*`, `"/ ?*"`},
	}

	output := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.Debug)
	fmt.Fprintf(output, "\tArgument\tShell escape with glob\tNot allowing Glob\t\n")
	for _, testItem := range testData {
		fmt.Fprintf(output, "\t%s\t%s\t%s\t\n",
			displayEscapedString(testItem.input),
			displayEscapedString(testItem.expectedEscape),
			displayEscapedString(testItem.expectedNoGlobQuote))
		t.Run(testItem.input, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testItem.expectedEscape, NewArg(testItem.input, ArgConfigEscape).String())
			assert.Equal(t, testItem.expectedNoGlobQuote, NewArg(testItem.input, ArgConfigKeepGlobQuote).String())
		})
	}
	output.Flush()
}

func displayEscapedString(input string) string {
	output := fmt.Sprintf("%q", input)
	if output[0] == '"' && output[len(output)-1] == '"' {
		output = output[1 : len(output)-1]
	}
	return output
}

func TestEmptyArgValue(t *testing.T) {
	t.Parallel()

	noValue := NewArg("", ArgConfigKeepGlobQuote)
	emptyValue := NewEmptyValueArg()

	assert.False(t, noValue.HasValue())
	assert.False(t, noValue.IsEmptyValue())
	assert.Equal(t, "", noValue.Value())

	assert.True(t, emptyValue.HasValue())
	assert.True(t, emptyValue.IsEmptyValue())
	assert.Equal(t, "", emptyValue.Value())
	if !platform.IsWindows() {
		assert.Equal(t, `""`, emptyValue.String())
	}
}
