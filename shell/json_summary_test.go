package shell

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	summary = `{"message_type":"summary","files_new":10,"files_changed":11,"files_unmodified":211,"dirs_new":0,"dirs_changed":12,"dirs_unmodified":58,"data_blobs":6,"tree_blobs":7,"data_added":8,"total_files_processed":232,"total_bytes_processed":362946952,"total_duration":0.220900201,"snapshot_id":"196f3b36"}`
)

func TestScanJsonSummary(t *testing.T) {
	// example of restic output (beginning and end of the output)
	resticOutput := `{"message_type":"status","percent_done":0,"total_files":1,"total_bytes":10244}
{"message_type":"status","percent_done":0.028711419769115988,"total_files":213,"files_done":13,"total_bytes":362948126,"bytes_done":10420756,"current_files":["/go/src/github.com/creativeprojects/resticprofile/build/restic","/go/src/github.com/creativeprojects/resticprofile/build/resticprofile"]}
{"message_type":"status","percent_done":0.9763572825280271,"total_files":213,"files_done":163,"total_bytes":362948126,"bytes_done":354367046,"current_files":["/go/src/github.com/creativeprojects/resticprofile/resticprofile_darwin","/go/src/github.com/creativeprojects/resticprofile/resticprofile_linux"]}
{"message_type":"status","seconds_elapsed":1,"percent_done":1,"total_files":213,"files_done":212,"total_bytes":362948126,"bytes_done":362948126,"current_files":["/go/src/github.com/creativeprojects/resticprofile/resticprofile_linux"]}
{"message_type":"summary","files_new":213,"files_changed":11,"files_unmodified":12,"dirs_new":58,"dirs_changed":2,"dirs_unmodified":3,"data_blobs":402,"tree_blobs":59,"data_added":296530781,"total_files_processed":236,"total_bytes_processed":362948126,"total_duration":1.009156009,"snapshot_id":"6daa8ef6"}
`

	if runtime.GOOS == "windows" {
		// change the source
		resticOutput = strings.ReplaceAll(resticOutput, "\n", eol)
	}

	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	// Start writing into the pipe, line by line
	go func() {
		lines := strings.Split(resticOutput, "\n")
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			if runtime.GOOS == "windows" {
				// https://github.com/restic/restic/issues/3111
				writer.WriteString("\r\x1b[2K")
			}
			writer.WriteString(line + eol)
		}
		writer.Close()
	}()

	// Read the stream and send back to output buffer
	summary := &monitor.Summary{}
	output := &strings.Builder{}
	err = ScanBackupJson(reader, summary, output)
	require.NoError(t, err)

	// Check what we read back is right (should be empty)
	assert.Equal(t, eol, output.String())

	// Check the values found are right
	assert.Equal(t, 213, summary.FilesNew)
	assert.Equal(t, 11, summary.FilesChanged)
	assert.Equal(t, 12, summary.FilesUnmodified)
	assert.Equal(t, 58, summary.DirsNew)
	assert.Equal(t, 2, summary.DirsChanged)
	assert.Equal(t, 3, summary.DirsUnmodified)
	assert.Equal(t, uint64(296530781), summary.BytesAdded)
	assert.Equal(t, uint64(362948126), summary.BytesTotal)
	assert.Equal(t, 236, summary.FilesTotal)
}

func TestScanJsonError(t *testing.T) {
	resticOutput := `Fatal: unable to open config file: Stat: stat /Volumes/RAMDisk/self/config: no such file or directory
Is there a repository at the following location?
/Volumes/RAMDisk/self
`
	if runtime.GOOS == "windows" {
		// change the source
		resticOutput = strings.ReplaceAll(resticOutput, "\n", eol)
	}

	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	// Start writing into the pipe, line by line
	go func() {
		lines := strings.Split(resticOutput, "\n")
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			writer.WriteString(line + eol)
		}
		writer.Close()
	}()

	// Read the stream and send back to output buffer
	summary := &monitor.Summary{}
	output := &strings.Builder{}
	err = ScanBackupJson(reader, summary, output)
	require.NoError(t, err)

	// Check what we read back is right
	assert.Equal(t, resticOutput+eol, output.String())

	// Check the values found are right
	assert.Equal(t, 0, summary.FilesNew)
	assert.Equal(t, 0, summary.FilesChanged)
	assert.Equal(t, 0, summary.FilesUnmodified)
	assert.Equal(t, 0, summary.DirsNew)
	assert.Equal(t, 0, summary.DirsChanged)
	assert.Equal(t, 0, summary.DirsUnmodified)
	assert.Equal(t, uint64(0), summary.BytesAdded)
	assert.Equal(t, uint64(0), summary.BytesTotal)
	assert.Equal(t, 0, summary.FilesTotal)
}
