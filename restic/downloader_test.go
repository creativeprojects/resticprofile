package restic

import (
	"context"
	"fmt"
	"path"
	"testing"

	sup "github.com/creativeprojects/go-selfupdate"
	"github.com/stretchr/testify/assert"
)

func TestDownloadBinary(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	temp := t.TempDir()

	versions := []string{
		"latest",
		"0.13.0",
		"0.12.1",
	}

	for _, version := range versions {
		version := version
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			executable := path.Join(temp, fmt.Sprintf("restic-%s", version))
			err := DownloadBinary(executable, version)
			assert.NoError(t, err)

			actualVersion, err := GetVersion(executable)
			assert.NoError(t, err)
			assert.NotEmpty(t, actualVersion)
			if version == "latest" {
				assert.NotEqual(t, version, actualVersion)
			} else {
				assert.Equal(t, version, actualVersion)
			}
		})
	}
}

func TestSourceChecksRepo(t *testing.T) {
	var err error
	ctx := context.Background()
	_, _, err = defaultUpdater.DetectVersion(ctx, sup.NewRepositorySlug("other", repo), "latest")
	assert.EqualError(t, err, `expected owner "restic" == "other" && repo "restic" == "restic"`)
	_, _, err = defaultUpdater.DetectVersion(ctx, sup.NewRepositorySlug(owner, "other"), "latest")
	assert.EqualError(t, err, `expected owner "restic" == "restic" && repo "restic" == "other"`)
}
