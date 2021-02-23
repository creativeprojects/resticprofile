package shell

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"runtime"
	"strings"
	"time"
)

// Summary of the profile run
type Summary struct {
	Duration        time.Duration
	FilesNew        int
	FilesChanged    int
	FilesUnmodified int
	DirsNew         int
	DirsChanged     int
	DirsUnmodified  int
	FilesTotal      int
	BytesAdded      uint64
	BytesTotal      uint64
}

// ScanBackup should populate the backup summary values from the standard output
func ScanBackup(r io.Reader, summary *Summary, w io.Writer) error {
	eol := "\n"
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	}
	rawBytes, unit, duration := 0.0, "", ""
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		w.Write([]byte(scanner.Text() + eol))
		// scan content - it's all right if the line does not match
		_, _ = fmt.Sscanf(scanner.Text(), "Files: %d new, %d changed, %d unmodified", &summary.FilesNew, &summary.FilesChanged, &summary.FilesUnmodified)
		_, _ = fmt.Sscanf(scanner.Text(), "Dirs: %d new, %d changed, %d unmodified", &summary.DirsNew, &summary.DirsChanged, &summary.DirsUnmodified)

		n, err := fmt.Sscanf(scanner.Text(), "Added to the repo: %f %3s", &rawBytes, &unit)
		if n == 2 && err == nil {
			summary.BytesAdded = unformatBytes(rawBytes, unit)
		}

		n, err = fmt.Sscanf(scanner.Text(), "processed %d files, %f %3s in %s", &summary.FilesTotal, &rawBytes, &unit, &duration)
		if n == 4 && err == nil {
			summary.BytesTotal = unformatBytes(rawBytes, unit)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func unformatBytes(value float64, unit string) uint64 {
	switch strings.TrimSpace(unit) {
	case "KiB":
		value *= 1024
	case "MiB":
		value *= 1024 * 1024
	case "GiB":
		value *= 1024 * 1024 * 1024
	case "TiB":
		value *= 1024 * 1024 * 1024 * 1024
	}
	return uint64(math.Round(value))
}
