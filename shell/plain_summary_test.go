package shell

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeScanBackup(t *testing.T) {
	t.Parallel()

	source := `repository 2e92db7f opened successfully, password is correct
created new cache in /Users/home/Library/Caches/restic

Files:         209 new,     2 changed,    12 unmodified
Dirs:           58 new,     1 changed,    11 unmodified
Added to the repo: 282.768 MiB

processed 223 files, 346.107 MiB in 0:00
snapshot 07ab30a5 saved
`

	if runtime.GOOS == "windows" {
		// change the source
		source = strings.ReplaceAll(source, "\n", platform.LineSeparator)
	}

	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	// Start writing into the pipe, line by line
	go func() {
		lines := strings.Split(source, "\n")
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			writer.WriteString(line + platform.LineSeparator)
		}
		writer.Close()
	}()

	// Read the stream and send back to output buffer
	summary := &monitor.Summary{}
	output := &strings.Builder{}
	err = ScanBackupPlain(reader, summary, output)
	require.NoError(t, err)

	// Check what we read back is right
	assert.Equal(t, source+platform.LineSeparator, output.String())

	// Check the values found are right
	assert.Equal(t, 209, summary.FilesNew)
	assert.Equal(t, 2, summary.FilesChanged)
	assert.Equal(t, 12, summary.FilesUnmodified)
	assert.Equal(t, 58, summary.DirsNew)
	assert.Equal(t, 1, summary.DirsChanged)
	assert.Equal(t, 11, summary.DirsUnmodified)
	assert.Equal(t, uint64(296503738), summary.BytesAdded)
	assert.Equal(t, uint64(362919494), summary.BytesTotal)
	assert.Equal(t, 223, summary.FilesTotal)
}
