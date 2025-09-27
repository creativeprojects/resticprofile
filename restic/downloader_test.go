package restic

import (
	"context"
	"path/filepath"
	"testing"

	sup "github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadBinary(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip()
	}

	versions := []string{
		"latest",
		"0.13.0",
		"0.12.1",
	}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			executable := platform.Executable(filepath.Join(t.TempDir(), "restic"))
			err := DownloadBinary(executable, version)
			require.NoError(t, err)

			actualVersion, err := GetVersion(executable)
			require.NoError(t, err)
			assert.NotEmpty(t, actualVersion)
			if version == "latest" {
				assert.NotEqual(t, version, actualVersion)
			} else {
				assert.Equal(t, version, actualVersion)
			}
		})
	}
}

func TestNoVersion(t *testing.T) {
	_, err := GetVersion("echo")
	assert.Error(t, err)
}

func TestSourceChecksRepo(t *testing.T) {
	var err error
	ctx := context.Background()
	_, _, err = defaultUpdater.DetectVersion(ctx, sup.NewRepositorySlug("other", repo), "latest")
	assert.EqualError(t, err, `expected owner "restic" == "other" && repo "restic" == "restic"`)
	_, _, err = defaultUpdater.DetectVersion(ctx, sup.NewRepositorySlug(owner, "other"), "latest")
	assert.EqualError(t, err, `expected owner "restic" == "restic" && repo "restic" == "other"`)
}
