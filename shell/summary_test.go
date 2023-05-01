package shell

import (
	"fmt"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

var (
	eol = "\n"
)

func init() {
	if platform.IsWindows() {
		eol = "\r\n"
	}
}

func writeLinesToFile(input string, output io.WriteCloser) {
	defer func() { _ = output.Close() }()

	json := strings.Contains(input, "{") &&
		strings.Contains(input, "}")

	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimRight(line, "\r")
		if platform.IsWindows() && json {
			// https://github.com/restic/restic/issues/3111
			_, _ = fmt.Fprint(output, "\r\x1b[2K")
		}
		_, _ = fmt.Fprint(output, line, eol)
	}
}

func TestGetOutputScanner(t *testing.T) {
	tests := []struct {
		command     string
		plain, json ScanOutput
	}{
		{command: "any", plain: nil, json: nil},
		{command: constants.CommandBackup, plain: scanBackupPlain, json: scanBackupJson},
	}

	ref := func(t any) string { return fmt.Sprintf("%p%#v", t, t) }

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			assert.Equal(t, ref(GetOutputScanner(test.command, false)), ref(test.plain))
			assert.Equal(t, ref(GetOutputScanner(test.command, true)), ref(test.json))
		})
	}
}
