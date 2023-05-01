package shell

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/monitor"
	"io"
	"math"
	"time"
)

type resticJsonMessage struct {
	MessageType string `json:"message_type"`
}

type resticJsonBackupSummary struct {
	resticJsonMessage           // "summary"
	FilesNew            int64   `json:"files_new"`
	FilesChanged        int64   `json:"files_changed"`
	FilesUnmodified     int64   `json:"files_unmodified"`
	DirsNew             int64   `json:"dirs_new"`
	DirsChanged         int64   `json:"dirs_changed"`
	DirsUnmodified      int64   `json:"dirs_unmodified"`
	DataBlobs           int64   `json:"data_blobs"`
	TreeBlobs           int64   `json:"tree_blobs"`
	DataAdded           uint64  `json:"data_added"`
	TotalFilesProcessed int64   `json:"total_files_processed"`
	TotalBytesProcessed uint64  `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"`
	SnapshotID          string  `json:"snapshot_id"`
	DryRun              bool    `json:"dry_run"`
}

type resticJsonBackupStatus struct {
	resticJsonMessage          // "status"
	Action            string   `json:"action"`
	ActionDuration    float64  `json:"duration"`
	SecondsElapsed    int      `json:"seconds_elapsed"`
	SecondsRemaining  int      `json:"seconds_remaining"`
	PercentDone       float64  `json:"percent_done"`
	TotalFiles        int64    `json:"total_files"`
	FilesDone         int64    `json:"files_done"`
	TotalBytes        uint64   `json:"total_bytes"`
	BytesDone         uint64   `json:"bytes_done"`
	ErrorCount        int64    `json:"error_count"`
	CurrentFiles      []string `json:"current_files"`
}

type resticJsonBackupVerboseStatus struct {
	resticJsonMessage          // "verbose_status"
	Action             string  `json:"action"`
	Item               string  `json:"item"`
	Duration           float64 `json:"duration"` // in seconds
	DataSize           uint64  `json:"data_size"`
	DataSizeInRepo     uint64  `json:"data_size_in_repo"`
	MetadataSize       uint64  `json:"metadata_size"`
	MetadataSizeInRepo uint64  `json:"metadata_size_in_repo"`
	TotalFiles         int64   `json:"total_files"`
}

var (
	jsonPrefix              = []byte(`{"message_type":"`)
	jsonSummaryPrefix       = []byte(`{"message_type":"summary",`)
	jsonStatusPrefix        = []byte(`{"message_type":"status",`)
	jsonVerboseStatusPrefix = []byte(`{"message_type":"verbose_status",`)
	jsonSuffix              = []byte("}")
)

// scanBackupJson should populate the backup summary values from the output of the --json flag
var scanBackupJson ScanOutput = func(r io.Reader, provider monitor.Provider, w io.Writer) error {
	scanner := bufio.NewScanner(io.TeeReader(r, w))
	stored, added := uint64(0), uint64(0)
	compressed, uncompressed := uint64(0), uint64(0)

	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimPrefix(line, bogusPrefix)

		if bytes.HasPrefix(line, jsonPrefix) && bytes.HasSuffix(line, jsonSuffix) {
			clog.Infof(string(line))

			var err error
			if bytes.HasPrefix(line, jsonVerboseStatusPrefix) {
				parsed := new(resticJsonBackupVerboseStatus)
				if err = json.Unmarshal(line, &parsed); err == nil {
					switch parsed.Action {
					case "new":
						stored += parsed.DataSizeInRepo + parsed.MetadataSizeInRepo
						added += parsed.DataSize + parsed.MetadataSize
					case "modified":
						compressed += parsed.DataSizeInRepo + parsed.MetadataSizeInRepo
						uncompressed += parsed.DataSize + parsed.MetadataSize
					}
				}
			} else if bytes.HasPrefix(line, jsonStatusPrefix) {
				parsed := new(resticJsonBackupStatus)
				if err = json.Unmarshal(line, &parsed); err == nil && parsed.Action == "" {
					provider.ProvideStatus(func(status *monitor.Status) {
						status.TotalBytes = parsed.TotalBytes
						status.BytesDone = parsed.BytesDone
						status.TotalFiles = parsed.TotalFiles
						status.CurrentFiles = parsed.CurrentFiles
						status.FilesDone = parsed.FilesDone
						status.ErrorCount = parsed.ErrorCount
						status.PercentDone = math.Min(1, math.Max(0, parsed.PercentDone))
						status.SecondsElapsed = parsed.SecondsElapsed
						status.SecondsRemaining = parsed.SecondsRemaining
					})
				}
			} else if bytes.HasPrefix(line, jsonSummaryPrefix) {
				parsed := new(resticJsonBackupSummary)
				if err = json.Unmarshal(line, &parsed); err == nil {
					provider.UpdateSummary(func(summary *monitor.Summary) {
						summary.DryRun = parsed.DryRun
						summary.Extended = true
						summary.FilesNew = parsed.FilesNew
						summary.FilesChanged = parsed.FilesChanged
						summary.FilesUnmodified = parsed.FilesUnmodified
						summary.DirsNew = parsed.DirsNew
						summary.DirsChanged = parsed.DirsChanged
						summary.DirsUnmodified = parsed.DirsUnmodified
						summary.FilesTotal = parsed.TotalFilesProcessed
						summary.BytesAdded = parsed.DataAdded
						summary.BytesStored = stored
						summary.BytesTotal = parsed.TotalBytesProcessed
						summary.Duration = time.Duration(math.Round(float64(time.Second) * parsed.TotalDuration))
						summary.SnapshotID = parsed.SnapshotID
						if added+uncompressed > 0 && stored+compressed > 0 {
							summary.Compression = float64(stored+compressed) / float64(added+uncompressed)
						}
					})
				}
			}
			if err != nil {
				clog.Trace("parsing failed with %q for line:\n%s", err.Error(), string(line))
			}
		}
	}
	return scanner.Err()
}

func HideJson(output io.Writer) io.WriteCloser {
	return LineOutputFilter(output, func(line []byte) bool {
		line = bytes.TrimPrefix(line, bogusPrefix)
		return !bytes.HasPrefix(line, jsonPrefix) || !bytes.HasSuffix(line, jsonSuffix)
	})
}
