package remote

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fileName string
		data     []byte
	}{
		{
			name:     "send empty file",
			fileName: "empty.txt",
			data:     []byte{},
		},
		{
			name:     "send file with content",
			fileName: "test.txt",
			data:     []byte("test content"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			buf := new(bytes.Buffer)
			tar := NewTar(buf)

			// Execute
			err := tar.SendFile(tt.fileName, tt.data)
			require.NoError(t, err)
			tar.Close()

			// Verify the tar contains the correct file
			fs := afero.NewMemMapFs()
			err = extractTarToFs(buf.Bytes(), fs)
			assert.NoError(t, err)

			// Check file exists and has correct content
			fileExists, err := afero.Exists(fs, tt.fileName)
			assert.NoError(t, err)
			assert.True(t, fileExists)

			content, err := afero.ReadFile(fs, tt.fileName)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, content)
		})
	}
}

func TestSendFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     map[string]string
		filePaths []string
	}{
		{
			name:      "send multiple files",
			files:     map[string]string{"file1.txt": "content1", "file2.txt": "content2"},
			filePaths: []string{"file1.txt", "file2.txt"},
		},
		{
			name:      "send empty file among others",
			files:     map[string]string{"empty.txt": "", "notempty.txt": "content"},
			filePaths: []string{"empty.txt", "notempty.txt"},
		},
		{
			name:      "send non-existent file",
			files:     map[string]string{"exists.txt": "content"},
			filePaths: []string{"exists.txt", "nonexistent.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			buf := new(bytes.Buffer)
			tar := NewTar(buf)

			// Create a memory filesystem with the test files
			memFs := afero.NewMemMapFs()
			for name, content := range tt.files {
				err := afero.WriteFile(memFs, name, []byte(content), 0644)
				assert.NoError(t, err)
			}

			tar.WithFs(memFs)

			// Execute
			err := tar.SendFiles(tt.filePaths)
			require.NoError(t, err)
			tar.Close()

			// Verify the tar contains the correct files
			outputFs := afero.NewMemMapFs()
			err = extractTarToFs(buf.Bytes(), outputFs)
			assert.NoError(t, err)

			// Check each expected file exists and has correct content
			for name, expectedContent := range tt.files {
				// Only check files that were in the filePaths list
				included := false
				for _, path := range tt.filePaths {
					if path == name {
						included = true
						break
					}
				}

				if !included {
					continue
				}

				fileExists, err := afero.Exists(outputFs, name)
				assert.NoError(t, err)

				if _, ok := tt.files[name]; ok {
					assert.True(t, fileExists)

					content, err := afero.ReadFile(outputFs, name)
					assert.NoError(t, err)
					assert.Equal(t, []byte(expectedContent), content)
				}
			}
		})
	}
}

// Helper function to extract tar contents to an afero filesystem
func extractTarToFs(tarData []byte, fs afero.Fs) error {
	reader := bytes.NewReader(tarData)
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			file, err := fs.Create(header.Name)
			if err != nil {
				return err
			}
			if _, err := io.CopyN(file, tr, 1000); err != nil && err != io.EOF {
				file.Close()
				return err
			}
			file.Close()
		}
	}
	return nil
}
