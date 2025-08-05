//go:build !windows

package fuse

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var memfsContents = map[string]string{
	"emptydir/":                     "",
	"file.txt":                      "content",
	"dir/subfile.txt":               "other content",
	"dir with space/other file.txt": "different content",
}

func TestMemFS(t *testing.T) {
	const fileMode = 0o764

	files := make([]File, 0)
	now := time.Now()
	for filename, fileContents := range memfsContents {
		h := &tar.Header{
			Name:    filename,
			Size:    int64(len(fileContents)),
			Mode:    fileMode,
			Uid:     100,
			Gid:     100,
			ModTime: now,
		}

		isDir := strings.HasSuffix(filename, "/")
		if isDir {
			h.Typeflag = tar.TypeDir
		}

		files = append(files, File{
			name:     filename,
			fileInfo: h.FileInfo(),
			data:     []byte(fileContents),
		})
	}

	mnt := t.TempDir()
	closeMount, err := MountFS(mnt, files)
	if err != nil && strings.Contains(err.Error(), "no FUSE mount utility found") {
		t.Skip("no FUSE mount utility found")
	}
	require.NoError(t, err, "cannot mount FS")
	defer closeMount()

	for filename, fileContents := range memfsContents {
		fullPath := filepath.Join(mnt, filename)

		filestat, err := os.Stat(fullPath)
		require.NoErrorf(t, err, "os.Stat %q", filename)

		if strings.HasSuffix(filename, "/") {
			assert.True(t, filestat.IsDir(), "is dir %q", filename)

		} else {
			assert.False(t, filestat.IsDir(), "is file %q", filename)

			contents, err := os.ReadFile(fullPath)
			assert.NoErrorf(t, err, "read %q", filename)

			assert.Equalf(t, fileContents, string(contents), "file %q", filename)
		}
	}
}
