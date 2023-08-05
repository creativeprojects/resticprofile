package shell

import (
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/monitor/mocks"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

func TestScanJsonSummary(t *testing.T) {
	// example of restic output (beginning and end of the output)
	resticOutput := `{"message_type":"status","percent_done":0,"total_files":1,"total_bytes":10244}
{"message_type":"status","percent_done":0.028711419769115988,"total_files":213,"files_done":13,"total_bytes":362948126,"bytes_done":10420756,"current_files":["/go/src/github.com/creativeprojects/resticprofile/build/restic","/go/src/github.com/creativeprojects/resticprofile/build/resticprofile"]}
{"message_type":"status","percent_done":0.9763572825280271,"error_count":1,"total_files":213,"files_done":163,"total_bytes":362948126,"bytes_done":354367046,"current_files":["/go/src/github.com/creativeprojects/resticprofile/resticprofile_darwin","/go/src/github.com/creativeprojects/resticprofile/resticprofile_linux"]}
{"message_type":"status","seconds_elapsed":1,"percent_done":1,"error_count":2,"total_files":213,"files_done":212,"total_bytes":362948126,"bytes_done":362948126,"current_files":["/go/src/github.com/creativeprojects/resticprofile/resticprofile_linux"]}
{"message_type":"summary","files_new":213,"files_changed":11,"files_unmodified":12,"dirs_new":58,"dirs_changed":2,"dirs_unmodified":3,"data_blobs":402,"tree_blobs":59,"data_added":296530781,"total_files_processed":236,"total_bytes_processed":362948126,"total_duration":1.009156009,"snapshot_id":"6daa8ef6"}
`

	if platform.IsWindows() {
		// change the source
		resticOutput = strings.ReplaceAll(resticOutput, "\n", eol)
	}

	reader, writer := io.Pipe()
	defer reader.Close()
	go writeLinesToFile(resticOutput, writer)

	// Read the stream and send back to output buffer
	var status monitor.Status
	receiver := mocks.NewReceiver(t)
	receiver.EXPECT().Status(mock.Anything).Run(func(s monitor.Status) {
		status = s
	}).Times(4)
	provider := monitor.NewProvider(receiver)
	output := util.NewSyncWriter(&strings.Builder{})

	err := scanBackupJson(reader, provider, HideJson(output))
	require.NoError(t, err)

	// Check what we read back is right (should be trailing newline only)
	_ = output.Locked(func(out *strings.Builder) error {
		assert.Equal(t, eol, out.String())
		return nil
	})

	// Check the status
	assert.Equal(t, []string{"/go/src/github.com/creativeprojects/resticprofile/resticprofile_linux"}, status.CurrentFiles)
	assert.Equal(t, int64(2), status.ErrorCount)
	assert.Equal(t, float64(1), status.PercentDone)
	assert.Equal(t, int64(213), status.TotalFiles)
	assert.Equal(t, uint64(362948126), status.TotalBytes)
	assert.Equal(t, int64(212), status.FilesDone)
	assert.Equal(t, uint64(362948126), status.BytesDone)
	assert.Equal(t, 1, status.SecondsElapsed)
	assert.Equal(t, 0, status.SecondsRemaining)

	// Check the summary
	summary := provider.CurrentSummary()
	assert.Equal(t, int64(213), summary.FilesNew)
	assert.Equal(t, int64(11), summary.FilesChanged)
	assert.Equal(t, int64(12), summary.FilesUnmodified)
	assert.Equal(t, int64(58), summary.DirsNew)
	assert.Equal(t, int64(2), summary.DirsChanged)
	assert.Equal(t, int64(3), summary.DirsUnmodified)
	assert.Equal(t, uint64(296530781), summary.BytesAdded)
	assert.Equal(t, uint64(362948126), summary.BytesTotal)
	assert.Equal(t, int64(236), summary.FilesTotal)
	assert.Equal(t, "6daa8ef6", summary.SnapshotID)
	assert.True(t, summary.Extended)
	assert.Equal(t, time.Duration(math.Round(1.009156009*float64(time.Second))), summary.Duration)
}

func TestScanJsonError(t *testing.T) {
	resticOutput := `Fatal: unable to open config file: Stat: stat /Volumes/RAMDisk/self/config: no such file or directory
Is there a repository at the following location?
/Volumes/RAMDisk/self
`
	if platform.IsWindows() {
		// change the source
		resticOutput = strings.ReplaceAll(resticOutput, "\n", eol)
	}

	reader, writer := io.Pipe()
	defer reader.Close()
	go writeLinesToFile(resticOutput, writer)

	// Read the stream and send back to output buffer
	provider := monitor.NewProvider()
	output := util.NewSyncWriter(&strings.Builder{})
	err := scanBackupJson(reader, provider, HideJson(output))
	require.NoError(t, err)

	// Check what we read back is right
	_ = output.Locked(func(out *strings.Builder) error {
		assert.Equal(t, resticOutput+eol, out.String())
		return nil
	})

	// Check the values found are right
	summary := provider.CurrentSummary()
	assert.Equal(t, int64(0), summary.FilesNew)
	assert.Equal(t, int64(0), summary.FilesChanged)
	assert.Equal(t, int64(0), summary.FilesUnmodified)
	assert.Equal(t, int64(0), summary.DirsNew)
	assert.Equal(t, int64(0), summary.DirsChanged)
	assert.Equal(t, int64(0), summary.DirsUnmodified)
	assert.Equal(t, uint64(0), summary.BytesAdded)
	assert.Equal(t, uint64(0), summary.BytesTotal)
	assert.Equal(t, int64(0), summary.FilesTotal)
	assert.Equal(t, "", summary.SnapshotID)
	assert.False(t, summary.Extended)
}
