package status

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	eol = "\n"
)

func init() {
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	}
}

func TestPipeScan(t *testing.T) {
	source := `repository 2e92db7f opened successfully, password is correct
created new cache in /Users/home/Library/Caches/restic

Files:         209 new,     2 changed,    12 unmodified
Dirs:           58 new,     1 changed,    11 unmodified
Added to the repo: 282.768 MiB

processed 223 files, 346.107 MiB in 0:00
snapshot 07ab30a5 saved
`

	filesNew, filesChanged, filesUnmodified := 0, 0, 0
	dirsNew, dirsChanged, dirsUnmodified := 0, 0, 0
	filesTotal, bytesAdded, bytesTotal := 0, 0.0, 0.0
	bytesAddedUnit, bytesTotalUnit, duration := "", "", ""

	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	// Start writing into the pipe, line by line
	go func() {
		lines := strings.Split(source, "\n")
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			writer.WriteString(line + eol)
		}
		writer.Close()
	}()

	// Read it back
	buffer := &strings.Builder{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		buffer.WriteString(scanner.Text() + eol)
		// scan content - it's all right if the line does not match
		_, _ = fmt.Sscanf(scanner.Text(), "Files: %d new, %d changed, %d unmodified", &filesNew, &filesChanged, &filesUnmodified)
		_, _ = fmt.Sscanf(scanner.Text(), "Dirs: %d new, %d changed, %d unmodified", &dirsNew, &dirsChanged, &dirsUnmodified)
		_, _ = fmt.Sscanf(scanner.Text(), "Added to the repo: %f %3s", &bytesAdded, &bytesAddedUnit)
		_, _ = fmt.Sscanf(scanner.Text(), "processed %d files, %f %3s in %s", &filesTotal, &bytesTotal, &bytesTotalUnit, &duration)
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	// Check what we read back is right
	assert.Equal(t, source+eol, buffer.String())

	// Check the values found are right
	assert.Equal(t, 209, filesNew)
	assert.Equal(t, 2, filesChanged)
	assert.Equal(t, 12, filesUnmodified)
	assert.Equal(t, 58, dirsNew)
	assert.Equal(t, 1, dirsChanged)
	assert.Equal(t, 11, dirsUnmodified)
	assert.Equal(t, 282.768, bytesAdded)
	assert.Equal(t, "MiB", bytesAddedUnit)
	assert.Equal(t, 346.107, bytesTotal)
	assert.Equal(t, "MiB", bytesTotalUnit)
	assert.Equal(t, 223, filesTotal)
	assert.Equal(t, "0:00", duration)
}
