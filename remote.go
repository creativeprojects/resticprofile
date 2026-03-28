package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/fuse"
	"github.com/creativeprojects/resticprofile/remote"
)

func loadRemoteFiles(ctx context.Context, endpoint string) ([]fuse.File, *remote.Manifest, error) {
	var parameters *remote.Manifest

	client := http.DefaultClient
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Accept", "application/x-tar")

	resp, err := client.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := &bytes.Buffer{}
		_, _ = buf.ReadFrom(resp.Body)
		return nil, nil, fmt.Errorf("http error %d: %q", resp.StatusCode, strings.TrimSpace(buf.String()))
	}

	if resp.Header.Get("Content-Type") != "application/x-tar" {
		return nil, nil, fmt.Errorf("unexpected content type: %s", resp.Header.Get("Content-Type"))
	}

	files := []fuse.File{}
	reader := tar.NewReader(resp.Body)
	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read tar header: %w", err)
		}
		if !filepath.IsLocal(hdr.Name) {
			return nil, nil, fmt.Errorf("invalid file name: %s", hdr.Name)
		}
		if hdr.Name == constants.ManifestFilename {
			clog.Debugf("downloading manifest (%d bytes)", hdr.Size)
			parameters, err = getManifestParameters(reader)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read manifest: %w", err)
			}
		} else {
			clog.Debugf("downloading file %s (%d bytes)", hdr.Name, hdr.Size)
			data := make([]byte, hdr.Size)
			read, err := reader.Read(data)
			if err != nil && err != io.EOF {
				return nil, nil, fmt.Errorf("failed to download file content: %w", err)
			}
			if read != int(hdr.Size) {
				return nil, nil, fmt.Errorf("file size mismatch: expected %d, got %d", hdr.Size, read)
			}
			files = append(files, *fuse.NewFile(hdr.Name, hdr.FileInfo(), data))
		}
	}

	return files, parameters, nil
}

func getManifestParameters(reader io.Reader) (*remote.Manifest, error) {
	manifest := &remote.Manifest{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}
	return manifest, nil
}

// setupRemoteConfiguration downloads the configuration files from the remote endpoint and mounts the virtual FS
func setupRemoteConfiguration(ctx context.Context, remoteEndpoint string) (func(), *remote.Manifest, error) {
	files, parameters, err := loadRemoteFiles(ctx, remoteEndpoint)
	if err != nil {
		return nil, nil, err
	}

	closeMountpoint := func() {}
	mountpoint := parameters.Mountpoint
	if mountpoint == "" {
		// generates a temporary directory
		mountpoint, err = os.MkdirTemp("", "resticprofile-")
		if err != nil {
			return nil, parameters, fmt.Errorf("failed to create mount directory: %w", err)
		}
		closeMountpoint = func() {
			err = os.Remove(mountpoint)
			if err != nil {
				clog.Errorf("failed to remove mountpoint: %v", err)
			}
		}
	}

	closeFs, err := fuse.MountFS(mountpoint, files)
	if err != nil {
		return closeMountpoint, parameters, err
	}

	wd, _ := os.Getwd()
	err = os.Chdir(mountpoint)
	if err != nil {
		return func() {
			closeFs()
			closeMountpoint()
		}, parameters, fmt.Errorf("failed to change directory: %w", err)
	}

	return func() {
		_ = os.Chdir(wd)
		closeFs()
		closeMountpoint()
	}, parameters, nil
}
