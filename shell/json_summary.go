package shell

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"runtime"

	"github.com/creativeprojects/resticprofile/monitor"
)

type ResticJsonSummary struct {
	MessageType         string  `json:"message_type"`
	FilesNew            int     `json:"files_new"`
	FilesChanged        int     `json:"files_changed"`
	FilesUnmodified     int     `json:"files_unmodified"`
	DirsNew             int     `json:"dirs_new"`
	DirsChanged         int     `json:"dirs_changed"`
	DirsUnmodified      int     `json:"dirs_unmodified"`
	DataBlobs           int     `json:"data_blobs"`
	TreeBlobs           int     `json:"tree_blobs"`
	DataAdded           uint64  `json:"data_added"`
	DataAddedPacked     uint64  `json:"data_added_packed"`
	TotalFilesProcessed int     `json:"total_files_processed"`
	TotalBytesProcessed uint64  `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"`
	SnapshotID          string  `json:"snapshot_id"`
}

// ScanBackupJson should populate the backup summary values from the output of the --json flag
var ScanBackupJson ScanOutput = func(r io.Reader, summary *monitor.Summary, w io.Writer) error {
	bogusPrefix := []byte("\r\x1b[2K")
	jsonPrefix := []byte(`{"message_type":"`)
	summaryPrefix := []byte(`{"message_type":"summary",`)
	jsonSuffix := []byte("}")
	eol := "\n"
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimPrefix(line, bogusPrefix)
		if bytes.HasPrefix(line, jsonPrefix) && bytes.HasSuffix(line, jsonSuffix) {
			if bytes.HasPrefix(line, summaryPrefix) {
				jsonSummary := ResticJsonSummary{}
				err := json.Unmarshal(line, &jsonSummary)
				if err != nil {
					continue
				}
				summary.FilesNew = jsonSummary.FilesNew
				summary.FilesChanged = jsonSummary.FilesChanged
				summary.FilesUnmodified = jsonSummary.FilesUnmodified
				summary.DirsNew = jsonSummary.DirsNew
				summary.DirsChanged = jsonSummary.DirsChanged
				summary.DirsUnmodified = jsonSummary.DirsUnmodified
				summary.FilesTotal = jsonSummary.TotalFilesProcessed
				summary.BytesAdded = jsonSummary.DataAdded
				summary.BytesAddedPacked = jsonSummary.DataAddedPacked
				summary.BytesTotal = jsonSummary.TotalBytesProcessed
			}
			continue
		}
		// write to the output if the line wasn't a json message
		_, _ = w.Write(line)
		_, _ = w.Write([]byte(eol))
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
