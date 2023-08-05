package console

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util/ansi"
	"golang.org/x/exp/slices"
)

const (
	asciiBarStyle       = 0
	defaultBarStyle     = 6
	slimDefaultBarStyle = 4
	minBarLength        = 3
	maxBarLength        = 18
	maxFilesInStatus    = 7
)

type Progress struct {
	profile *config.Profile
	command string
	files   []string
	frame   int
}

func NewProgress(profile *config.Profile) *Progress {
	return &Progress{
		profile: profile,
	}
}

func (p *Progress) Start(command string) {
	p.command = command
	p.frame = 0
}

func (p *Progress) Summary(command string, summary monitor.Summary, stderr string, result error) {
	defer func() { p.command = "" }()

	term.SetStatus(nil)
	term.WaitForStatus()

	if !summary.Extended || p.command != command {
		return
	}

	buffer := &bytes.Buffer{}
	switch command {
	case constants.CommandBackup:
		p.createBackupSummary(buffer, summary, result)
	}

	if buffer.Len() > 0 {
		if result == nil {
			clog.Infof("%s on profile '%s' succeeded:\n\n%s", command, p.profile.Name, buffer.String())
		} else {
			clog.Errorf("%s on profile '%s' failed:\n\n%s", command, p.profile.Name, buffer.String())
		}
	}
}

func (p *Progress) createBackupSummary(buffer *bytes.Buffer, summary monitor.Summary, result error) {
	tw := tabwriter.NewWriter(buffer, 4, 4, 1, ' ', 0)
	defer tw.Flush()

	addOp := "Added"
	if summary.DryRun {
		addOp = "Would add"
	}

	stats := "%s:\t%5d new,\t%5d changed,\t%5d unmodified\n"
	_, _ = fmt.Fprintf(tw, stats, "Files", summary.FilesNew, summary.FilesChanged, summary.FilesUnmodified)
	_, _ = fmt.Fprintf(tw, stats, "Dirs", summary.DirsNew, summary.DirsChanged, summary.DirsUnmodified)
	if summary.BytesStored > 0 {
		_, _ = fmt.Fprintf(tw, "%s to the repository: %s (%s stored)\n", addOp, formatBytes(summary.BytesAdded), formatBytes(summary.BytesStored))
	} else {
		_, _ = fmt.Fprintf(tw, "%s to the repository: %s\n", addOp, formatBytes(summary.BytesAdded))
	}
	_, _ = fmt.Fprintf(tw, "processed %d files, %s in %s\n", summary.FilesTotal, formatBytes(summary.BytesTotal), formatDuration(summary.Duration))
	if summary.SnapshotID != "" {
		_, _ = fmt.Fprintf(tw, "snapshot %s saved\n", summary.SnapshotID)
	}
	if result != nil {
		_, _ = fmt.Fprintf(tw, "error: %s\n", result.Error())
	}
}

func (p *Progress) Status(status monitor.Status) {
	if !term.OutputIsTerminal() {
		return
	}

	width, height := term.OsStdoutTerminalSize()
	withAnsi := term.IsColorableOutput()
	lines := p.createStatus(withAnsi, width, height, &status)
	term.SetStatus(lines)

	p.frame++
	if p.frame >= 16 {
		p.frame = 0
	}
}

func (p *Progress) createStatus(withAnsi bool, width, height int, status *monitor.Status) (lines []string) {
	if withAnsi && height > 4 && canUnicode {
		// multiline version
		height--
		if height > 7 {
			lines = append(lines, p.createStatusHeaderLine(width/3*2, status))
			height--
		}
		lines = append(lines, "")
		lines = append(lines, p.createMultilineFilesStatus(width, height-3, status)...)
		lines = append(lines, "")
		lines = append(lines, p.createStatusLine(withAnsi, false, width, status))
	} else {
		// slim version
		withFiles := width >= 40
		lines = append(lines, p.createStatusLine(withAnsi, withFiles, width, status))
	}
	return
}

func (p *Progress) createFramedLine(width int, content, prefix, suffix, trailer []rune) string {
	line := strings.Builder{}
	contentLen := 0

	if content != nil && prefix != nil && suffix != nil {
		cl, _ := ansi.RunesLength(content, -1)
		pl, _ := ansi.RunesLength(prefix, -1)
		sl, _ := ansi.RunesLength(suffix, -1)
		contentLen = cl + pl + sl

		line.WriteString(ansi.Gray(string(prefix)))
		line.WriteString(string(content))
		line.WriteString(ansi.Gray(string(suffix)))
	}

	// trailer
	{
		gray, notGray := ansi.ColorSequence(ansi.Gray)
		line.WriteString(gray)
		tr := string(trailer)
		tl, _ := ansi.RunesLength(trailer, -1)
		for remaining := width - contentLen; remaining >= tl; remaining -= tl {
			line.WriteString(tr)
		}
		line.WriteString(notGray)
	}

	return line.String()
}

var (
	utfHeaderPrefix = []rune("╌╌⟪  ")
	utfHeaderSuffix = []rune("  ⟫╌╌")
	utfHeaderTail   = []rune("  ")
	utfDelimiter    = []rune(" ᜶ ")
	asciiDelimiter  = " | "
)

func (p *Progress) createStatusHeaderLine(width int, status *monitor.Status) string {
	if status.TotalFiles > 0 {
		d := string(utfDelimiter)
		header := fmt.Sprintf("%s %s %s %s %s",
			ansi.Bold(status.TotalFiles),
			ansi.Gray("files total "+d),
			ansi.Bold(formatBytes(status.TotalBytes)),
			ansi.Gray(" "+d+" elapsed "),
			ansi.Bold(formatDuration(time.Second*time.Duration(status.SecondsElapsed))),
		)
		return p.createFramedLine(width, []rune(header), utfHeaderPrefix, utfHeaderSuffix, utfHeaderTail)
	}
	return p.createFramedLine(width, nil, nil, nil, utfHeaderTail)
}

var spinnerFrames = []rune("◜◜◝◝◞◞◟◟")

func (p *Progress) createStatusLine(withAnsi, files bool, width int, status *monitor.Status) string {
	line := strings.Builder{}

	remaining := width
	filesReserved := width / 3
	if files && width > 30 {
		remaining -= filesReserved
	}

	// select delimiter
	ansiDelimiter, delimiterLen := "", 0
	{
		d := asciiDelimiter
		if canUnicode {
			d = string(utfDelimiter)
			if remaining > 40 {
				d = " " + d + " "
			}
		}
		delimiterLen, _ = ansi.RunesLength([]rune(d), -1)
		ansiDelimiter = ansi.Gray(d)
	}

	// progress
	barWidth := int(math.Max(minBarLength, math.Min(maxBarLength, float64(remaining)/3)))
	bar := formatProgress(barStyle, barWidth, withAnsi, status.PercentDone)
	if canUnicode && width > 40 {
		var spinner string
		if status.PercentDone < 0.999 {
			spinner = string(spinnerFrames[p.frame%len(spinnerFrames)])
		} else {
			spinner = ansi.Green("✔")
		}
		bar = ansi.Gray(string(utfHeaderPrefix)) + spinner + " " + bar
	}
	remaining -= len(bar) + delimiterLen
	line.WriteString(bar)

	// volume & throughput
	if remaining > 10 {
		line.WriteString(ansiDelimiter)
		processed := fmt.Sprintf("%s, %d errors", formatBytes(status.BytesDone), status.ErrorCount)
		remaining -= len(processed) + delimiterLen
		line.WriteString(processed)
	}
	if remaining > 10 && status.SecondsElapsed > 0 {
		line.WriteString(ansiDelimiter)
		speed := fmt.Sprintf("%s/s", formatBytes(status.BytesDone/uint64(status.SecondsElapsed)))
		remaining -= len(speed) + delimiterLen
		line.WriteString(speed)
	}

	// time
	if remaining > 10 && status.SecondsElapsed > 5 && status.SecondsRemaining < 7*24*60*60 {
		line.WriteString(ansiDelimiter)
		timeRemaining := "ETA " + formatDuration(time.Second*time.Duration(status.SecondsRemaining))
		remaining -= len(timeRemaining) + delimiterLen
		line.WriteString(timeRemaining)
	}

	// current files (if enabled for status line)
	remaining += filesReserved
	if remaining > 16 && files {
		line.WriteString(ansiDelimiter)

		fileCount := len(status.CurrentFiles)
		_, _ = fmt.Fprintf(&line, "[%d] ", fileCount)
		remaining -= 4

		if fileCount > 0 {
			fileLen := remaining / fileCount
			if fileCount > 1 {
				fileLen -= delimiterLen
			}
			if fileLen < 20-delimiterLen {
				fileLen = 20 - delimiterLen
			}
			for i, file := range status.CurrentFiles {
				if len(file) > fileLen {
					file = file[len(file)-fileLen:]
				}
				if i > 0 && remaining > 0 {
					remaining -= delimiterLen
					line.WriteString(ansiDelimiter)
				}
				if remaining <= 0 {
					break
				}
				line.WriteString(file)
			}
		}
	}

	line.WriteString(ansi.Gray("  ⟫"))

	return line.String()
}

func (p *Progress) createMultilineFilesStatus(width, height int, status *monitor.Status) (lines []string) {
	filesCount := height
	if filesCount > maxFilesInStatus {
		filesCount = maxFilesInStatus
	}

	// sort and truncate
	p.files = append(p.files, status.CurrentFiles...)
	sort.Strings(p.files)
	p.files = slices.Compact(p.files)
	if len(p.files) > filesCount {
		p.files = p.files[len(p.files)-filesCount:]
	}

	// format output
	maxFileLen := 0
	for _, file := range p.files {
		if file = filepath.Base(file); len(file) > maxFileLen {
			maxFileLen = len(file)
		}
	}
	lineFormat := fmt.Sprintf("  %%s %%-%ds  %%s  ", maxFileLen)

	previousDir := "|initial|"
	lines = make([]string, 0, len(p.files))
	current := make([]string, 0, len(status.CurrentFiles))
	for _, file := range p.files {
		inCurrent := slices.Contains(status.CurrentFiles, file)

		dir := filepath.Dir(file)
		file = filepath.Base(file)
		if dir == previousDir {
			dir = ""
		} else {
			previousDir = dir
		}

		if pathLen := 11 + maxFileLen + len(dir); pathLen > width {
			dir = dir[len(dir)-(pathLen-width):]
			if width >= 30 {
				dir = "..." + dir[3:]
			}
		}
		if dir != "" {
			dir = fmt.Sprintf("[ %s ]", dir)
		}

		if inCurrent {
			current = append(current, fmt.Sprintf(lineFormat, ansi.Bold(ansi.Cyan(">")), file, ansi.Gray(dir)))
		} else {
			lines = append(lines, ansi.Gray(fmt.Sprintf(lineFormat, " ", file, dir)))
		}
	}
	lines = append(lines, current...)
	return
}

var byteUnits = []string{"Bytes", "KiB", "MiB", "GiB", "TiB"}

func formatBytes(bytes uint64) string {
	value := float64(bytes)
	unit := 0
	for ; unit < len(byteUnits)-1; unit++ {
		if value < 1024 {
			break
		} else {
			value /= 1024
		}
	}
	format := "%4.0f %s"
	if unit > 0 {
		digits := int(math.Max(0, math.Ceil(3-math.Log10(value))-1))
		if digits > 0 {
			format = fmt.Sprintf("%%.%df %%s", digits)
		}
	}
	return fmt.Sprintf(format, value, byteUnits[unit])
}

func formatDuration(duration time.Duration) string {
	const TimeOnly = "15:04:05" // use time.TimeOnly when go 1.20 is min ver
	tf := time.UnixMilli(duration.Milliseconds()).UTC().Format(TimeOnly)
	if days := math.Floor(duration.Hours() / 24); days >= 1 {
		return fmt.Sprintf("%dd %s", int(days), tf)
	}
	return strings.TrimPrefix(tf, "00:")
}

var barStyles = [][]rune{
	[]rune(" -+*#"), // ASCII only
	[]rune(" ░▒▓█"),
	[]rune(" ▏▎▍▌▋▊▉█"),
	[]rune(" ▁▂▃▄▅▆▇█"),
	[]rune(" ⡀⡄⡆⡇⡏⡟⡿⣿"), // slim-default
	[]rune(" ⡀⣀⣠⣤⣦⣶⣾⣿"),
	[]rune(" ⠄⠆⠇⠏⠟⠿"), //   default
	[]rune(" ⠄⠤⠴⠶⠷⠿"),
}

var barStyle = func() int {
	idx, err := strconv.Atoi(os.Getenv("RESTICPROFILE_PROGRESS_BAR_STYLE"))
	if err != nil {
		idx = asciiBarStyle
		if term.IsColorableOutput() {
			idx = defaultBarStyle
		}
	}
	return idx
}()

var canUnicode = barStyle > 0

func formatProgress(style, width int, withAnsi bool, progress float64) string {
	out := strings.Builder{}
	remaining := width

	if style >= 0 && style < len(barStyles) {
		if width > 11 {
			remaining = 6
			width -= 6
		} else {
			if style == defaultBarStyle {
				style = slimDefaultBarStyle
			}
			remaining = 0
		}

		chars := barStyles[style]
		ticks := len(chars) * width
		lit := int(math.Round(float64(ticks) * progress))

		for i := 0; i < width; i++ {
			idx := 0
			if lit >= len(chars) {
				idx = len(chars) - 1
				lit -= len(chars)
			} else if lit > 0 {
				idx = lit - 1
				lit = 0
			}
			out.WriteRune(chars[idx])
		}

		if withAnsi {
			greenBar := ansi.Green(out.String())
			out.Reset()
			out.WriteString(greenBar)
		}
	}

	if remaining >= 4 {
		if out.Len() > 0 && remaining > 4 {
			out.WriteString(" ")
			remaining--
		}
		format := "%4.1f%%"
		if remaining < 5 {
			format = "%3.0f%%"
		} else if progress >= .999 {
			format = "%4.0f%%"
		}
		_, _ = fmt.Fprintf(&out, format, progress*100)
	}

	return out.String()
}

// Verify interface
var _ monitor.Receiver = &Progress{}
