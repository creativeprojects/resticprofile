package shell

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/monitor"
)

// scanBackupPlain should populate the backup summary values from the standard output
var scanBackupPlain ScanOutput = func(r io.Reader, provider monitor.Provider, w io.Writer) error {
	scanner := bufio.NewScanner(io.TeeReader(r, w))
	for scanner.Scan() {
		line := scanner.Text()

		// scan content - it's all right if the line does not match
		provider.UpdateSummary(func(summary *monitor.Summary) {
			_, _ = fmt.Sscanf(line, "Files: %d new, %d changed, %d unmodified", &summary.FilesNew, &summary.FilesChanged, &summary.FilesUnmodified)
			_, _ = fmt.Sscanf(line, "Dirs: %d new, %d changed, %d unmodified", &summary.DirsNew, &summary.DirsChanged, &summary.DirsUnmodified)
			_, _ = fmt.Sscanf(line, "snapshot %s saved", &summary.SnapshotID)

			// restic < 14
			bytes, unit := 0.0, ""
			n, _ := fmt.Sscanf(line, "Added to the repo: %f %3s", &bytes, &unit)
			if n == 2 {
				summary.BytesAdded = unformatBytes(bytes, unit)
			}

			// restic >=14
			storedBytes, storedUnit, addVerb := 0.0, "", ""
			n, _ = fmt.Sscanf(line, "%s to the repository: %f %3s (%f %3s stored)", &addVerb, &bytes, &unit, &storedBytes, &storedUnit)
			if n >= 3 && addVerb == "Added" || addVerb == "Would add" {
				summary.DryRun = strings.HasPrefix(addVerb, "Would")
				summary.BytesAdded = unformatBytes(bytes, unit)
				if n == 5 {
					summary.BytesStored = unformatBytes(storedBytes, storedUnit)
				}
			}

			duration := ""
			n, _ = fmt.Sscanf(line, "processed %d files, %f %3s in %s", &summary.FilesTotal, &bytes, &unit, &duration)
			if n == 4 {
				summary.BytesTotal = unformatBytes(bytes, unit)

				duration = strings.TrimSpace(duration)
				if len(duration) == 4 || len(duration) == 6 {
					duration = "0" + duration
				}
				if len(duration) == 5 {
					duration = "00:" + duration
				}
				const TimeOnly = "15:04:05" // use time.TimeOnly when go 1.20 is min ver
				if p, err := time.Parse(TimeOnly, duration); err == nil {
					summary.Duration = p.Sub(
						time.Date(p.Year(), p.Month(), p.Day(), 0, 0, 0, 0, p.Location()))
				}
			}
		})
	}
	return scanner.Err()
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
