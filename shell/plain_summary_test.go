package shell

import (
	"fmt"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
	"time"
)

func TestPipeScanBackup(t *testing.T) {
	sources := []string{`
repository 2e92db7f opened successfully, password is correct
created new cache in /Users/home/Library/Caches/restic

Files:         209 new,     2 changed,    12 unmodified
Dirs:           58 new,     1 changed,    11 unmodified
Added to the repo: 282.768 MiB

processed 223 files, 346.107 MiB in 0:02
snapshot 07ab30a5 saved
	`, `
repository 2e92db7f opened successfully, password is correct
created new cache in /Users/home/Library/Caches/restic

Files:         209 new,     2 changed,    12 unmodified
Dirs:           58 new,     1 changed,    11 unmodified
Added to the repository: 282.768 MiB

processed 223 files, 346.107 MiB in 0:02
snapshot 07ab30a5 saved
	`, `
repository 2e92db7f opened successfully, password is correct
created new cache in /Users/home/Library/Caches/restic

Files:         209 new,     2 changed,    12 unmodified
Dirs:           58 new,     1 changed,    11 unmodified
Added to the repository: 282.768 MiB (140.641 MiB stored)

processed 223 files, 346.107 MiB in 0:02
snapshot 07ab30a5 saved
	`}

	for i, source := range sources {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {

			if platform.IsWindows() {
				// change the source
				source = strings.ReplaceAll(source, "\n", eol)
			}

			reader, writer := io.Pipe()
			defer reader.Close()
			go writeLinesToFile(source, writer)

			// Read the stream and send back to output buffer
			provider := monitor.NewProvider()
			output := &strings.Builder{}
			err := scanBackupPlain(reader, provider, output)
			require.NoError(t, err)

			// Check what we read back is right
			assert.Equal(t, source+eol, output.String())

			// Check the values found are right
			summary := provider.CurrentSummary()
			assert.Equal(t, int64(209), summary.FilesNew)
			assert.Equal(t, int64(2), summary.FilesChanged)
			assert.Equal(t, int64(12), summary.FilesUnmodified)
			assert.Equal(t, int64(58), summary.DirsNew)
			assert.Equal(t, int64(1), summary.DirsChanged)
			assert.Equal(t, int64(11), summary.DirsUnmodified)
			assert.Equal(t, unformatBytes(282.768, "MiB"), summary.BytesAdded)
			if i == 2 {
				assert.Equal(t, unformatBytes(140.641, "MiB"), summary.BytesStored)
			}
			assert.Equal(t, unformatBytes(346.107, "MiB"), summary.BytesTotal)
			assert.Equal(t, int64(223), summary.FilesTotal)
			assert.Equal(t, "07ab30a5", summary.SnapshotID)
			assert.False(t, summary.Extended)
			assert.Equal(t, time.Second*2, summary.Duration)
		})
	}
}

func TestUnformatBytes(t *testing.T) {
	assert.Equal(t, uint64(1), unformatBytes(1, "Bytes"))
	assert.Equal(t, uint64(1024), unformatBytes(1, "KiB"))
	assert.Equal(t, uint64(512), unformatBytes(0.5, "KiB"))
	assert.Equal(t, uint64(1024*1024), unformatBytes(1, "MiB"))
	assert.Equal(t, uint64(1024*1024*1024), unformatBytes(1, "GiB"))
	assert.Equal(t, uint64(1024*1024*1024*1024), unformatBytes(1, "TiB"))
}
