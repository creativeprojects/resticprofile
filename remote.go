package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/spf13/afero"
)

func loadRemoteConfiguration(fs afero.Fs, endpoint string) (*remote.Manifest, error) {
	var parameters *remote.Manifest

	client := http.DefaultClient
	request, err := http.NewRequest("GET", endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Accept", "application/x-tar")

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := &bytes.Buffer{}
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("http error %d: %q", resp.StatusCode, strings.TrimSpace(buf.String()))
	}

	if resp.Header.Get("Content-Type") != "application/x-tar" {
		return nil, fmt.Errorf("unexpected content type: %s", resp.Header.Get("Content-Type"))
	}

	reader := tar.NewReader(resp.Body)
	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}
		if !filepath.IsLocal(hdr.Name) {
			return nil, fmt.Errorf("invalid file name: %s", hdr.Name)
		}
		if hdr.Name == remote.ManifestFilename {
			clog.Debugf("downloading manifest (%d bytes)", hdr.Size)
			parameters, err = getManifestParameters(reader, hdr.Size)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %w", err)
			}
		} else {
			clog.Debugf("downloading file %s (%d bytes)", hdr.Name, hdr.Size)
			err = copyFile(fs, reader, hdr.Name, hdr.Size)
			if err != nil {
				return nil, fmt.Errorf("failed to copy file: %w", err)
			}
		}
	}

	return parameters, nil
}

func copyFile(fs afero.Fs, reader io.Reader, filename string, size int64) error {
	file, err := fs.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	read, err := io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	if read != size {
		return fmt.Errorf("file size mismatch: expected %d, got %d", size, read)
	}
	return nil
}

func getManifestParameters(reader io.Reader, size int64) (*remote.Manifest, error) {
	manifest := &remote.Manifest{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}
	return manifest, nil
}
