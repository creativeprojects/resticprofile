package ansi

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

var ansiColor = func() (c *color.Color) {
	c = color.New(color.FgCyan)
	c.EnableColor()
	return
}()

var colored = ansiColor.SprintFunc()

func TestLineLengthWriter(t *testing.T) {
	tests := []struct {
		input, expected string
		chunks, scale   int
	}{
		// test non-breakable
		{input: strings.Repeat("-", 50), expected: strings.Repeat("-", 50), chunks: 15},

		// test breakable without columns
		{
			input: strings.Repeat("word ", 20),
			expected: "" +
				strings.TrimSpace(strings.Repeat("word ", 8)) + "\n" +
				strings.TrimSpace(strings.Repeat("word ", 8)) + "\n" +
				strings.Repeat("word ", 4),
			chunks: 5, scale: 6,
		},

		// test breakable with ANSI color
		{
			input: strings.Repeat(colored("word "), 20),
			expected: "" +
				strings.Repeat(colored("word "), 7) + colored("word\n") +
				strings.Repeat(colored("word "), 7) + colored("word\n") +
				strings.Repeat(colored("word "), 4),
		},

		// test breakable with 2 columns
		{
			input: "word word  word word  " +
				strings.Repeat("word ", 20),
			expected: "" +
				"word word  word word  " +
				"word word word\n" +
				strings.Repeat("                      word word word\n", 5) +
				"                      word word ",
			chunks: 3, scale: 15,
		},

		// test breakable with 2 columns and ANSI color
		{
			input: colored("word word  word word  ") +
				strings.Repeat(colored("word "), 20),
			expected: "" +
				colored("word word  word word  ") +
				colored("word ") + colored("word ") + colored("word\n                      ") +
				strings.Repeat(colored("word ")+colored("word ")+colored("word\n                      "), 5) +
				colored("word ") + colored("word "),
		},

		// test breakable with 2 columns and unicode character
		{
			input: "w😁rd wo😁d  wor😁  😁ord 😁😁😁😁  w😁rd wo😁d  wor😁 😁ord 😁😁😁😁",
			expected: "w😁rd wo😁d  wor😁  😁ord 😁😁😁😁  w😁rd wo😁d \n" +
				"                 wor😁 😁ord 😁😁😁😁",
		},
		{
			input: "word word  word  word word  word word  word word word",
			expected: "word word  word  word word  word word \n" +
				"                 word word word",
		},

		// test breakable with 2 columns, colors and unicode character
		{
			input: colored("w😁rd wo😁d  wor😁 😁ord  ") +
				strings.Repeat(colored("wor😁 ")+colored("😁ord "), 2),
			expected: "" +
				colored("w😁rd wo😁d  wor😁 😁ord  ") +
				colored("wor😁 ") + colored("😁ord ") + colored("wor😁\n                      ") +
				colored("😁ord "),
		},

		// test real-world content
		{
			input: `
Usage of resticprofile:
   resticprofile [resticprofile flags] [profile name.][restic command] [restic flags]
   resticprofile [resticprofile flags] [profile name.][resticprofile command] [command specific flags]

resticprofile flags:
  -c, --config string        configuration file (default "profiles")
      --dry-run              display the restic commands instead of running them
  -f, --format string        file format of the configuration (default is to use the file extension)
  -h, --help                 display this help
      --lock-wait duration   wait up to duration to acquire a lock (syntax "1h5m30s")
  -l, --log string           logs to a target instead of the console
  -n, --name string          profile name (default "default")
      --no-ansi              disable ansi control characters (disable console colouring)
      --no-lock              skip profile lock file
      --no-prio              don't set any priority on load: used when started from a service that has already set the priority
  -q, --quiet                display only warnings and errors
      --theme string         console colouring theme (dark, light, none) (default "light")
      --trace                display even more debugging information
  -v, --verbose              display some debugging information
  -w, --wait                 wait at the end until the user presses the enter key


resticprofile own commands:
   help          display help (run in verbose mode for detailed information)
   version       display version (run in verbose mode for detailed information)
   self-update   update to latest resticprofile (use -q/--quiet flag to update without confirmation)
   profiles      display profile names from the configuration file
   show          show all the details of the current profile
   schedule      schedule jobs from a profile (use --all flag to schedule all jobs of all profiles)
   unschedule    remove scheduled jobs of a profile (use --all flag to unschedule all profiles)
   status        display the status of scheduled jobs (use --all flag for all profiles)
   generate      generate resources (--random-key [size], --bash-completion & --zsh-completion)

Documentation available at https://creativeprojects.github.io/resticprofile/
`,
			expected: `
Usage of resticprofile:
   resticprofile [resticprofile flags]
   [profile name.][restic command]
   [restic flags]
   resticprofile [resticprofile flags]
   [profile name.][resticprofile
   command] [command specific flags]

resticprofile flags:
  -c, --config string       
                         configuration
                         file (default
                         "profiles")
      --dry-run              display
                         the restic
                         commands
                         instead of
                         running them
  -f, --format string        file
                         format of the
                         configuration
                         (default is to
                         use the file
                         extension)
  -h, --help                 display
                         this help
      --lock-wait duration   wait up to
      duration to acquire a lock
      (syntax "1h5m30s")
  -l, --log string           logs to a
                         target instead
                         of the console
  -n, --name string          profile
                         name (default
                         "default")
      --no-ansi              disable
                         ansi control
                         characters
                         (disable
                         console
                         colouring)
      --no-lock              skip
                         profile lock
                         file
      --no-prio              don't set
                         any priority
                         on load: used
                         when started
                         from a service
                         that has
                         already set
                         the priority
  -q, --quiet                display
                         only warnings
                         and errors
      --theme string         console
                         colouring
                         theme (dark,
                         light, none)
                         (default
                         "light")
      --trace                display
                         even more
                         debugging
                         information
  -v, --verbose              display
                         some debugging
                         information
  -w, --wait                 wait at
                         the end until
                         the user
                         presses the
                         enter key


resticprofile own commands:
   help          display help (run in
                 verbose mode for
                 detailed information)
   version       display version (run
                 in verbose mode for
                 detailed information)
   self-update   update to latest
                 resticprofile (use
                 -q/--quiet flag to
                 update without
                 confirmation)
   profiles      display profile names
                 from the configuration
                 file
   show          show all the details
                 of the current profile
   schedule      schedule jobs from a
                 profile (use --all
                 flag to schedule all
                 jobs of all profiles)
   unschedule    remove scheduled jobs
                 of a profile (use
                 --all flag to
                 unschedule all
                 profiles)
   status        display the status of
                 scheduled jobs (use
                 --all flag for all
                 profiles)
   generate      generate resources
                 (--random-key [size],
                 --bash-completion &
                 --zsh-completion)

Documentation available at
https://creativeprojects.github.io/resticprofile/
`,
		},
	}

	for i, test := range tests {
		if test.scale == 0 {
			test.scale = 1
		}
		for chunkSize := 0; chunkSize <= test.chunks; chunkSize++ {
			t.Run(fmt.Sprintf("%d-%d", i, chunkSize), func(t *testing.T) {
				buffer := bytes.Buffer{}
				writer := NewLineLengthWriter(&buffer, 40)
				input := []byte(test.input)

				var (
					n   int
					err error
				)
				switch chunkSize {
				case 0:
					n, err = writer.Write(input)
					assert.Equal(t, len(input), n)
				default:
					for len(input) > 0 && err == nil {
						length := test.scale * chunkSize
						if length > len(input) {
							length = len(input)
						}
						n, err = writer.Write(input[:length])
						assert.Equal(t, length, n)
						input = input[length:]
					}
				}

				assert.Nil(t, err)
				assert.Equal(t, test.expected, buffer.String())
			})
		}
	}
}
