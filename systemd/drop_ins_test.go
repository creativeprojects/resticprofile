//go:build !darwin && !windows

package systemd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropInFileExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		file          string
		setupFs       func() afero.Fs
		expectedValue bool
	}{
		{
			name: "file exists",
			file: "/path/to/file",
			setupFs: func() afero.Fs {
				testFs := afero.NewMemMapFs()
				_, err := testFs.Create("/path/to/file")
				require.NoError(t, err)
				return testFs
			},
			expectedValue: true,
		},
		{
			name: "file does not exist",
			file: "/path/to/nonexistent",
			setupFs: func() afero.Fs {
				testFs := afero.NewMemMapFs()
				return testFs
			},
			expectedValue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testFs := tc.setupFs()

			unit := Unit{
				fs: testFs,
			}

			result := unit.DropInFileExists(tc.file)
			assert.Equal(t, tc.expectedValue, result)
		})
	}
}
