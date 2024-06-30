package main

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"

	"github.com/creativeprojects/clog"
	"github.com/spf13/afero"
)

func loadRemoteConfiguration(fs afero.Fs, endpoint string) error {
	client := http.DefaultClient
	request, err := http.NewRequest("GET", endpoint, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Accept", "application/x-tar")

	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "application/x-tar" {
		return fmt.Errorf("unexpected content type: %s", resp.Header.Get("Content-Type"))
	}

	reader := tar.NewReader(resp.Body)
	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}
		clog.Debugf("downloading file %s (%d bytes)", hdr.Name, hdr.Size)
		copyFile(fs, reader, hdr.Name)
	}

	return nil
}

func copyFile(fs afero.Fs, reader io.Reader, filename string) error {
	file, err := fs.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	return nil
}
